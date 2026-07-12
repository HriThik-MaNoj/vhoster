package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Manager ties together parsing, mutation, backup, locking, and atomic
// writing for a single hosts file.
type Manager struct {
	HostsPath string
	BackupDir string
	Force     bool // skip shadowed-host check (--force flag)

	lockFile *os.File
}

// NewManager returns a Manager targeting the given hosts file and backup dir.
func NewManager(hostsPath, backupDir string) *Manager {
	return &Manager{HostsPath: hostsPath, BackupDir: backupDir}
}

// List returns the current managed entries (read-only; does not lock).
func (m *Manager) List() ([]Entry, error) {
	p, err := Parse(m.HostsPath)
	if err != nil {
		return nil, err
	}
	return p.Managed, nil
}

// BackupList returns available backups, newest first.
func (m *Manager) BackupList() ([]string, error) {
	return ListBackups(m.BackupDir)
}

// LatestBackupPath returns the newest backup path, or "" if none.
func (m *Manager) LatestBackupPath() (string, error) {
	return LatestBackup(m.BackupDir)
}

// Add merges hostStrs under ipStr into the managed block. It refuses
// protected localhost IPs, duplicate managed hosts, and hosts that already
// exist outside the managed block (silent shadowing). The file is backed
// up and atomically rewritten under an exclusive lock.
func (m *Manager) Add(ipStr string, hostStrs []string) error {
	if isReservedIP(ipStr) {
		return fmt.Errorf("IP %s is protected (localhost) — vhoster won't manage it", ipStr)
	}
	if err := m.lock(); err != nil {
		return err
	}
	defer m.unlock()

	p, err := Parse(m.HostsPath)
	if err != nil {
		return err
	}

	existing := make(map[string]bool)
	for _, e := range p.Managed {
		for _, h := range e.Hosts {
			existing[e.IP+" "+h] = true
		}
	}
	for _, h := range hostStrs {
		if existing[ipStr+" "+h] {
			return fmt.Errorf("host %q is already managed under %s", h, ipStr)
		}
	}
	if !m.Force {
		if shadowed := shadowedHosts(p, hostStrs); len(shadowed) > 0 {
			return fmt.Errorf("host(s) %v already exist outside the managed block — resolve manually or pick different names", shadowed)
		}
	}

	if _, err := Backup(m.HostsPath, m.BackupDir); err != nil {
		return err
	}

	entries := p.Managed
	found := false
	for i := range entries {
		if entries[i].IP == ipStr {
			entries[i].Hosts = appendUnique(entries[i].Hosts, hostStrs)
			found = true
			break
		}
	}
	if !found {
		cp := make([]string, len(hostStrs))
		copy(cp, hostStrs)
		entries = append(entries, Entry{IP: ipStr, Hosts: cp})
	}

	return WriteFile(m.HostsPath, p.Render(entries))
}

// Remove deletes hostnames from the managed block. If ip is non-empty,
// every host mapped to that IP is removed. Empty entries (no hosts left)
// are dropped entirely. The file is backed up before the change.
func (m *Manager) Remove(hostsToRemove []string, ip string) error {
	if err := m.lock(); err != nil {
		return err
	}
	defer m.unlock()

	p, err := Parse(m.HostsPath)
	if err != nil {
		return err
	}
	if _, err := Backup(m.HostsPath, m.BackupDir); err != nil {
		return err
	}

	var kept []Entry
	for _, e := range p.Managed {
		if ip != "" && e.IP == ip {
			continue
		}
		var remain []string
		for _, h := range e.Hosts {
			if contains(hostsToRemove, h) {
				continue
			}
			remain = append(remain, h)
		}
		if len(remain) > 0 {
			kept = append(kept, Entry{IP: e.IP, Hosts: remain})
		}
	}
	return WriteFile(m.HostsPath, p.Render(kept))
}

// RestoreLatest restores the most recent backup (after a safety backup).
func (m *Manager) RestoreLatest() error {
	if err := m.lock(); err != nil {
		return err
	}
	defer m.unlock()
	latest, err := LatestBackup(m.BackupDir)
	if err != nil {
		return err
	}
	if latest == "" {
		return fmt.Errorf("no backups found in %s", m.BackupDir)
	}
	return Restore(m.HostsPath, latest, m.BackupDir)
}

// RestoreBackup restores a specific backup file (after a safety backup).
func (m *Manager) RestoreBackup(path string) error {
	if err := m.lock(); err != nil {
		return err
	}
	defer m.unlock()
	return Restore(m.HostsPath, path, m.BackupDir)
}

// --- locking ---

func (m *Manager) lock() error {
	lp := lockPath(m.BackupDir)
	if err := os.MkdirAll(filepath.Dir(lp), 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	f, err := os.OpenFile(lp, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("open lock file: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return fmt.Errorf("another vhoster process is holding the lock (%w)", err)
	}
	m.lockFile = f
	return nil
}

func (m *Manager) unlock() {
	if m.lockFile != nil {
		_ = syscall.Flock(int(m.lockFile.Fd()), syscall.LOCK_UN)
		_ = m.lockFile.Close()
		m.lockFile = nil
	}
}

func lockPath(backupDir string) string {
	return filepath.Join(backupDir, "vhoster.lock")
}

// --- helpers ---

func shadowedHosts(p *Parsed, hosts []string) []string {
	want := make(map[string]bool, len(hosts))
	for _, h := range hosts {
		want[h] = true
	}
	var shadowed []string
	lines := make([]string, 0, len(p.Pre)+len(p.Post))
	lines = append(lines, p.Pre...)
	lines = append(lines, p.Post...)
	for _, line := range lines {
		if e, ok := parseEntryLine(line); ok {
			for _, h := range e.Hosts {
				if want[h] {
					shadowed = append(shadowed, h)
				}
			}
		}
	}
	return shadowed
}

func appendUnique(base, add []string) []string {
	seen := make(map[string]bool, len(base))
	for _, b := range base {
		seen[b] = true
	}
	for _, a := range add {
		if !seen[a] {
			base = append(base, a)
			seen[a] = true
		}
	}
	return base
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
