package tui

import (
	"context"
	"encoding/json"
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
	Comments(ctx context.Context, issueID string) ([]youtrack.Comment, error)
}

type pane int

const (
	detailsPane pane = iota
	commentsPane
)

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
	pane     pane

	comments        map[string][]youtrack.Comment
	commentsLoading bool
	commentsErr     error
}

type issuesMsg struct {
	issues []youtrack.Issue
	err    error
}

type commentsMsg struct {
	issueID  string
	comments []youtrack.Comment
	err      error
}

func newModel(ctx context.Context, client Client, opts Options) model {
	if opts.Top <= 0 {
		opts.Top = 50
	}
	return model{ctx: ctx, client: client, opts: opts, loading: true, comments: make(map[string][]youtrack.Comment)}
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
		case "tab":
			if m.pane == detailsPane {
				m.pane = commentsPane
				return m.withSelectedCommentsLoading()
			}
			m.pane = detailsPane
		case "j", "down":
			if m.selected < len(m.issues)-1 {
				m.selected++
				if m.pane == commentsPane {
					return m.withSelectedCommentsLoading()
				}
			}
		case "k", "up":
			if m.selected > 0 {
				m.selected--
				if m.pane == commentsPane {
					return m.withSelectedCommentsLoading()
				}
			}
		case "g", "home":
			m.selected = 0
			if m.pane == commentsPane {
				return m.withSelectedCommentsLoading()
			}
		case "G", "end":
			if len(m.issues) > 0 {
				m.selected = len(m.issues) - 1
				if m.pane == commentsPane {
					return m.withSelectedCommentsLoading()
				}
			}
		case "r":
			m.loading = true
			m.err = nil
			m.comments = make(map[string][]youtrack.Comment)
			m.commentsErr = nil
			m.commentsLoading = false
			return m, m.loadIssues
		}
	case issuesMsg:
		m.loading = false
		m.err = msg.err
		m.issues = msg.issues
		if m.selected >= len(m.issues) {
			m.selected = max(0, len(m.issues)-1)
		}
	case commentsMsg:
		if m.currentIssueID() == msg.issueID {
			m.commentsLoading = false
			m.commentsErr = msg.err
		}
		if msg.err == nil {
			m.comments[msg.issueID] = msg.comments
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
	right := m.issuePane(detailWidth, bodyHeight)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + footer()
}

func (m model) loadIssues() tea.Msg {
	issues, err := m.client.Issues(m.ctx, youtrack.IssueListOptions{Query: m.opts.Query, Top: m.opts.Top})
	return issuesMsg{issues: issues, err: err}
}

func (m model) withSelectedCommentsLoading() (tea.Model, tea.Cmd) {
	issueID := m.currentIssueID()
	if issueID == "" {
		return m, nil
	}
	if _, ok := m.comments[issueID]; ok {
		m.commentsLoading = false
		m.commentsErr = nil
		return m, nil
	}
	m.commentsLoading = true
	m.commentsErr = nil
	return m, m.loadComments(issueID)
}

func (m model) loadComments(issueID string) tea.Cmd {
	return func() tea.Msg {
		comments, err := m.client.Comments(m.ctx, issueID)
		return commentsMsg{issueID: issueID, comments: comments, err: err}
	}
}

func (m model) currentIssueID() string {
	if len(m.issues) == 0 || m.selected < 0 || m.selected >= len(m.issues) {
		return ""
	}
	return m.issues[m.selected].IDReadable
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

func (m model) issuePane(width, height int) string {
	if len(m.issues) == 0 {
		return panelStyle.Width(width).Height(height).Render("No issues")
	}
	if m.pane == commentsPane {
		return m.issueComments(width, height)
	}
	return m.issueDetail(width, height)
}

func (m model) issueDetail(width, height int) string {
	issue := m.issues[m.selected]
	lines := []string{
		titleStyle.Render(issue.IDReadable),
		issue.Summary,
		"",
		"Project: " + issue.Project.ShortName,
		"Resolved: " + resolvedText(issue.Resolved),
		"",
		trimBlank(issue.Description),
	}
	fields := issueFields(issue)
	if len(fields) > 0 {
		lines = append(lines, "", titleStyle.Render("Fields"))
		lines = append(lines, fields...)
	}
	return panelStyle.Width(width).Height(height).Render(truncateBlock(strings.Join(lines, "\n"), width-4, height-2))
}

func (m model) issueComments(width, height int) string {
	issueID := m.currentIssueID()
	if m.commentsLoading {
		return panelStyle.Width(width).Height(height).Render("Loading comments...")
	}
	if m.commentsErr != nil {
		return panelStyle.Width(width).Height(height).Render("Error: " + m.commentsErr.Error())
	}
	comments := m.comments[issueID]
	if len(comments) == 0 {
		return panelStyle.Width(width).Height(height).Render(titleStyle.Render(issueID) + "\n\nNo comments")
	}

	lines := []string{titleStyle.Render(issueID), titleStyle.Render("Comments"), ""}
	for _, comment := range comments {
		author := firstNonEmpty(comment.Author.FullName, comment.Author.Name, comment.Author.Login)
		lines = append(lines, author+": "+strings.TrimSpace(comment.Text), "")
	}
	return panelStyle.Width(width).Height(height).Render(truncateBlock(strings.Join(lines, "\n"), width-4, height-2))
}

func footer() string {
	return helpStyle.Render("j/k move  tab details/comments  g/G top/bottom  r refresh  q quit")
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

func resolvedText(value any) string {
	if value == nil {
		return "no"
	}
	return "yes"
}

func issueFields(issue youtrack.Issue) []string {
	if len(issue.Custom) == 0 {
		return nil
	}
	var fields []struct {
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(issue.Custom, &fields); err != nil {
		return nil
	}

	lines := make([]string, 0, len(fields))
	for _, field := range fields {
		value := fieldValue(field.Value)
		if value == "" {
			continue
		}
		lines = append(lines, field.Name+": "+value)
	}
	return lines
}

func fieldValue(data json.RawMessage) string {
	if len(data) == 0 || string(data) == "null" {
		return ""
	}

	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		return text
	}

	var object struct {
		Presentation string `json:"presentation"`
		Name         string `json:"name"`
		Login        string `json:"login"`
		Text         string `json:"text"`
	}
	if err := json.Unmarshal(data, &object); err == nil {
		return firstNonEmpty(object.Presentation, object.Name, object.Login, object.Text)
	}

	var objects []struct {
		Presentation string `json:"presentation"`
		Name         string `json:"name"`
		Login        string `json:"login"`
		Text         string `json:"text"`
	}
	if err := json.Unmarshal(data, &objects); err == nil {
		values := make([]string, 0, len(objects))
		for _, object := range objects {
			value := firstNonEmpty(object.Presentation, object.Name, object.Login, object.Text)
			if value != "" {
				values = append(values, value)
			}
		}
		return strings.Join(values, ", ")
	}

	var primitive any
	if err := json.Unmarshal(data, &primitive); err == nil {
		return fmt.Sprint(primitive)
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

var (
	panelStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	titleStyle    = lipgloss.NewStyle().Bold(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("12"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)
