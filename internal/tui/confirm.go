package tui

import tea "github.com/charmbracelet/bubbletea"

func (m model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		action := m.confirmAction
		m.state = viewList
		m.confirmAction = nil
		m.confirmMsg = ""
		if action != nil {
			return m, func() tea.Msg { return action() }
		}
		return m, nil
	case "n", "N", "esc":
		m.state = viewList
		m.confirmAction = nil
		m.confirmMsg = ""
		return m, nil
	}
	return m, nil
}
