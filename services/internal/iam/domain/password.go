package domain

import "regexp"

var (
	passwordLower = regexp.MustCompile(`[a-z]`)
	passwordUpper = regexp.MustCompile(`[A-Z]`)
	passwordDigit = regexp.MustCompile(`[0-9]`)
)

// ValidatePassword checks password against policy requirements.
// Returns ErrPasswordPolicy if the password doesn't meet the requirements.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordPolicy
	}
	if !passwordLower.MatchString(password) {
		return ErrPasswordPolicy
	}
	if !passwordUpper.MatchString(password) {
		return ErrPasswordPolicy
	}
	if !passwordDigit.MatchString(password) {
		return ErrPasswordPolicy
	}
	return nil
}
