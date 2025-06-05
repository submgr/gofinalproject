package utils

import (
	"fmt"
	"os"

	"gopkg.in/gomail.v2"
)

func SendRecoveryCode(email, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_FROM"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Восстановление пароля - Доска объявлений")
	m.SetBody("text/plain", fmt.Sprintf("Ваш код для восстановления пароля: %s\n\nЕсли вы не запрашивали восстановление пароля, проигнорируйте это письмо.", code))

	d := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		587,
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASS"),
	)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
} 