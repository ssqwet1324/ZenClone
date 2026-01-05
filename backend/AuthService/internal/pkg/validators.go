package pkg

import (
	"fmt"
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	passwordRegex8 = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*]{8,}$`)
)

// PasswordRegex8 - валидатор для пароля минимум 8 символов
func PasswordRegex8(fl validator.FieldLevel) bool {
	return passwordRegex8.MatchString(fl.Field().String())
}

// Has4EnLetters  - проверка на 4 символа
func Has4EnLetters(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	count := 0
	for _, r := range s {
		if unicode.IsLetter(r) && unicode.Is(unicode.Latin, r) {
			count++
			if count >= 4 {
				return true
			}
		}
	}
	return false
}

// Has2Letters  - проверка на 2 символа
func Has2Letters(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	count := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			count++
			if count >= 2 {
				return true
			}
		}
	}
	return false
}

// Has1Letters  - проверка на 1 символ
func Has1Letters(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	count := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			count++
			if count >= 1 {
				return true
			}
		}
	}
	return false
}

// Has3Letters  - проверка на 3 символа
func Has3Letters(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	count := 0
	for _, r := range s {
		if unicode.IsLetter(r) && unicode.Is(unicode.Latin, r) {
			count++
			if count >= 3 {
				return true
			}
		}
	}
	return false
}

// ValidationErrorToMessage - ошибки при валидации
func ValidationErrorToMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return fmt.Sprintf("Minimum %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("Maximum %s characters", fe.Param())
	case "alphanum":
		return "Only letters and numbers allowed"
	case "has3letters":
		return "At least 3 letters required"
	case "has4enletters":
		return "At least 4 English letters required"
	case "has2letters":
		return "At least 2 letters required"
	case "passwordregex8":
		return "Password too weak (minimum 8 characters)"
	case "has1letters":
		return "At least 1 letter required"
	default:
		return "Invalid value"
	}
}
