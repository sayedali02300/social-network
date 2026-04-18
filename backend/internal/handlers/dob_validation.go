package handlers

import (
	"errors"
	"fmt"
	"time"
)

const dobLayout = "2006-01-02"

func validateRealisticDateOfBirth(value string) error {
	dob, err := time.Parse(dobLayout, value)
	if err != nil {
		return errors.New("dateOfBirth must use YYYY-MM-DD format")
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	oldestAllowed := today.AddDate(-120, 0, 0)

	if dob.After(today) {
		return errors.New("dateOfBirth cannot be in the future")
	}

	if dob.Before(oldestAllowed) {
		return fmt.Errorf("dateOfBirth is too old; use a date on or after %s", oldestAllowed.Format(dobLayout))
	}

	thirteenYearsAgo := today.AddDate(-13, 0, 0)
	if dob.After(thirteenYearsAgo) {
		return errors.New("users must be at least 13 years old")
	}
	return nil
}
