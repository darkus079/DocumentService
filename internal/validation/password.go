package validation

import (
	"regexp"
	"unicode"
)

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return &ValidationError{Field: "password", Message: "password must be at least 8 characters long"}
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsDigit(char) {
			hasDigit = true
		} else if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower {
		return &ValidationError{Field: "password", Message: "password must contain at least 2 letters in different cases"}
	}

	if !hasDigit {
		return &ValidationError{Field: "password", Message: "password must contain at least 1 digit"}
	}

	if !hasSpecial {
		return &ValidationError{Field: "password", Message: "password must contain at least 1 special character"}
	}

	return nil
}

func ValidateLogin(login string) error {
	if len(login) < 8 {
		return &ValidationError{Field: "login", Message: "login must be at least 8 characters long"}
	}

	matched, err := regexp.MatchString("^[a-zA-Z0-9]+$", login)
	if err != nil {
		return &ValidationError{Field: "login", Message: "invalid login format"}
	}

	if !matched {
		return &ValidationError{Field: "login", Message: "login must contain only latin letters and digits"}
	}

	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) FieldName() string {
	return e.Field
}
