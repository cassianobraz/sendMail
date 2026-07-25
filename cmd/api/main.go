package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	htmlTemplate "text/template"

	"github.com/cassianobraz/sendMail/cmd/initializers"
	templateFS "github.com/cassianobraz/sendMail/internal/template"
	"gopkg.in/gomail.v2"
)

func init() {
	initializers.LoadEnvVariables()
}

func main() {
	host := os.Getenv("HOST")
	userMail := os.Getenv("USER_EMAIL")
	password := os.Getenv("PASSWORD_EMAIL")

	dialer := gomail.NewDialer(host, 587, userMail, password)

	msg := gomail.NewMessage()
	msg.SetHeader("From", userMail)
	msg.SetHeader("To", userMail)
	msg.SetHeader("Subject", "Sending with mail Go")
	msg.SetBody("text/html", getBody())

	img, err := templateFS.Files.ReadFile("gopher.png")
	if err != nil {
		panic(err)
	}

	msg.Embed("gopher.png", gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(img)
		return err
	}))

	if err := dialer.DialAndSend(msg); err != nil {
		panic(err)
	}

	fmt.Println("Sending message.")
}

func getBody() string {
	t := htmlTemplate.Must(
		htmlTemplate.ParseFS(templateFS.Files, "mail.html"),
	)

	var buff bytes.Buffer
	t.Execute(&buff, nil)

	return buff.String()
}
