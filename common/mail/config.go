package mail

// MailConfig 邮件配置
type MailConfig struct {
	SMTP     SMTPConfig `json:"smtp"`
	Template Template   `json:"template"`
}

// SMTPConfig SMTP服务器配置
type SMTPConfig struct {
	Host     string `json:"host"`     // SMTP服务器地址
	Port     int    `json:"port"`     // SMTP服务器端口
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码或授权码
	From     string `json:"from"`     // 发件人邮箱
	FromName string `json:"from_name"` // 发件人名称
}

// Template 邮件模板配置
type Template struct {
	VerifyCodeSubject string `json:"verify_code_subject"` // 验证码邮件主题
	VerifyCodeBody    string `json:"verify_code_body"`    // 验证码邮件内容模板
}

// DefaultMailConfig 默认邮件配置
func DefaultMailConfig() MailConfig {
	return MailConfig{
		SMTP: SMTPConfig{
			Host:     "smtp.qq.com",
			Port:     587,
			FromName: "EcShop验证服务",
		},
		Template: Template{
			VerifyCodeSubject: "【EcShop】邮箱验证码",
			VerifyCodeBody: `
亲爱的用户，您好！

您的邮箱验证码是：{{.Code}}

验证码有效期为 {{.ExpireMinutes}} 分钟，请尽快完成验证。

如果这不是您的操作，请忽略此邮件。

感谢您使用 EcShop！
`,
		},
	}
}
