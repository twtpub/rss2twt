package main

import (
	"errors"
	"regexp"
	"strings"
)

var (
	validName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\- ]*$`)

	ErrInvalidName = errors.New("error: invalid feed name")
)

func NormalizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

func ValidateName(name string) error {
	if !validName.MatchString(name) {
		return ErrInvalidName
	}
	return nil
}
