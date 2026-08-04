package sms

import (
	"regexp"
	"tuyou/pkg/errors"
)

var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

func ValidatorPhone(phone string) error {
	if !phoneRegex.MatchString(phone) {
		return errors.ErrInvalidPhoneFormat
	}
	return nil
}
