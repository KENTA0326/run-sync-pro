package mail

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/KENTA0326/run-sync-pro/internal/config"
)

// SMTPConfig は SMTP 送信設定（未設定なら IsEnabled false）。
type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

func LoadSMTPConfig() SMTPConfig {
	return SMTPConfig{
		Host:     config.ResolveString("", "SMTP_HOST", ""),
		Port:     config.ResolveString("", "SMTP_PORT", "587"),
		User:     config.ResolveString("", "SMTP_USER", ""),
		Password: config.ResolveString("", "SMTP_PASSWORD", ""),
		From:     config.ResolveString("", "SMTP_FROM", ""),
	}
}

func (c SMTPConfig) IsEnabled() bool {
	return strings.TrimSpace(c.Host) != "" && strings.TrimSpace(c.From) != ""
}

// SendPasswordResetEmail はリセットリンク付きメールを送る。
func (c SMTPConfig) SendPasswordResetEmail(to, resetURL string) error {
	subject := "RunSync Pro パスワード再設定"
	body := fmt.Sprintf("以下のリンクからパスワードを再設定してください（有効期限あり）。\n\n%s\n\n心当たりがない場合はこのメールを無視してください。", resetURL)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		c.From, to, subject, body))
	addr := fmt.Sprintf("%s:%s", c.Host, c.Port)
	auth := smtp.PlainAuth("", c.User, c.Password, c.Host)
	return smtp.SendMail(addr, auth, c.From, []string{to}, msg)
}
