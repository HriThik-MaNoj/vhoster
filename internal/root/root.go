// Package root handles privilege checks for vhoster.
package root

import (
	"fmt"
	"os"
)

// IsRoot reports whether the process is running as root (euid 0).
func IsRoot() bool {
	return os.Geteuid() == 0
}

// EnsureRoot exits with code 2 and a helpful message when the process is
// not root. If skip is true (a non-default --hosts path is in use for
// local testing) the requirement is waived.
func EnsureRoot(skip bool) {
	if skip {
		return
	}
	if !IsRoot() {
		fmt.Fprintln(os.Stderr, "vhoster: run with sudo — this command modifies /etc/hosts")
		os.Exit(2)
	}
}
