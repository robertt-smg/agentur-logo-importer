// see also https://github.com/josephspurrier/goversioninfo
//
//go:generate bash ./get_version.sh
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"time"

	"agentur-logo-importer/brevo"
	"agentur-logo-importer/config"
	"path/filepath"
	"runtime"
	"strings"

	"html/template"

	//"github.com/golang/glog"
	glog "github.com/kpango/glg"
	"github.com/pkg/sftp"

	// "gitlab.com/fti-go/pkg/ntlm.git"

	"golang.org/x/crypto/ssh"
	"gopkg.in/ini.v1"
)

const (
	mime = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
)

var (
	log         *os.File
	name        = "agentur-logo-importer"
	description = "Agentur Logo Upload"

	//version     = "undefined"
	//build = "undefined"

	isSuccessOrLogError = func(err error, detail string) error {
		if err != nil {
			glog.Error("Error: ", detail, err)
		}
		return err
	}
)

func getIniCfg(path *string) *ini.File {
	if path == nil {
		s, _ := getIniPath()
		path = &s
	}
	cfg, err := ini.Load(*path)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	return cfg
}

type this struct {
	cfg      *ini.File
	hostname string
	port     string
	smtpUser string
	smtpPW   string
	from     string
}
type logo struct {
	Serverpath string
	Localpath  string
	Destpath   string
	Country    string
	Agency     string
	Subagents  []string
	Downloaded bool
	Uploaded   bool
	Kimed      bool
	FormatOk   bool
	Err        error
}

func new(cfg *ini.File) *this {
	t := &this{}
	t.cfg = cfg
	t.hostname = t.cfg.Section("smtp_server").Key("hostname").String()
	t.port = t.cfg.Section("smtp_server").Key("port").String()
	t.smtpUser = t.cfg.Section("smtp_server").Key("smtpUser").String()
	t.smtpPW = t.cfg.Section("smtp_server").Key("smtpPW").String()
	t.from = t.cfg.Section("email").Key("from").String()
	return t
}

func (t *this) setLog() {
	flag.Parse()
	flag.Lookup("alsologtostderr").Value.Set("true")
	//flag.Lookup("log_dir").Value.Set("q:/airquest/system/fti")
	logPath := t.cfg.Section("").Key("logPath").String()
	//flag.Lookup("log_dir").Value.Set(logPath)
	log = glog.FileWriter(logPath+"/"+name+time.Now().Format("_20060102150405.log"), 0666)

	glog.Get().
		SetMode(glog.BOTH). // default is STD
		AddLevelWriter(glog.INFO, log).
		AddLevelWriter(glog.ERR, log)

}

func publicKeyFile(file string) ssh.AuthMethod {
	key, err := os.ReadFile(file)
	if err != nil {
		glog.Fatal("unable to read private key: %v", err)
		return nil
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		glog.Fatal("unable to parse private key: %v", err)
		return nil
	}
	return ssh.PublicKeys(signer)
}

