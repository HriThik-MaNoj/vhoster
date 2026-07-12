package validate

import (
	"errors"
	"fmt"
	"net"
)

// ValidateIP parses s as an IPv4 or IPv6 address. The empty string, the
// nil parse result, and the unspecified addresses 0.0.0.0 / :: are
// rejected. The canonical form is returned via net.IP.String() by callers.
func ValidateIP(s string) (net.IP, error) {
	if s == "" {
		return nil, errors.New("empty IP address")
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address %q", s)
	}
	if ip.IsUnspecified() {
		return nil, fmt.Errorf("IP address %q is unspecified (0.0.0.0/:: not allowed)", s)
	}
	return ip, nil
}
