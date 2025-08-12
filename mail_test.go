package main

import (
	"fmt"
	"net/smtp"

	glog "github.com/kpango/glg"

	//"gitlab.com/fti-go/pkg/ntlm.git"

	"testing"
)

func TestExample(t *testing.T) {
	// Connect to the remote SMTP server.
	c, err := smtp.Dial("mail.example.com:25")
	if err != nil {
		glog.Fatal(err)
	}

	// Set the sender and recipient first
	if err := c.Mail("sender@example.org"); err != nil {
		glog.Fatal(err)
	}
	if err := c.Rcpt("recipient@example.net"); err != nil {
		glog.Fatal(err)
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		glog.Fatal(err)
	}
	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		glog.Fatal(err)
	}
	err = wc.Close()
	if err != nil {
		glog.Fatal(err)
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		glog.Fatal(err)
	}
}

// variables to make ExamplePlainAuth compile, without adding
// unnecessary noise there.
var (
	from = "ticketshop.devops@fti.de"
	msg  = []byte("Subject: discount Gophers!\r\n" +
		"\r\n" +
		"This is the email body.\r\n")
	recipients = []string{"robert.thurnreiter@fti.de"}
)

func TestExamplePlainAuth(t *testing.T) {
	// hostname is used by PlainAuth to validate the TLS certificate.
	hostname := "localhost"
	auth := smtp.PlainAuth("", "", "", hostname)

	//auth := NtlmAuth("", "muc01_airquest", "a1rqu3st", hostname)

	err := SendMailIgnoreTLS(hostname+":1025", auth, from, recipients, msg)
	if err != nil {
		glog.Fatal(err)
	}
}

func TestExampleSendMail(t *testing.T) {
	// Set up authentication information.
	auth := smtp.PlainAuth("", "user@example.com", "password", "mail.example.com")

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{"recipient@example.net"}
	msg := []byte("To: recipient@example.net\r\n" +
		"Subject: discount Gophers!\r\n" +
		"\r\n" +
		"This is the email body.\r\n")
	err := SendMailIgnoreTLS("mail.example.com:25", auth, "sender@example.org", to, msg)
	if err != nil {
		glog.Fatal(err)
	}
}
