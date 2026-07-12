package hosts_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HriThik-MaNoj/vhoster/internal/hosts"
)

func testHostsPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "hosts")
}

func testManager(t *testing.T) *hosts.Manager {
	t.Helper()
	dir := t.TempDir()
	return hosts.NewManager(filepath.Join(dir, "hosts"), filepath.Join(dir, "backups"))
}

func writeHosts(t *testing.T, path, content string) {
	t.Helper()
	_ = os.WriteFile(path, []byte(content), 0644)
}

func TestParseEmpty(t *testing.T) {
	mgr := testManager(t)
	entries, err := mgr.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestParseNoBlock(t *testing.T) {
	p, err := hosts.Parse(testHostsPath(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Managed) != 0 {
		t.Fatalf("expected no managed entries")
	}
}

func TestParseManaged(t *testing.T) {
	content := `127.0.0.1 localhost

# BEGIN vhoster (managed - do not edit by hand)
10.0.0.5 api.local admin.local
10.0.0.6 staging.local
# END vhoster

192.168.1.1 something`
	path := testHostsPath(t)
	writeHosts(t, path, content)
	p, err := hosts.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	if !p.HasBlock {
		t.Fatal("expected block marker")
	}
	if len(p.Managed) != 2 {
		t.Fatalf("expected 2 managed entries, got %d", len(p.Managed))
	}
	if p.Managed[0].IP != "10.0.0.5" || len(p.Managed[0].Hosts) != 2 {
		t.Fatalf("bad first entry: %+v", p.Managed[0])
	}
}

func TestAddAndList(t *testing.T) {
	mgr := testManager(t)
	if err := mgr.Add("10.0.0.5", []string{"api.local", "admin.local"}); err != nil {
		t.Fatal(err)
	}
	entries, err := mgr.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].IP != "10.0.0.5" || len(entries[0].Hosts) != 2 {
		t.Fatalf("bad entry: %+v", entries[0])
	}
}

func TestAddExistingHost(t *testing.T) {
	mgr := testManager(t)
	_ = mgr.Add("10.0.0.5", []string{"api.local"})
	err := mgr.Add("10.0.0.5", []string{"api.local"})
	if err == nil {
		t.Fatal("expected error for duplicate host")
	}
}

func TestAddMergeIP(t *testing.T) {
	mgr := testManager(t)
	_ = mgr.Add("10.0.0.5", []string{"api.local"})
	_ = mgr.Add("10.0.0.5", []string{"admin.local"})
	entries, _ := mgr.List()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if len(entries[0].Hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(entries[0].Hosts))
	}
}

func TestReservedIP(t *testing.T) {
	mgr := testManager(t)
	err := mgr.Add("127.0.0.1", []string{"test.local"})
	if err == nil {
		t.Fatal("expected error for reserved IP")
	}
}

func TestRemove(t *testing.T) {
	mgr := testManager(t)
	_ = mgr.Add("10.0.0.5", []string{"api.local", "admin.local"})
	if err := mgr.Remove([]string{"admin.local"}, ""); err != nil {
		t.Fatal(err)
	}
	entries, _ := mgr.List()
	if len(entries) != 1 || len(entries[0].Hosts) != 1 || entries[0].Hosts[0] != "api.local" {
		t.Fatalf("bad state after remove: %+v", entries)
	}
}

func TestRemoveByIP(t *testing.T) {
	mgr := testManager(t)
	_ = mgr.Add("10.0.0.5", []string{"api.local"})
	_ = mgr.Add("10.0.0.6", []string{"other.local"})
	if err := mgr.Remove(nil, "10.0.0.5"); err != nil {
		t.Fatal(err)
	}
	entries, _ := mgr.List()
	if len(entries) != 1 || entries[0].IP != "10.0.0.6" {
		t.Fatalf("expected only 10.0.0.6: %+v", entries)
	}
}

func TestBackupRestore(t *testing.T) {
	mgr := testManager(t)
	// Seed a hosts file so the first Add creates a backup of it.
	_ = os.WriteFile(mgr.HostsPath, []byte("127.0.0.1 localhost\n"), 0644)
	_ = mgr.Add("10.0.0.5", []string{"api.local"})

	paths, err := mgr.BackupList()
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) < 1 {
		t.Fatal("expected at least 1 backup")
	}
	latest, err := mgr.LatestBackupPath()
	if err != nil || latest == "" {
		t.Fatal("expected latest backup")
	}

	// Make a change, restore
	_ = mgr.Add("10.0.0.6", []string{"staging.local"})
	entries, _ := mgr.List()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries before restore")
	}

	if err := mgr.RestoreLatest(); err != nil {
		t.Fatal(err)
	}
	entries, _ = mgr.List()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after restore, got %d", len(entries))
	}
}

func TestBackupRotation(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "hosts")
	_ = os.WriteFile(src, []byte("127.0.0.1 localhost\n"), 0644)

	// Call Backup 15 times — the last 5 should be pruned.
	for range 15 {
		_, _ = hosts.Backup(src, filepath.Join(dir, "backups"))
	}

	paths, err := hosts.ListBackups(filepath.Join(dir, "backups"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) > 10 {
		t.Fatalf("expected at most 10 backups after rotation, got %d", len(paths))
	}
}

func TestUnmanagedPreserved(t *testing.T) {
	content := `127.0.0.1 localhost

# some comment
192.168.1.1 existing.local`

	mgr := testManager(t)
	writeHosts(t, mgr.HostsPath, content)
	_ = mgr.Add("10.0.0.5", []string{"api.local"})

	data, _ := os.ReadFile(mgr.HostsPath)
	if !strings.Contains(string(data), "192.168.1.1 existing.local") {
		t.Fatal("unmanaged lines not preserved")
	}
	if !strings.Contains(string(data), "api.local") {
		t.Fatal("api.local not in output")
	}
}

func TestAddToExistingBlock(t *testing.T) {
	content := `127.0.0.1 localhost

# BEGIN vhoster (managed - do not edit by hand)
10.0.0.5 api.local
# END vhoster

192.168.1.1 something.local`

	mgr := testManager(t)
	writeHosts(t, mgr.HostsPath, content)
	_ = mgr.Add("10.0.0.5", []string{"admin.local"})

	p, _ := hosts.Parse(mgr.HostsPath)
	if len(p.Managed) != 1 || len(p.Managed[0].Hosts) != 2 {
		t.Fatalf("expected 2 hosts under 10.0.0.5: %+v", p.Managed)
	}
}
