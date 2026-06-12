package tui

import "github.com/charmbracelet/lipgloss"

var (
	panelStyle           = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("8")).Padding(0, 1)
	titleStyle           = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	selectedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("12")).Bold(true)
	sectionStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Background(lipgloss.Color("0"))
	sectionSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("6")).Bold(true)
	helpStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	commandStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4"))
	statusStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)
