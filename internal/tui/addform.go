package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/HriThik-MaNoj/vhoster/internal/validate"
)

func (m model) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = viewList
		m.formErr = ""
		return m, nil
	case "tab", "shift+tab":
		if m.formField == 0 {
			m.formField = 1
			m.ipInput.Blur()
			m.hostsInput.Focus()
		} else {
			m.formField = 0
			m.hostsInput.Blur()
			m.ipInput.Focus()
		}
		return m, textinput.Blink
	case "enter":
		return m.submitForm()
	}
	return m.forwardInputs(msg)
}

func (m model) submitForm() (tea.Model, tea.Cmd) {
	ip, hosts, errs := validate.ValidateAdd(m.ipInput.Value(), validate.ParseHostField(m.hostsInput.Value()))
	if len(errs) > 0 {
		m.formErr = strings.Join(errs, "; ")
		return m, nil
	}
	ipCanonical := ip.String()
	mgr := m.manager
	m.state = viewList
	m.formErr = ""
	return m, func() tea.Msg { return doAdd(mgr, ipCanonical, hosts) }
}

// forwardInputs passes a message to whichever textinput is focused. Used
// for typing keys and for the blink command.
func (m model) forwardInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.state != viewAddForm {
		return m, nil
	}
	var cmd tea.Cmd
	if m.formField == 0 {
		m.ipInput, cmd = m.ipInput.Update(msg)
	} else {
		m.hostsInput, cmd = m.hostsInput.Update(msg)
	}
	return m, cmd
}
