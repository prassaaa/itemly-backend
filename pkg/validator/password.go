package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	hasUppercase = regexp.MustCompile(`[A-Z]`)
	hasLowercase = regexp.MustCompile(`[a-z]`)
	hasDigit     = regexp.MustCompile(`[0-9]`)
	hasSpecial   = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`]")
)

func PasswordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return hasUppercase.MatchString(password) &&
		hasLowercase.MatchString(password) &&
		hasDigit.MatchString(password) &&
		hasSpecial.MatchString(password)
}
