package tui

import "github.com/charmbracelet/lipgloss"

var (
	panelStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("12"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)
