package main

import (
	"bytes"
	"crypto/tls"
	"fmt"

	log "github.com/sirupsen/logrus"
	gomail "gopkg.in/gomail.v2"
)

func sendEmail(config configuration, m *gomail.Message) error {
	log.Debugf("sending mail")
	if *test {
		text, err := messageToString(m)
		if err != nil {
			return fmt.Errorf("could not print mail: %v", err)
		}
		log.Printf("[MAIL] %s", text)
		return nil
	}
	d := gomail.Dialer{Host: config.Mailserver, Port: config.Mailport, Username: config.MailUsername, Password: config.MailPassword}
	if config.MailSkipTLS {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // nolint: gosec
	}
	return d.DialAndSend(m)
}

func sendErrorMessage(config configuration, errorMessage error) error {
	log.Debugf("sending error mail")
	m := gomail.NewMessage()
	m.SetHeader("From", config.Mailfrom)
	m.SetHeader("To", config.Mailtoerror)
	m.SetHeader("Subject", "ERROR in pastebin_scraper")
	m.SetBody("text/plain", fmt.Sprintf("%v", errorMessage))

	err := sendEmail(config, m)
	return err
}

func messageToString(m *gomail.Message) (string, error) {
	buf := new(bytes.Buffer)
	_, err := m.WriteTo(buf)
	if err != nil {
		return "", fmt.Errorf("could not convert message to string: %v", err)
	}
	return buf.String(), nil
}
