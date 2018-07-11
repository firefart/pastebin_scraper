package main

import (
	"crypto/tls"
	"fmt"

	gomail "gopkg.in/gomail.v2"
)

func sendEmail(m *gomail.Message) error {
	debugOutput("sending mail")
	d := gomail.Dialer{Host: config.Mailserver, Port: config.Mailport}
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // nolint: gas
	err := d.DialAndSend(m)
	return err
}
func sendErrorMessage(errorMessage error) error {
	debugOutput("sending error mail")
	m := gomail.NewMessage()
	m.SetHeader("From", config.Mailfrom)
	m.SetHeader("To", config.Mailtoerror)
	m.SetHeader("Subject", "ERROR in pastebin_scraper")
	m.SetBody("text/plain", fmt.Sprintf("%v", errorMessage))

	err := sendEmail(m)
	return err
}
