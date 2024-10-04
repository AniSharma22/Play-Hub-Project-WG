package validation

import (
	"github.com/go-playground/validator"
	"regexp"
	"strings"
	"unicode"
)

// checks if the email is of the format name.surname@watchguard.com
func IsValidEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	re := regexp.MustCompile(`^[a-zA-Z]+\.[a-zA-Z]+@watchguard\.com$`)
	return re.MatchString(strings.ToLower(email))
}

// Password must have 1 uppercase, 1 lowercase, 1 special character and minimum 8 length
func IsValidPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	var hasUpper, hasLower, hasSpecial bool

	if len(password) < 8 {
		return false
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasSpecial
}

// checks if length of number is 10 and starts with 6,7,8 or 9
func IsValidPhoneNumber(fl validator.FieldLevel) bool {
	phoneNumber := fl.Field().String()
	re := regexp.MustCompile(`^[6-9]\d{9}$`)
	return re.MatchString(phoneNumber)
}

func IsValidGender(fl validator.FieldLevel) bool {
	gender := fl.Field().String()
	gender = strings.ToLower(gender)
	return gender == "male" || gender == "female" || gender == "other"
}
