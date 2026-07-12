package validate

import (
	"errors"
	"fmt"
	"strings"
)

// ValidateHostname checks s against RFC 1123: dot-separated labels, each
// 1–63 chars, beginning and ending with an alphanumeric character with
// interior hyphens allowed, total length ≤ 253, and a TLD that is not
// entirely numeric. Underscores are rejected.
func ValidateHostname(s string) error {
	if s == "" {
		return errors.New("empty hostname")
	}
	if len(s) > 253 {
		return fmt.Errorf("hostname %q too long (%d > 253)", s, len(s))
	}
	labels := strings.Split(s, ".")
	for i, label := range labels {
		if err := validateLabel(label); err != nil {
			return fmt.Errorf("hostname %q: %w", s, err)
		}
		if i == len(labels)-1 && isAllDigits(label) {
			return fmt.Errorf("hostname %q: TLD must not be all-numeric", s)
		}
	}
	return nil
}

func validateLabel(label string) error {
	n := len(label)
	if n == 0 {
		return errors.New("empty label")
	}
	if n > 63 {
		return fmt.Errorf("label %q too long (%d > 63)", label, n)
	}
	if !isAlphaNum(label[0]) {
		return fmt.Errorf("label %q must start with an alphanumeric character", label)
	}
	if !isAlphaNum(label[n-1]) {
		return fmt.Errorf("label %q must end with an alphanumeric character", label)
	}
	for i := 1; i < n-1; i++ {
		c := label[i]
		if !isAlphaNum(c) && c != '-' {
			return fmt.Errorf("label %q contains invalid character %q", label, c)
		}
	}
	return nil
}

func isAlphaNum(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9')
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
