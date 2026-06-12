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
	fieldPanelStyle      = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("6")).Padding(0, 1)
	fieldLabelStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	fieldDefaultStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	fieldProjectStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	fieldPersonStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	fieldGoodStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	fieldWarnStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	fieldBadStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	fieldInfoStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	evidencePanelStyle   = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("12")).Padding(0, 1)
	evidenceCountStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	evidenceHintStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)
