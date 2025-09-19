package mail

import (
	"crypto/tls"
	"fmt"

	"gopkg.in/gomail.v2"
)

// SMTPClient SMTP客户端
type SMTPClient struct {
	config SMTPConfig
}

// NewSMTPClient 创建SMTP客户端
func NewSMTPClient(config SMTPConfig) *SMTPClient {
	return &SMTPClient{
		config: config,
	}
}

// SendEmail 发送邮件
func (s *SMTPClient) SendEmail(to []string, subject, body string) error {
	// 创建邮件消息
	m := gomail.NewMessage()

	// 设置发件人
	m.SetHeader("From", m.FormatAddress(s.config.From, s.config.FromName))

	// 设置收件人
	m.SetHeader("To", to...)

	// 设置邮件主题
	m.SetHeader("Subject", subject)

	// 设置邮件正文（HTML格式）
	m.SetBody("text/html", body)

	// 创建SMTP拨号器
	d := gomail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)

	// 对于QQ邮箱等，需要设置TLS配置
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	return nil
}
