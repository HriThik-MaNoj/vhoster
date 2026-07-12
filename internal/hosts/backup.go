package hosts

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// maxBackups is how many timestamped backups are retained in backupDir.
const maxBackups = 10

// Backup copies src into backupDir as hosts.<unixnano>.bak and prunes
// older backups beyond maxBackups. It returns the new backup path, or
// "" when src did not exist (nothing to back up).
func Backup(src, backupDir string) (string, error) {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}
	in, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("open source: %w", err)
	}
	defer in.Close()

	dst := filepath.Join(backupDir, fmt.Sprintf("hosts.%d.bak", time.Now().UnixNano()))
	out, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("create backup: %w", err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		_ = os.Remove(dst)
		return "", fmt.Errorf("copy backup: %w", err)
	}
	if err := out.Close(); err != nil {
		return "", fmt.Errorf("close backup: %w", err)
	}

	pruneBackups(backupDir)
	return dst, nil
}

// ListBackups returns backup file paths in backupDir, newest first.
func ListBackups(backupDir string) ([]string, error) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "hosts.") || !strings.HasSuffix(e.Name(), ".bak") {
			continue
		}
		paths = append(paths, filepath.Join(backupDir, e.Name()))
	}
	sort.Slice(paths, func(i, j int) bool {
		// Parse embedded timestamps for correct ordering, falling back
		// to file modification time when parsing fails.
		ni := parseNanos(paths[i])
		nj := parseNanos(paths[j])
		if ni != nj {
			return ni > nj
		}
		fi, errI := os.Stat(paths[i])
		fj, errJ := os.Stat(paths[j])
		if errI != nil || errJ != nil {
			return paths[i] > paths[j]
		}
		return fi.ModTime().After(fj.ModTime())
	})
	return paths, nil
}

// LatestBackup returns the most recent backup path, or "" if none.
func LatestBackup(backupDir string) (string, error) {
	paths, err := ListBackups(backupDir)
	if err != nil {
		return "", err
	}
	if len(paths) == 0 {
		return "", nil
	}
	return paths[0], nil
}

// Restore copies the backup at backupPath over dst. The current dst is
// backed up first so a restore can itself be undone.
func Restore(dst, backupPath, backupDir string) error {
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}
	if _, err := os.Stat(dst); err == nil {
		if _, err := Backup(dst, backupDir); err != nil {
			return fmt.Errorf("safety backup before restore: %w", err)
		}
	}
	in, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("open backup: %w", err)
	}
	defer in.Close()
	data, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("read backup: %w", err)
	}
	return WriteFile(dst, string(data))
}

// parseNanos extracts the Unix-nano timestamp from a backup filename like
// hosts.<unixnano>.bak. It returns 0 when the name cannot be parsed.
func parseNanos(path string) int64 {
	base := filepath.Base(path)
	s := strings.TrimPrefix(base, "hosts.")
	s = strings.TrimSuffix(s, ".bak")
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func pruneBackups(backupDir string) {
	paths, err := ListBackups(backupDir)
	if err != nil || len(paths) <= maxBackups {
		return
	}
	for _, p := range paths[maxBackups:] {
		_ = os.Remove(p)
	}
}
