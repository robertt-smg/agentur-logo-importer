package main

import (
	"bytes"
	"crypto/tls"
	_ "crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
)

type errorList struct {
	Code           int      `json:"code"`
	Description    string   `json:"description"`
	DetailMessages []string `json:"detailMessages"`
}
type clientSession struct {
	LoginName  string `json:"loginName"`
	Password   string `json:"password"`
	ServerName string `json:"serverName"`
}
type authenticationToken struct {
	WebServiceLogID string `json:"loginName"`
	Service         string `json:"service"`
	AgentEmail      string `json:"agentEmail"`
	AgencyName      string `json:"agencyName"`
	AgencyID        string `json:"agencyID"`
	SessionID       string `json:"sessionID"`
	ErrorList       []errorList
}

// TimeoutDialer  Net Request with Timeout
func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}
func (t *this) getWebClient(section string) (*http.Client, error) {
	var tr *http.Transport
	// Service specific
	timeout := t.cfg.Section(section).Key("timeout").MustDuration(time.Minute)

	bUseProxy := t.cfg.Section(section).Key("use_proxy").MustBool(false)
	if bUseProxy {
		// Proxy in general
		bUseProxy = t.cfg.Section("proxy").Key("use_proxy").MustBool(false)
	}
	if bUseProxy {
		// URL Config
		proxyString := t.cfg.Section("proxy").Key("proxy_url").String()
		proxyURL, _ := url.Parse(proxyString)
		tr = &http.Transport{
			Dial: TimeoutDialer(timeout, timeout),
			// DisableCompression: true,
			TLSClientConfig: &tls.Config{

				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS12,
				/*CipherSuites: []uint16{
					tls.TLS_RSA_WITH_RC4_128_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				},*/
			},
			Proxy: http.ProxyURL(proxyURL),
		}
	} else {
		tr = &http.Transport{
			Dial: TimeoutDialer(timeout, timeout),
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}
	}

	var netClient = &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}
	return netClient, nil
}
func (t *this) uploadToRedbox(netClient *http.Client, url string, clientSession clientSession, logo *logo) {

	glog.V(2).Infoln("-> ", logGetCurrentFuncName(), clientSession.LoginName)

	b, err := json.Marshal(clientSession)
	//reqString := string(b)
	//glog.V(2).Infoln("-> ", logGetCurrentFuncName(), reqString)
	body := ioutil.NopCloser(bytes.NewBuffer([]byte(b)))

	// Build the request
	req, err := http.NewRequest("POST", url, body)
	if isSuccessOrLogError(err, "Cannot create web request") == nil {
		req.Header.Add("Content-Type", "application/json; charset=utf-8")
		req.Header.Add("Content-Encoding", "application/json")
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Length", fmt.Sprintf("%d", len(b)))

		resp, err := netClient.Do(req)
		if isSuccessOrLogError(err, "Cannot send web request") == nil {

			defer resp.Body.Close()
			var auth authenticationToken
			// Use json.Decode for reading streams of JSON data
			err := json.NewDecoder(resp.Body).Decode(&auth)
			if isSuccessOrLogError(err, "Cannot authenticate agency:"+logo.Agency) == nil {
				if len(auth.ErrorList) > 0 {
					err = fmt.Errorf("%d %s", auth.ErrorList[0].Code, auth.ErrorList[0].Description)
					glog.Errorln("Error: ", auth.ErrorList[0].Code, auth.ErrorList[0].Description)
					logo.Err = err
				} else {
					glog.V(2).Infoln(logGetCurrentFuncName(), auth.SessionID)
					err = t.uploadFile(logo, auth.SessionID, netClient)
					if err == nil {
						logo.Uploaded = true
					}
					logo.Err = err
				}
			}
		}
	}
}
func (t *this) uploadLogos(logos map[string]*logo) {

	glog.V(2).Infoln("-> ", logGetCurrentFuncName())

	restURL := t.cfg.Section("redbox").Key("restUrl").String()
	restApiClientSession := t.cfg.Section("redbox").Key("restApiClientSession").String()
	serverName := t.cfg.Section("redbox").Key("serverName").String()
	password := t.cfg.Section("redbox").Key("password").String()

	netClient, _ := t.getWebClient("redbox")

	for _, logo := range logos {

		if logo.Downloaded {
			url := fmt.Sprintf("%s%s", restURL, restApiClientSession)

			clientSession := clientSession{}
			clientSession.LoginName = logo.Agency
			clientSession.Password = password
			clientSession.ServerName = serverName

			t.uploadToRedbox(netClient, url, clientSession, logo)

			logo.Subagents = t.loadSubagents(logo)
			for _, subagent := range logo.Subagents {
				clientSession.LoginName = subagent

				t.uploadToRedbox(netClient, url, clientSession, logo)
			}
		}
	}
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func (t *this) uploadFile(logo *logo, sessionID string, netClient *http.Client) (ret error) {

	glog.V(2).Infoln("-> ", logGetCurrentFuncName(), logo.Agency)

	restURL := t.cfg.Section("redbox").Key("restUrl").String()
	restUploadFile := t.cfg.Section("redbox").Key("restUploadFile").String()

	url := fmt.Sprintf("%s%s/%s", restURL, sessionID, restUploadFile)

	f, err := t.mustOpen(logo.Localpath) // local logo file
	if isSuccessOrLogError(err, "Cannot open file:"+logo.Localpath) == nil {

		//prepare the reader instances to encode
		values := map[string]io.Reader{
			"file": f,
			//"sessionId": strings.NewReader(sessionID),
		}

		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		for key, r := range values {
			var fw io.Writer
			if x, ok := r.(io.Closer); ok {
				defer x.Close()
			}
			// Add an image file
			if _, ok := r.(*os.File); ok {
				filename := logo.Agency + ".jpg"
				h := make(textproto.MIMEHeader)
				h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
					escapeQuotes(key), escapeQuotes(filename)))
				h.Set("Content-Type", "image/jpeg")

				if fw, err = w.CreatePart(h); err != nil {

					break
				}
			} else {
				// Add other fields
				if fw, err = w.CreateFormField(key); err != nil {
					break
				}
			}
			if _, err = io.Copy(fw, r); err != nil {
				break
			}

		}
		// Don't forget to close the multipart writer.
		// If you don't close it, your request will be missing the terminating boundary.
		w.Close()
		if isSuccessOrLogError(err, "Failed to build form-multipart") == nil {
			// Now that you have a form, you can submit it to your handler.
			req, err := http.NewRequest("POST", url, &b)
			if isSuccessOrLogError(err, "Failed to creare http request") == nil {
				// Don't forget to set the content type, this will contain the boundary.
				req.Header.Set("Content-Type", w.FormDataContentType())

				// Submit the request
				res, err := netClient.Do(req)
				if isSuccessOrLogError(err, "POST file failed") == nil {

					// Check the response
					if res.StatusCode != http.StatusOK {
						ret = fmt.Errorf("bad status: %s", res.Status)
						glog.Errorln("Error: ", ret)
					}
				} else {
					ret = err
				}
			}
		} else {
			ret = err
		}
	} else {
		ret = err
	}
	return ret
}

func (t *this) mustOpen(f string) (*os.File, error) {
	r, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	return r, nil
}
