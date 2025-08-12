package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	glog "github.com/kpango/glg"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func GetSSHConnection() *ssh.Client {
	config := &ssh.ClientConfig{
		User: "vhost1_sftp02",
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			publicKeyFile("./ssh_prq41_typo3_prd"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", "10.201.1.76:22", config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}

	return conn
}

func GetSftpClient(conn *ssh.Client) *sftp.Client {
	client, err := sftp.NewClient(conn)
	if err != nil {
		panic("Failed to create client: " + err.Error())
	}
	return client
}

func TestGetFiles(t *testing.T) {
	conn := GetSSHConnection()
	client := GetSftpClient(conn)

	var paths []string
	format := "txt"
	logoPath := "AgenturLogoUploads"

	w := client.Walk(logoPath)
	for w.Step() {
		if w.Err() != nil {
			continue
		}
		wPath := w.Path()
		fileFormat := filepath.Ext(wPath)
		if fileFormat == "."+format {
			paths = append(paths, w.Path())
		}
	}

	//fmt.Println(paths)

	dirNames, err := client.ReadDir(logoPath)
	if err != nil {
		glog.Fatal(err, "No dir was found")
	}

	//create files array bzw. map
	files := make(map[string]map[string]string)
	for _, dir := range dirNames {
		for _, path := range paths {
			pathDir, filename := sftp.Split(path)
			dirName := dir.Name()
			if strings.Contains(pathDir, "/"+dirName+"/") {
				files[dirName] = map[string]string{
					filename: path,
				}
			}
		}
	}

	fmt.Println(files)
}

func TestCopyServerfilesToLokal(t *testing.T) {
	conn := GetSSHConnection()
	client := GetSftpClient(conn)

	files := make(map[string]map[string]string)
	files["de"] = map[string]string{"test2.txt": "AgenturLogoUploads/de/test2.txt"}

	lokalPath := ".\\logos\\"

	for _, file := range files {
		for filename, path := range file {
			fmt.Println(path)
			fmt.Println(filename)

			// Open the source file
			srcFile, err := client.Open(path)
			if err != nil {
				glog.Fatal(err)
			}
			defer srcFile.Close()

			// Create the destination file
			dstFile, err := os.Create(lokalPath + filename)
			if err != nil {
				glog.Fatal(err)
			}
			defer dstFile.Close()

			// Copy the file
			srcFile.WriteTo(dstFile)
		}

	}
}
