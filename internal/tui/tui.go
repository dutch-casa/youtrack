package tui

import (
	"context"
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dutchcaz/youtrack/internal/youtrack"
)

type Options struct {
	Query string
	Top   int
}

type Client interface {
	Issues(ctx context.Context, opts youtrack.IssueListOptions) ([]youtrack.Issue, error)
}

func Run(ctx context.Context, client Client, opts Options, out io.Writer) error {
	model := newModel(ctx, client, opts)
	program := tea.NewProgram(model, tea.WithOutput(out), tea.WithAltScreen())
	_, err := program.Run()
	return err
}

type model struct {
	ctx      context.Context
	client   Client
	opts     Options
	issues   []youtrack.Issue
	selected int
	width    int
	height   int
	loading  bool
	err      error
}

type issuesMsg struct {
	issues []youtrack.Issue
	err    error
}

func newModel(ctx context.Context, client Client, opts Options) model {
	if opts.Top <= 0 {
		opts.Top = 50
	}
	return model{ctx: ctx, client: client, opts: opts, loading: true}
}

func (m model) Init() tea.Cmd {
	return m.loadIssues
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "j", "down":
			if m.selected < len(m.issues)-1 {
				m.selected++
			}
		case "k", "up":
			if m.selected > 0 {
				m.selected--
			}
		case "g", "home":
			m.selected = 0
		case "G", "end":
			if len(m.issues) > 0 {
				m.selected = len(m.issues) - 1
			}
		case "r":
			m.loading = true
			m.err = nil
			return m, m.loadIssues
		}
	case issuesMsg:
		m.loading = false
		m.err = msg.err
		m.issues = msg.issues
		if m.selected >= len(m.issues) {
			m.selected = max(0, len(m.issues)-1)
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		m.width = 100
	}
	if m.height == 0 {
		m.height = 30
	}

	bodyHeight := max(3, m.height-3)
	if m.loading {
		return panelStyle.Width(m.width).Height(bodyHeight).Render("Loading issues...") + "\n" + footer()
	}
	if m.err != nil {
		return panelStyle.Width(m.width).Height(bodyHeight).Render("Error: "+m.err.Error()) + "\n" + footer()
	}

	listWidth := max(28, m.width/3)
	detailWidth := max(40, m.width-listWidth-4)
	left := m.issueList(listWidth, bodyHeight)
	right := m.issueDetail(detailWidth, bodyHeight)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + footer()
}

func (m model) loadIssues() tea.Msg {
	issues, err := m.client.Issues(m.ctx, youtrack.IssueListOptions{Query: m.opts.Query, Top: m.opts.Top})
	return issuesMsg{issues: issues, err: err}
}

func (m model) issueList(width, height int) string {
	rows := make([]string, 0, min(len(m.issues), height))
	title := titleStyle.Render("Issues")
	rows = append(rows, title)
	for i, issue := range m.issues {
		line := fmt.Sprintf("%-12s %s", issue.IDReadable, issue.Summary)
		line = truncate(line, width-4)
		if i == m.selected {
			line = selectedStyle.Render(line)
		}
		rows = append(rows, line)
	}
	return panelStyle.Width(width).Height(height).Render(strings.Join(rows, "\n"))
}

func (m model) issueDetail(width, height int) string {
	if len(m.issues) == 0 {
		return panelStyle.Width(width).Height(height).Render("No issues")
	}
	issue := m.issues[m.selected]
	lines := []string{
		titleStyle.Render(issue.IDReadable),
		issue.Summary,
		"",
		"Project: " + issue.Project.ShortName,
		"",
		trimBlank(issue.Description),
	}
	return panelStyle.Width(width).Height(height).Render(truncateBlock(strings.Join(lines, "\n"), width-4, height-2))
}

func footer() string {
	return helpStyle.Render("j/k move  g/G top/bottom  r refresh  q quit")
}

func truncate(value string, width int) string {
	if width <= 0 || len(value) <= width {
		return value
	}
	if width <= 1 {
		return value[:width]
	}
	if width <= 3 {
		return value[:width]
	}
	return value[:width-3] + "..."
}

func truncateBlock(value string, width, height int) string {
	lines := strings.Split(value, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i := range lines {
		lines[i] = truncate(lines[i], width)
	}
	return strings.Join(lines, "\n")
}

func trimBlank(value string) string {
	if strings.TrimSpace(value) == "" {
		return "No description"
	}
	return value
}

var (
	panelStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	titleStyle    = lipgloss.NewStyle().Bold(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("12"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)