func (t *this) getSSHConnection() *ssh.Client {
	glog.Info(logGetCurrentFuncName())

	config := &ssh.ClientConfig{
		User: t.cfg.Section("webserver").Key("user").String(),
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			publicKeyFile(t.cfg.Section("webserver").Key("PrivatKeyFilePath").String()),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	server := t.cfg.Section("webserver").Key("addr").String()

	conn, err := ssh.Dial("tcp", server, config)
	if err != nil {
		t.sendMailToAdmin([]byte("Es konnte keine SSH Verbindung aufgebaut werden. Bitte die logs ansehen."))
		glog.Fatalf("Function: " + logGetCurrentFuncName() + ". Failed to dial: " + err.Error())
	}

	return conn
}

func (t *this) getSftpClient(conn *ssh.Client) *sftp.Client {
	client, err := sftp.NewClient(conn)
	if err != nil {
		t.sendMailToAdmin([]byte("Es konnte keine SFTP Session erzeugt werden. Bitte die logs ansehen."))
		glog.Fatalf("Failed to create client: " + err.Error())
	}
	return client
}

func (t *this) getFilesFromServer(client *sftp.Client) (logos map[string]*logo) {
	glog.Info(logGetCurrentFuncName())

	var paths []string
	format := t.cfg.Section("file").Key("format").String()
	validExt := "." + strings.ToLower(format)

	logoPath := t.cfg.Section("webserver").Key("logoPath").String()

	w := client.Walk(logoPath)
	for w.Step() {
		if w.Err() != nil {
			continue
		}
		wPath := w.Path()
		fileExt := strings.ToLower(filepath.Ext(wPath))
		if fileExt == validExt {
			paths = append(paths, w.Path())
		}
	}
	//create logos  map
	logos = make(map[string]*logo)

	for _, path := range paths {
		if !strings.Contains(path, "ignore") {

			pathDir, filename := sftp.Split(path)

			parts := strings.Split(pathDir, "/")
			agency := "-"
			if len(parts) > 2 {

				if len(parts) == 4 {
					// replace filename with dir, which should be equal to fe_user/agency name
					agency = parts[2]
				} else {
					fparts := strings.Split(filename, ".")
					if len(fparts) > 1 {
						agency = fparts[0]
					}
				}

				l := logo{}
				l.Country = parts[1]
				l.Serverpath = path
				l.Agency = agency
				logos[agency] = &l
			}
		}
	}
	return logos
}

func (t *this) convertToJPG(logos map[string]*logo) {
	for _, logo := range logos {
		if !logo.Downloaded {
			continue
		}
		sourcepath := logo.Localpath
		destinationPath := logo.Destpath
		formatOk, err := ConvertImageToJPEG(sourcepath, destinationPath)
		switch err {
		case nil:
			logo.FormatOk = formatOk

		case FAILED_TO_OPEN:
			logo.Err = err
			msg := fmt.Sprintf("Error failed to open image :%v  %v", err, logo)
			glog.Errorf(msg)

		case FAILED_TO_CONVERT:
			logo.Err = err
			msg := fmt.Sprintf("Error to convert image :%v  %v", err, logo)
			glog.Errorf(msg)
		}

	}
}
func (t *this) copyServerFilesToLokal(logos map[string]*logo, client *sftp.Client) {
	glog.Info(logGetCurrentFuncName())

	lokalPath := t.cfg.Section("paths").Key("lokalPath").String()

	for _, logo := range logos {

		// Open the source file
		srcFile, err := client.Open(logo.Serverpath)
		if err != nil {
			logo.Err = err
			msg := fmt.Sprintf("%s Error Open Remote: %s %v %v", logGetCurrentFuncName(), logo.Serverpath, err, logo)
			t.sendMailToAdmin([]byte(msg))
			glog.Fatalf(msg)
		}
		defer srcFile.Close()
		_, err = os.Stat(lokalPath)
		if err != nil && os.IsNotExist(err) {
			os.MkdirAll(lokalPath, os.ModePerm)
		}
		// Create the destination file
		filename := filepath.Join(lokalPath, logo.Agency+".jpg")
		dstFile, err := os.Create(filename)
		if err != nil {
			logo.Err = err
			msg := fmt.Sprintf("%s Error Create Local: %s %v %v", logGetCurrentFuncName(), filename, err, logo)
			t.sendMailToAdmin([]byte(msg))
			glog.Fatalf(msg)
		}
		defer dstFile.Close()

		// Copy the file
		_, err = srcFile.WriteTo(dstFile)
		if err != nil {
			logo.Err = err
			msg := fmt.Sprintf("%s Error Download: %s %v %v", logGetCurrentFuncName(), filename, err, logo)
			t.sendMailToAdmin([]byte(msg))
			glog.Fatalf(msg)
		} else {
			glog.Info(logGetCurrentFuncName(), logo.Agency, logo.Country, logo.Serverpath, logo.Localpath)
			logo.Localpath = filename
			logo.Destpath = filename
			logo.Downloaded = true
		}
	}
}
func (t *this) removeFilesOnWebserver(logos map[string]*logo, client *sftp.Client) {
	glog.Info(logGetCurrentFuncName())

	for _, logo := range logos {

		if logo.Downloaded && logo.Uploaded && logo.Kimed {
			glog.Info(logGetCurrentFuncName(), logo.Agency, logo.Country, logo.Serverpath)

			err := client.Remove(logo.Serverpath)
			if err != nil {
				msg := fmt.Sprintf("%s Error Delete Remote: %v %v", logGetCurrentFuncName(), err, logo)
				t.sendMailToAdmin([]byte(msg))
				glog.Fatalln(msg)
			} else {
				logo.Err = err
				glog.Info(logGetCurrentFuncName(), logo.Serverpath, "deleted")
			}

		}
	}
}

type mailData struct {
	Logos map[string]*logo
}

func (t *this) notifyHumans(logos map[string]*logo) {
	glog.Info(logGetCurrentFuncName())
	mailTmpl := t.cfg.Section("template").Key("mail_body").String()

	tmpl, err := template.New("mailTmpl").Parse(mailTmpl)
	if err != nil {
		glog.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, mailData{Logos: logos})
	if err == nil {
		msg := buf.String()

		//auth := NtlmAuth("", t.smtpUser, t.smtpPW, t.hostname)

		emails := strings.Split(t.cfg.Section("email_addresses").Key("all").String(), ",")
		recipients := []string{}
		for _, email := range emails {
			recipients = append(recipients, email)
		}
		mail := make(map[string]string)
		mail["smtpHost"] = "brevo.api"

		mail["smtpUser"] = "*"
		mail["smtpPassword"] = "*"
		mail["from"] = os.Getenv("Alert_From")
		mail["to"] = strings.Join(emails, ",")
		mail["subject"] = description
		mail["content"] = msg
		brevo.SendApiMail(mail)

	} else {
		glog.Error(err)
	}
}

func (t *this) sendMailToAdmin(errMsg []byte) {
	errMail := []byte("Subject: AgenturDownloadImage ERROR :( \r\n\r\n")
	errMail = append(errMail, errMsg...)

	adminEmails := t.cfg.Section("admin").Key("emails").String()

	mail := make(map[string]string)
	mail["smtpHost"] = "brevo.api"

	mail["smtpUser"] = "*"
	mail["smtpPassword"] = "*"
	mail["from"] = os.Getenv("Alert_From")
	mail["to"] = adminEmails
	mail["subject"] = description
	mail["content"] = string(errMsg)
	brevo.SendApiMail(mail)
}

func logGetCurrentFuncName() string {
	//stack = debug.Stack
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func main() {
	envPath := flag.String("env", ".env", "Path to environment file")
	iniPath := flag.String("ini", "app.ini", "Path to ini file")

	// Define custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	// Parse command line arguments
	flag.Parse()

	// Show usage if no arguments provided
	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration
	err := config.LoadConfig(*envPath)
	if err != nil {
		glog.Error("Failed to load configuration:", err)
		os.Exit(1)
	}
	cfg := getIniCfg(iniPath)

	t := new(cfg)

	t.setLog()

	glog.Info("Start ...")

	incommingLogo := t.cfg.Section("paths").Key("incommingLogo").String()

	var logos map[string]*logo
	if strings.HasPrefix(incommingLogo, "sftp://") {
		conn := t.getSSHConnection()
		client := t.getSftpClient(conn)
		defer client.Close()
		logos = t.getFilesFromServer(client)
		if len(logos) > 0 {
			t.copyServerFilesToLokal(logos, client)
		}
	} else {
		logos = t.getFilesFromLocal(incommingLogo)
		if len(logos) > 0 {
			t.copyLocalFilesToLokal(logos, incommingLogo)
		}
	}

	if len(logos) > 0 {
		t.convertToJPG(logos)
		if t.cfg.Section("redbox").Key("upload").MustBool() {
			t.uploadLogos(logos)
		}
		t.updateDB(logos)
		if t.cfg.Section("webserver").Key("delete").MustBool() && strings.HasPrefix(incommingLogo, "sftp://") {
			conn := t.getSSHConnection()
			client := t.getSftpClient(conn)
			defer client.Close()
			t.removeFilesOnWebserver(logos, client)
		}
		t.notifyHumans(logos)
	}

	glog.Info("Done ...")
	log.Sync()
	defer log.Close()
}
