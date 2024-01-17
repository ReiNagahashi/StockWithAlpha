package controllers

import (
	"bytes"
	"fmt"
	"net/smtp"
	"stock-with-alpha/config"
	"strconv"
	"text/template"
)

type EmailTemplate struct{
	Subject string
	Body 	string
}

func (et *EmailTemplate) Send(to string) error{
	host := config.Config.SmtpHostName
	port := config.Config.SmtpPort
	email := config.Config.Email
	auth := smtp.PlainAuth("", email, config.Config.SmtpPassword, host)

	t, err := template.New("email").Parse("Subject: {{.Subject}}\n\n{{.Body}}")
	if err != nil{
		return err
	}

	var body bytes.Buffer
	err = t.Execute(&body, et)
	if err != nil{
		return err
	}

	toAddresses := []string{to}

	msg := []byte("From:" + email + "\r\n" + 
					"To: " + to + "\r\n" + 
					body.String())

	err = smtp.SendMail(host + ":" + strconv.Itoa(port), auth, email, toAddresses, msg)
	if err != nil{
		return err
	}
	
	fmt.Println("Email successfully sent!")

	return err

}
