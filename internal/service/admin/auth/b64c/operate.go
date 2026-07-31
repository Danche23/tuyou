package b64c

import appErrors "tuyou/pkg/errors"

// Generate 生成验证码，返回验证码 ID 和 Base64 图片
func Generate() (id, b64s string, err error) {
	id, b64s, _, err = captcha.Generate()
	return
}

// Identify 校验验证码，验证通过返回 nil，失败返回业务错误
func Identify(captchaId, captcha1 string) error {
	if captchaId == "" || captcha1 == "" {
		return appErrors.ErrMissingParam
	}
	if !captcha.Verify(captchaId, captcha1, true) {
		return appErrors.ErrCaptchaInvalid
	}
	return nil
}
