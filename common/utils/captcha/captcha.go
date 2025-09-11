package captcha

import "github.com/mojocn/base64Captcha"

var store = base64Captcha.DefaultMemStore

// Generate 生成验证码
func Generate(driverType string, opt CustomOptions) (id, b64s, answer string, err error) {
	driver := NewDriver(driverType, opt)
	c := base64Captcha.NewCaptcha(driver, store)
	return c.Generate()
}

// Verify 校验验证码
func Verify(id, value string) bool {
	return store.Verify(id, value, true)
}
