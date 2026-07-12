package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Render rebuilds the full hosts-file text from a Parsed structure,
// rewriting the managed block from entries. When entries is empty and no
// block existed, no block is emitted; when a block existed it is preserved
// (even if empty) so future adds reuse it.
func (p *Parsed) Render(entries []Entry) string {
	var b strings.Builder
	for _, l := range p.Pre {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	if len(entries) > 0 || p.HasBlock {
		if len(p.Pre) > 0 && !endsWithBlank(p.Pre) {
			b.WriteByte('\n')
		}
		b.WriteString(beginMarker)
		b.WriteByte('\n')
		for _, e := range entries {
			b.WriteString(e.IP)
			for _, h := range e.Hosts {
				b.WriteByte(' ')
				b.WriteString(h)
			}
			b.WriteByte('\n')
		}
		b.WriteString(endMarker)
		b.WriteByte('\n')
	}
	for _, l := range p.Post {
		b.WriteString(l)
		b.WriteByte('\n')
	}
	return b.String()
}

func endsWithBlank(lines []string) bool {
	if len(lines) == 0 {
		return false
	}
	return strings.TrimSpace(lines[len(lines)-1]) == ""
}

// WriteFile atomically writes content to path, preserving the existing
// file's mode (defaulting to 0644). It writes to a temp file in the same
// directory, fsyncs, then renames over the target so a crash mid-write
// cannot corrupt /etc/hosts.
func WriteFile(path, content string) error {
	dir := filepath.Dir(path)
	mode := os.FileMode(0644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode()
	}

	tmp, err := os.CreateTemp(dir, ".vhoster.tmp.*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }

	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		cleanup()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		cleanup()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		cleanup()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}
