package hosts

// Markers delimit the vhoster-managed block in /etc/hosts. Everything
// between them is owned by vhoster and rewritten on each mutation.
const (
	beginMarker = "# BEGIN vhoster (managed - do not edit by hand)"
	endMarker   = "# END vhoster"
)

// reservedHosts are hostnames vhoster must never create or delete.
var reservedHosts = map[string]bool{
	"localhost": true,
}

// reservedIPs are IPs whose mappings are protected from management.
var reservedIPs = map[string]bool{
	"127.0.0.1": true,
	"::1":       true,
}

// Entry is a single managed /etc/hosts mapping: one IP with one or more
// canonical hostnames.
type Entry struct {
	IP    string
	Hosts []string
}

// isReservedIP reports whether ip is a protected localhost IP.
func isReservedIP(ip string) bool {
	return reservedIPs[ip]
}
