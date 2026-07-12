package validate_test

import (
	"net"
	"testing"

	"github.com/HriThik-MaNoj/vhoster/internal/validate"
)

func TestValidateIP(t *testing.T) {
	tests := []struct {
		input string
		valid bool
		want  string
	}{
		{"", false, ""},
		{"10.0.0.5", true, "10.0.0.5"},
		{"192.168.1.1", true, "192.168.1.1"},
		{"0.0.0.0", false, ""},
		{"::", false, ""},
		{"::1", true, "::1"},
		{"2001:db8::1", true, "2001:db8::1"},
		{"not-an-ip", false, ""},
		{"256.256.256.256", false, ""},
	}
	for _, tc := range tests {
		ip, err := validate.ValidateIP(tc.input)
		got := err == nil
		if got != tc.valid {
			t.Errorf("ValidateIP(%q): valid=%v, want=%v; err=%v", tc.input, got, tc.valid, err)
		}
		if err == nil && ip.String() != tc.want {
			t.Errorf("ValidateIP(%q): got %s, want %s", tc.input, ip.String(), tc.want)
		}
		if err == nil {
			// Verify it's a valid net.IP
			_ = net.IP.String(ip)
		}
	}
}

func TestValidateHostname(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"api.local", true},
		{"admin.local", true},
		{"my-service.example.com", true},
		{"a", true},
		{"a-b.c", true},
		{"localhost", true},
		{"", false},
		{"foo_bar", false},
		{"-bad", false},
		{"bad-", false},
		{"a." + string(make([]byte, 64)), false},        // label > 63 chars
		{string(make([]byte, 254)), false},               // total > 253
		{"192.168.1", false},                             // all-numeric TLD
		{"a..b", false},                                  // empty label
	}
	for _, tc := range tests {
		err := validate.ValidateHostname(tc.input)
		got := err == nil
		if got != tc.valid {
			t.Errorf("ValidateHostname(%q): valid=%v, want=%v; err=%v", tc.input, got, tc.valid, err)
		}
	}
}

func TestValidateAdd(t *testing.T) {
	ip, hosts, errs := validate.ValidateAdd("10.0.0.5", []string{"api.local", "admin.local"})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
	if ip == nil || ip.String() != "10.0.0.5" {
		t.Fatalf("unexpected IP: %v", ip)
	}
	if len(hosts) != 2 || hosts[0] != "api.local" || hosts[1] != "admin.local" {
		t.Fatalf("unexpected hosts: %v", hosts)
	}
}

func TestValidateAddBad(t *testing.T) {
	_, _, errs := validate.ValidateAdd("bad-ip", []string{"api.local"})
	if len(errs) == 0 {
		t.Fatal("expected errors for bad IP")
	}
}

func TestValidateAddNoHosts(t *testing.T) {
	_, _, errs := validate.ValidateAdd("10.0.0.5", nil)
	if len(errs) == 0 {
		t.Fatal("expected errors for nil hosts")
	}
}

func TestParseHostField(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"api.local admin.local", 2},
		{"api.local,admin.local", 2},
		{"api.local, admin.local staging.local", 3},
		{"", 0},
		{"   ", 0},
	}
	for _, tc := range tests {
		got := validate.ParseHostField(tc.input)
		if len(got) != tc.want {
			t.Errorf("ParseHostField(%q): got %d, want %d (%v)", tc.input, len(got), tc.want, got)
		}
	}
}
