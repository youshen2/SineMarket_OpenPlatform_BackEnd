package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

func ParseTemplate(templateFileName string, data interface{}) (string, error) {
	templatePath, err := filepath.Abs(fmt.Sprintf("templates/%s", templateFileName))
	if err != nil {
		return "", err
	}

	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func SendEmail(to, subject, body string) error {
	host := viper.GetString("smtp.host")
	port := viper.GetInt("smtp.port")
	user := viper.GetString("smtp.user")
	password := viper.GetString("smtp.password")
	senderEmail := viper.GetString("smtp.sender_email")
	senderName := viper.GetString("smtp.sender_name")

	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(senderEmail, senderName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(host, port, user, password)

	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		errChan <- d.DialAndSend(m)
	}()

	return nil
}
