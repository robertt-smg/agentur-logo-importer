// see also https://github.com/josephspurrier/goversioninfo
//go:generate bash ./get_version.sh
package main

import (
	"crypto/tls"
	"net/smtp"

	//"github.com/golang/glog"
	glog "github.com/kpango/glg"
	// "gitlab.com/fti-go/pkg/ntlm.git"
)

func SendMailIgnoreTLS(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	glog.Info(logGetCurrentFuncName(), addr, from)

	c, err := smtp.Dial(addr)

	if err != nil {
		return err
	}
	//if err := c.hello(); err != nil {
	//    return err
	//}
	if ok, _ := c.Extension("STARTTLS"); ok {
		/*if err = c.StartTLS(nil); err != nil*/ {
			//host, _, _ := net.SplitHostPort(addr)
			config := &tls.Config{
				InsecureSkipVerify: true,
				//ServerName:         "",
			}

			if err = c.StartTLS(config); err != nil {
				glog.Error("starttls with config ->", err)
				return err
			}
		}

	}
	if a != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(a); err != nil {
				return err
			}
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
