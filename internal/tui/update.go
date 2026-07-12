package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/HriThik-MaNoj/vhoster/internal/hosts"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		w := max(20, msg.Width-4)
		m.ipInput.Width = w
		m.hostsInput.Width = w
		return m, nil
	case reloadMsg:
		m.reload()
		return m, nil
	case backupsMsg:
		m.backups = msg.paths
		m.bcursor = 0
		m.state = viewBackups
		return m, nil
	case errMsg:
		m.err = msg.err
		m.info = ""
		return m, nil
	case infoMsg:
		m.info = msg.text
		m.err = ""
		m.reload()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m.forwardInputs(msg)
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		if m.state != viewList {
			m.state = viewList
			m.err = ""
			m.info = ""
			return m, nil
		}
		return m, tea.Quit
	}

	switch m.state {
	case viewList:
		return m.handleListKey(msg)
	case viewAddForm:
		return m.handleFormKey(msg)
	case viewConfirm:
		return m.handleConfirmKey(msg)
	case viewBackups:
		return m.handleBackupsKey(msg)
	}
	return m, nil
}

func (m model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	n := len(m.entries)
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "a":
		m.state = viewAddForm
		m.ipInput.Reset()
		m.hostsInput.Reset()
		m.ipInput.Focus()
		m.hostsInput.Blur()
		m.formField = 0
		m.formErr = ""
		return m, textinput.Blink
	case "d", "x":
		if n == 0 {
			return m, nil
		}
		e := m.entries[m.cursor]
		mgr := m.manager
		if m.hostCursor >= 0 && len(e.Hosts) > 1 {
			// Remove only the one host
			host := e.Hosts[m.hostCursor]
			m.confirmMsg = fmt.Sprintf("Remove %s from %s?", host, e.IP)
			m.confirmAction = func() tea.Msg { return doRemove(mgr, []string{host}, "") }
		} else {
			// Remove all hosts under this IP
			m.confirmMsg = fmt.Sprintf("Remove all hosts under %s?", e.IP)
			m.confirmAction = func() tea.Msg { return doRemove(mgr, nil, e.IP) }
		}
		m.state = viewConfirm
		return m, nil
	case "b":
		return m, m.loadBackupsCmd()
	case "r":
		m.err = ""
		m.info = ""
		m.reload()
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.hostCursor = -1
			m.err = ""
			m.info = ""
		}
	case "down", "j":
		if m.cursor < n-1 {
			m.cursor++
			m.hostCursor = -1
			m.err = ""
			m.info = ""
		}
	case "left", "h":
		if n > 0 && len(m.entries[m.cursor].Hosts) > 1 {
			if m.hostCursor < 0 {
				m.hostCursor = len(m.entries[m.cursor].Hosts) - 1
			} else if m.hostCursor > 0 {
				m.hostCursor--
			}
		}
	case "right", "l":
		if n > 0 && len(m.entries[m.cursor].Hosts) > 1 {
			if m.hostCursor < 0 {
				m.hostCursor = 0
			} else if m.hostCursor < len(m.entries[m.cursor].Hosts)-1 {
				m.hostCursor++
			}
		}
	}
	return m, nil
}

func (m model) loadBackupsCmd() tea.Cmd {
	return func() tea.Msg {
		paths, err := m.manager.BackupList()
		if err != nil {
			return errMsg{err.Error()}
		}
		return backupsMsg{paths: paths}
	}
}

// --- background mutation commands ---

func doAdd(mgr *hosts.Manager, ip string, hosts []string) tea.Msg {
	if err := mgr.Add(ip, hosts); err != nil {
		return errMsg{err.Error()}
	}
	return infoMsg{fmt.Sprintf("Added %v under %s", hosts, ip)}
}

func doRemove(mgr *hosts.Manager, hosts []string, ip string) tea.Msg {
	if err := mgr.Remove(hosts, ip); err != nil {
		return errMsg{err.Error()}
	}
	if ip != "" {
		return infoMsg{fmt.Sprintf("Removed all hosts under %s", ip)}
	}
	return infoMsg{fmt.Sprintf("Removed %v", hosts)}
}

func doRestore(mgr *hosts.Manager, path string) tea.Msg {
	var err error
	if path == "" {
		err = mgr.RestoreLatest()
	} else {
		err = mgr.RestoreBackup(path)
	}
	if err != nil {
		return errMsg{err.Error()}
	}
	return infoMsg{"Restored from backup"}
}
