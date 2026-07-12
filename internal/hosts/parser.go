package hosts

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

// Parsed is a /etc/hosts file split into unmanaged lines (preserved
// verbatim) around the managed block, plus the managed entries.
type Parsed struct {
	Pre      []string // unmanaged lines before the block
	Post     []string // unmanaged lines after the block
	Managed  []Entry  // entries inside the managed block
	HasBlock bool     // whether BEGIN/END markers were found
}

// Parse reads the hosts file at path and splits it. A missing file yields
// an empty Parsed so the first write creates it.
func Parse(path string) (*Parsed, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Parsed{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return parseBytes(data), nil
}

func parseBytes(data []byte) *Parsed {
	p := &Parsed{}
	inBlock := false
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		switch strings.TrimSpace(line) {
		case beginMarker:
			inBlock = true
			p.HasBlock = true
			continue
		case endMarker:
			inBlock = false
			continue
		}
		if inBlock {
			if e, ok := parseEntryLine(line); ok {
				p.Managed = append(p.Managed, e)
			}
			// blank/comment lines inside the block are dropped: vhoster owns it
			continue
		}
		if p.HasBlock {
			p.Post = append(p.Post, line)
		} else {
			p.Pre = append(p.Pre, line)
		}
	}
	return p
}

// parseEntryLine parses "IP host1 host2 ..." into an Entry. Lines that are
// blank, comments, or lack an IP plus at least one host are ignored.
func parseEntryLine(line string) (Entry, bool) {
	if i := strings.Index(line, "#"); i >= 0 {
		line = line[:i]
	}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return Entry{}, false
	}
	return Entry{IP: fields[0], Hosts: fields[1:]}, true
}
