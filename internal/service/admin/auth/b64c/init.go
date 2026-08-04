package b64c

import "github.com/mojocn/base64Captcha"

// captcha 全局验证码实例
var captcha *base64Captcha.Captcha

// init 在包首次加载时自动初始化
func init() {
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
	store := base64Captcha.DefaultMemStore
	captcha = base64Captcha.NewCaptcha(driver, store)
}
