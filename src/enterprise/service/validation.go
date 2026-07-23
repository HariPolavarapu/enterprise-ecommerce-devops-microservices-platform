package service

import (
	"errors"
	"strings"
)

func ValidateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return errors.New("email is required")
	}
	if !strings.Contains(email, "@") {
		return errors.New("email is invalid")
	}
	return nil
}

func ValidateRequired(value string, field string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(field + " is required")
	}
	return nil
}
