package main

import (
	"crypto/tls"
	"fmt"

	gomail "gopkg.in/gomail.v2"
)

func sendEmail(config configuration, m *gomail.Message) error {
	debugOutput("sending mail")
	d := gomail.Dialer{Host: config.Mailserver, Port: config.Mailport}
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // nolint: gas
	err := d.DialAndSend(m)
	return err
}
func sendErrorMessage(config configuration, errorMessage error) error {
	debugOutput("sending error mail")
	m := gomail.NewMessage()
	m.SetHeader("From", config.Mailfrom)
	m.SetHeader("To", config.Mailtoerror)
	m.SetHeader("Subject", "ERROR in pastebin_scraper")
	m.SetBody("text/plain", fmt.Sprintf("%v", errorMessage))

	err := sendEmail(config, m)
	return err
}
