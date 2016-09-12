package main

import (
	m "net/mail"
	"regexp"
	"strings"
)

var (
	usernameMatcher = regexp.MustCompile(`^[0-9A-Za-z_-]{1,30}$`)
	hexKeyMatcher   = regexp.MustCompile(`^[0-9A-Fa-f]{32,64}$`)
)

func validEmail(s string) bool {
	if len(s) > 255 {
		return false
	}
	if strings.ContainsAny(s, `<>`) {
		return false
	}
	_, err := m.ParseAddress(s)
	if err != nil {
		return false
	}
	return true
}

func validUsername(s string) bool {
	return usernameMatcher.MatchString(s)
}

func validHexKey(s string) bool {
	return hexKeyMatcher.MatchString(s)
}
