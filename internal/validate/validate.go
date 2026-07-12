package validate

import (
	"net"
	"strings"
)

// ValidateAdd validates an IP string plus a set of raw hostname strings
// for an add operation. It returns the parsed IP, a cleaned and de-duplicated
// list of hostnames (order preserved), and a slice of human-readable error
// strings (empty when everything is valid).
func ValidateAdd(ipStr string, hostStrs []string) (net.IP, []string, []string) {
	var errs []string

	ip, err := ValidateIP(ipStr)
	if err != nil {
		errs = append(errs, err.Error())
	}

	seen := make(map[string]bool)
	var hosts []string
	for _, h := range hostStrs {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		if !seen[h] {
			seen[h] = true
			hosts = append(hosts, h)
		}
	}
	if len(hosts) == 0 {
		errs = append(errs, "no hostnames provided")
	}
	for _, h := range hosts {
		if err := ValidateHostname(h); err != nil {
			errs = append(errs, err.Error())
		}
	}

	return ip, hosts, errs
}

// ParseHostField splits a single input string (as typed in the TUI host
// field or passed on the CLI) that may contain space- and/or comma-
// separated hostnames into a clean slice.
func ParseHostField(s string) []string {
	s = strings.ReplaceAll(s, ",", " ")
	var out []string
	for _, part := range strings.Fields(s) {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// FormatErrors joins a slice of error strings into a single message.
func FormatErrors(errs []string) string {
	return strings.Join(errs, "\n")
}
