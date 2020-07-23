package main

import (
	"errors"
	"regexp"
	"strings"
)

const (
	maxNameLength = 25 // avg 4.7 chars per word in English so ~5 words
)

var (
	validName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\- ]*$`)

	ErrInvalidName = errors.New("error: invalid feed name")
	ErrNameTooLong = errors.New("error: name is too long")
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
	if len(name) > maxNameLength {
		return ErrNameTooLong
	}
	return nil
}
