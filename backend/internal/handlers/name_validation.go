package handlers

import (
	"errors"
	"regexp"
)

const maxNameLength = 25

var lettersAndSpacesNameRegex = regexp.MustCompile(`^[A-Za-z]+(?:\s+[A-Za-z]+)*$`)

func validateRequiredLettersOnlyName(fieldLabel, value string) error {
	if value == "" {
		return errors.New(fieldLabel + " is required")
	}
	if len(value) > maxNameLength {
		return errors.New(fieldLabel + " must be 25 characters or fewer")
	}
	if !lettersAndSpacesNameRegex.MatchString(value) {
		return errors.New(fieldLabel + " must contain letters and spaces only")
	}
	return nil
}

func validateOptionalNickname(value string) error {
	if len(value) > maxNameLength {
		return errors.New("nickname must be 25 characters or fewer")
	}
	return nil
}
