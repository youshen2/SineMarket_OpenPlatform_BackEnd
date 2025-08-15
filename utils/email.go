package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"path/filepath"

	"github.com/spf13/viper"
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
	port := viper.GetString("smtp.port")
	user := viper.GetString("smtp.user")
	password := viper.GetString("smtp.password")
	senderEmail := viper.GetString("smtp.sender_email")
	senderName := viper.GetString("smtp.sender_name")

	auth := smtp.PlainAuth("", user, password, host)

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", senderName, senderEmail)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	addr := fmt.Sprintf("%s:%s", host, port)
	err := smtp.SendMail(addr, auth, senderEmail, []string{to}, msg.Bytes())

	return err
}
