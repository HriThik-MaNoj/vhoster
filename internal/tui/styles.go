package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Padding(0, 1)
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("57"))
	entryStyle    = lipgloss.NewStyle().Padding(0, 1)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	okStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	hintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	boxStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
)
