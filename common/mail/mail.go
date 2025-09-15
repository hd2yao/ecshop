package mail

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"text/template"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// MailService 邮件服务
type MailService struct {
	smtp     *SMTPClient
	store    *MailCodeRedisStore
	config   MailConfig
	template *template.Template
}

// NewMailService 创建邮件服务
func NewMailService(config MailConfig, module string, expiration time.Duration) (*MailService, error) {
	// 创建SMTP客户端
	smtp := NewSMTPClient(config.SMTP)

	// 创建Redis存储
	store := NewMailCodeRedisStore(module, expiration)

	// 解析邮件模板
	tmpl, err := template.New("verify_code").Parse(config.Template.VerifyCodeBody)
	if err != nil {
		return nil, fmt.Errorf("解析邮件模板失败: %v", err)
	}

	return &MailService{
		smtp:     smtp,
		store:    store,
		config:   config,
		template: tmpl,
	}, nil
}

// VerifyCodeData 验证码邮件模板数据
type VerifyCodeData struct {
	Code          string // 验证码
	ExpireMinutes int    // 过期时间（分钟）
	Email         string // 邮箱地址
}

// SendVerifyCode 发送验证码邮件
func (m *MailService) SendVerifyCode(email string, codeLength int) (string, error) {
	// 验证邮箱格式
	if !isValidEmail(email) {
		return "", fmt.Errorf("邮箱格式不正确")
	}

	// 检查是否频繁发送（可选：增加发送间隔限制）
	if m.store.Exists(email) {
		ttl := m.store.GetTTL(email)
		if ttl > 9*time.Minute { // 如果还有9分钟以上的有效期，说明刚发送不久
			return "", fmt.Errorf("验证码发送过于频繁，请稍后再试")
		}
	}

	// 生成验证码
	code := generateRandomCode(codeLength)

	// 准备邮件模板数据
	data := VerifyCodeData{
		Code:          code,
		ExpireMinutes: 10, // 10分钟有效期
		Email:         email,
	}

	// 渲染邮件内容
	var body bytes.Buffer
	if err := m.template.Execute(&body, data); err != nil {
		return "", fmt.Errorf("渲染邮件模板失败: %v", err)
	}

	// 发送邮件
	if err := m.smtp.SendEmail([]string{email}, m.config.Template.VerifyCodeSubject, body.String()); err != nil {
		return "", fmt.Errorf("发送邮件失败: %v", err)
	}

	// 保存验证码到Redis
	if err := m.store.Set(email, code); err != nil {
		logx.Errorf("保存验证码到Redis失败: %v", err)
		// 注意：邮件已发送，即使Redis保存失败也返回成功，但记录日志
	}

	logx.Infof("验证码邮件发送成功: email=%s", email)
	return code, nil
}

// VerifyCode 验证邮件验证码
func (m *MailService) VerifyCode(email, code string) bool {
	if email == "" || code == "" {
		return false
	}

	// 验证邮箱格式
	if !isValidEmail(email) {
		return false
	}
	
	return m.store.Verify(email, code, true)
}

// CheckCodeExists 检查验证码是否存在
func (m *MailService) CheckCodeExists(email string) bool {
	return m.store.Exists(email)
}

// GetCodeTTL 获取验证码剩余有效时间
func (m *MailService) GetCodeTTL(email string) time.Duration {
	return m.store.GetTTL(email)
}

// generateRandomCode 生成随机数字验证码
func generateRandomCode(length int) string {
	if length <= 0 {
		length = 6 // 默认6位
	}

	rand.Seed(time.Now().UnixNano())
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		code[i] = byte(rand.Intn(10) + '0') // 生成0-9的数字
	}
	return string(code)
}

// isValidEmail 简单的邮箱格式验证
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}

	// 基本格式检查：包含@符号，且@前后都有内容
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}

	// 检查域名部分包含点号
	if !strings.Contains(parts[1], ".") {
		return false
	}

	return true
}
