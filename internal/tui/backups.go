package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) handleBackupsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = viewList
		return m, nil
	case "up", "k":
		if m.bcursor > 0 {
			m.bcursor--
		}
	case "down", "j":
		if m.bcursor < len(m.backups)-1 {
			m.bcursor++
		}
	case "enter":
		if len(m.backups) == 0 {
			return m, nil
		}
		path := m.backups[m.bcursor]
		mgr := m.manager
		m.confirmMsg = fmt.Sprintf("Restore %s from %s?", m.manager.HostsPath, path)
		m.confirmAction = func() tea.Msg { return doRestore(mgr, path) }
		m.state = viewConfirm
		return m, nil
	}
	return m, nil
}
