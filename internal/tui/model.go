// Package tui implements the interactive Bubble Tea interface for vhoster.
package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/HriThik-MaNoj/vhoster/internal/hosts"
)

type viewState int

const (
	viewList viewState = iota
	viewAddForm
	viewConfirm
	viewBackups
)

// confirmFn is a deferred action executed when the user confirms a prompt.
type confirmFn func() tea.Msg

type model struct {
	manager *hosts.Manager
	state   viewState

	width, height int

	entries    []hosts.Entry
	cursor     int   // which entry is selected
	hostCursor int   // which host within that entry is selected (-1 = all)

	ipInput    textinput.Model
	hostsInput textinput.Model
	formField  int // 0 = ip, 1 = hosts
	formErr    string

	confirmMsg    string
	confirmAction confirmFn

	backups []string
	bcursor int

	err  string
	info string
}

// Run launches the TUI for the given manager.
func Run(m *hosts.Manager) error {
	p := tea.NewProgram(newModel(m), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func newModel(m *hosts.Manager) model {
	ip := textinput.New()
	ip.Placeholder = "e.g. 10.0.0.5"
	ip.CharLimit = 45
	ip.Focus()

	hs := textinput.New()
	hs.Placeholder = "api.local admin.local"
	hs.CharLimit = 200

	return model{
		manager:    m,
		state:      viewList,
		ipInput:    ip,
		hostsInput: hs,
		formField:  0,
	}
}

func (model) Init() tea.Cmd {
	return func() tea.Msg { return reloadMsg{} }
}

// reload re-reads managed entries from disk into the model.
func (m *model) reload() {
	entries, err := m.manager.List()
	if err != nil {
		m.err = err.Error()
		return
	}
	m.entries = entries
	if m.cursor > len(entries)-1 {
		m.cursor = max(0, len(entries)-1)
	}
}

// --- messages ---

type reloadMsg struct{}

type errMsg struct{ err string }
type infoMsg struct{ text string }
type backupsMsg struct{ paths []string }
