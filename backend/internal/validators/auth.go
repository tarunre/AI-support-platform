package validators

import (
	"errors"
	"unicode"
)

// ValidatePasswordStrength enforces complexity rules on top of the min-length
// binding tag. A valid password must contain at least one uppercase letter,
// one lowercase letter, one digit, and one special character.
func ValidatePasswordStrength(password string) error {
	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	switch {
	case !hasUpper:
		return errors.New("password must contain at least one uppercase letter")
	case !hasLower:
		return errors.New("password must contain at least one lowercase letter")
	case !hasDigit:
		return errors.New("password must contain at least one digit")
	case !hasSpecial:
		return errors.New("password must contain at least one special character")
	}
	return nil
}
