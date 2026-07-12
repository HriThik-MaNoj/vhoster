package tui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (m model) View() string {
	switch m.state {
	case viewAddForm:
		return m.viewAddForm()
	case viewConfirm:
		return m.viewConfirm()
	case viewBackups:
		return m.viewBackups()
	default:
		return m.viewList()
	}
}

func (m model) viewList() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("vhoster — managed /etc/hosts entries"))
	b.WriteString("\n\n")

	if len(m.entries) == 0 {
		b.WriteString(dimStyle.Render("  no managed entries — press 'a' to add"))
		b.WriteString("\n")
	} else {
		for i, e := range m.entries {
			line := fmt.Sprintf("%-16s", e.IP)
			// Render each host, highlighting the selected one
			for j, h := range e.Hosts {
				if j > 0 {
					line += ","
				}
				if i == m.cursor && (m.hostCursor == j || m.hostCursor < 0) && len(e.Hosts) > 1 {
					line += fmt.Sprintf(" [%s]", h)
				} else {
					line += " " + h
				}
			}
			if i == m.cursor {
				b.WriteString(selectedStyle.Render("▶ " + line))
			} else {
				b.WriteString(entryStyle.Render("  " + line))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if m.err != "" {
		b.WriteString(errorStyle.Render("✖ " + m.err))
		b.WriteString("\n")
	} else if m.info != "" {
		b.WriteString(okStyle.Render("✓ " + m.info))
		b.WriteString("\n")
	}
	b.WriteString(hintStyle.Render("a:add  d:remove  b:backups  r:refresh  q:quit"))
	return b.String()
}

func (m model) viewAddForm() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("vhoster — add vhosts"))
	b.WriteString("\n\n")
	b.WriteString("IP address:\n")
	b.WriteString(m.ipInput.View())
	b.WriteString("\n\nVhosts (space or comma separated):\n")
	b.WriteString(m.hostsInput.View())
	b.WriteString("\n\n")
	if m.formErr != "" {
		b.WriteString(errorStyle.Render("✖ " + m.formErr))
	} else {
		b.WriteString(dimStyle.Render("enter: submit   tab: next field   esc: cancel"))
	}
	b.WriteString("\n")
	return b.String()
}

func (m model) viewConfirm() string {
	body := m.confirmMsg + "\n\ny: confirm    n / esc: cancel"
	return boxStyle.Render(body)
}

func (m model) viewBackups() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("vhoster — restore from backup"))
	b.WriteString("\n\n")
	if len(m.backups) == 0 {
		b.WriteString(dimStyle.Render("  no backups available"))
		b.WriteString("\n")
	} else {
		for i, p := range m.backups {
			label := formatBackupTime(p)
			if i == m.bcursor {
				b.WriteString(selectedStyle.Render("▶ " + label))
			} else {
				b.WriteString(entryStyle.Render("  " + label))
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("enter: restore   esc: back"))
	b.WriteString("\n")
	return b.String()
}

// formatBackupTime parses a "hosts.<unixnano>.bak" path into a
// human-readable relative timestamp.
func formatBackupTime(path string) string {
	base := filepath.Base(path)
	s := strings.TrimPrefix(base, "hosts.")
	s = strings.TrimSuffix(s, ".bak")
	nanos, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return path
	}
	t := time.Unix(0, nanos)
	elapsed := time.Since(t)
	if elapsed < 0 {
		elapsed = 0
	}
	var rel string
	switch {
	case elapsed < time.Minute:
		rel = "just now"
	case elapsed < time.Hour:
		m := int(elapsed.Minutes())
		if m == 1 {
			rel = "1 minute ago"
		} else {
			rel = fmt.Sprintf("%d minutes ago", m)
		}
	case elapsed < 24*time.Hour:
		h := int(elapsed.Hours())
		m := int(elapsed.Minutes()) % 60
		if h == 1 {
			rel = fmt.Sprintf("1 hour %d min ago", m)
		} else {
			rel = fmt.Sprintf("%d hours %d min ago", h, m)
		}
	default:
		d := int(elapsed.Hours() / 24)
		if d == 1 {
			rel = "yesterday"
		} else {
			rel = fmt.Sprintf("%d days ago", d)
		}
	}
	return fmt.Sprintf("%s (%s)", filepath.Base(path), rel)
}
