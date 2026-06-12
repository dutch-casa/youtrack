package tui

import (
	"context"
	"io"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dutchcaz/youtrack/internal/youtrack"
)

type Options struct {
	Query string
	Top   int
}

type Client interface {
	Issues(ctx context.Context, opts youtrack.IssueListOptions) ([]youtrack.Issue, error)
	Comments(ctx context.Context, issueID string) ([]youtrack.Comment, error)
	Attachments(ctx context.Context, opts youtrack.AttachmentListOptions) ([]youtrack.Attachment, error)
	Activities(ctx context.Context, opts youtrack.ActivityListOptions) ([]youtrack.Activity, error)
	IssueLinks(ctx context.Context, opts youtrack.IssueLinkListOptions) ([]youtrack.IssueLink, error)
	ApplyCommand(ctx context.Context, req youtrack.ApplyCommandRequest) (youtrack.CommandResult, error)
}

type pane int

const (
	detailsPane pane = iota
	commentsPane
	linksPane
	activitiesPane
	attachmentsPane
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
	status   string
	detail   viewport.Model

	commandMode    bool
	commandInput   textinput.Model
	commandRunning bool
	commandErr     error

	comments        map[string][]youtrack.Comment
	commentsLoading bool
	commentsErr     error

	attachments        map[string][]youtrack.Attachment
	attachmentsLoading bool
	attachmentsErr     error

	activities        map[string][]youtrack.Activity
	activitiesLoading bool
	activitiesErr     error

	links        map[string][]youtrack.IssueLink
	linksLoading bool
	linksErr     error
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

type attachmentsMsg struct {
	issueID     string
	attachments []youtrack.Attachment
	err         error
}

type activitiesMsg struct {
	issueID    string
	activities []youtrack.Activity
	err        error
}

type linksMsg struct {
	issueID string
	links   []youtrack.IssueLink
	err     error
}

type commandMsg struct {
	issueID string
	query   string
	err     error
}

func newModel(ctx context.Context, client Client, opts Options) model {
	if opts.Top <= 0 {
		opts.Top = 50
	}
	commandInput := textinput.New()
	commandInput.Prompt = ": "
	commandInput.Placeholder = "State Fixed"
	commandInput.CharLimit = 512
	detail := viewport.New(0, 0)
	return model{
		ctx:          ctx,
		client:       client,
		opts:         opts,
		loading:      true,
		detail:       detail,
		commandInput: commandInput,
		comments:     make(map[string][]youtrack.Comment),
		attachments:  make(map[string][]youtrack.Attachment),
		activities:   make(map[string][]youtrack.Activity),
		links:        make(map[string][]youtrack.IssueLink),
	}
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
		if m.commandMode {
			return m.updateCommandInput(msg)
		}
		if updated, ok := m.scrollDetail(msg); ok {
			return updated, nil
		}
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case ":":
			if m.currentIssueID() != "" && !m.commandRunning {
				m.commandMode = true
				m.commandInput.Reset()
				m.commandInput.Focus()
				m.commandErr = nil
				m.status = ""
				return m, textinput.Blink
			}
		case "tab":
			m.pane = m.pane.next()
			m.detail.GotoTop()
			return m.withSelectedPaneLoading()
		case "j", "down":
			if m.selected < len(m.issues)-1 {
				m.selected++
				m.detail.GotoTop()
				return m.withSelectedPaneLoading()
			}
		case "k", "up":
			if m.selected > 0 {
				m.selected--
				m.detail.GotoTop()
				return m.withSelectedPaneLoading()
			}
		case "g", "home":
			m.selected = 0
			m.detail.GotoTop()
			return m.withSelectedPaneLoading()
		case "G", "end":
			if len(m.issues) > 0 {
				m.selected = len(m.issues) - 1
				m.detail.GotoTop()
				return m.withSelectedPaneLoading()
			}
		case "r":
			m.loading = true
			m.err = nil
			m.detail.GotoTop()
			m.comments = make(map[string][]youtrack.Comment)
			m.commentsErr = nil
			m.commentsLoading = false
			m.attachments = make(map[string][]youtrack.Attachment)
			m.attachmentsErr = nil
			m.attachmentsLoading = false
			m.activities = make(map[string][]youtrack.Activity)
			m.activitiesErr = nil
			m.activitiesLoading = false
			m.links = make(map[string][]youtrack.IssueLink)
			m.linksErr = nil
			m.linksLoading = false
			m.status = ""
			m.commandErr = nil
			m.commandMode = false
			m.commandRunning = false
			m.commandInput.Blur()
			m.commandInput.Reset()
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
			m.detail.GotoTop()
		}
		if msg.err == nil {
			m.comments[msg.issueID] = msg.comments
		}
	case attachmentsMsg:
		if m.currentIssueID() == msg.issueID {
			m.attachmentsLoading = false
			m.attachmentsErr = msg.err
			m.detail.GotoTop()
		}
		if msg.err == nil {
			m.attachments[msg.issueID] = msg.attachments
		}
	case activitiesMsg:
		if m.currentIssueID() == msg.issueID {
			m.activitiesLoading = false
			m.activitiesErr = msg.err
			m.detail.GotoTop()
		}
		if msg.err == nil {
			m.activities[msg.issueID] = msg.activities
		}
	case linksMsg:
		if m.currentIssueID() == msg.issueID {
			m.linksLoading = false
			m.linksErr = msg.err
			m.detail.GotoTop()
		}
		if msg.err == nil {
			m.links[msg.issueID] = msg.links
		}
	case commandMsg:
		m.commandRunning = false
		m.commandErr = msg.err
		if msg.err != nil {
			m.status = ""
			return m, nil
		}
		m.status = "Applied " + msg.query + " to " + msg.issueID
		m.commandInput.Reset()
		delete(m.comments, msg.issueID)
		delete(m.attachments, msg.issueID)
		delete(m.activities, msg.issueID)
		delete(m.links, msg.issueID)
		return m, m.loadIssues
	}
	return m, nil
}

func (m model) scrollDetail(msg tea.KeyMsg) (model, bool) {
	switch {
	case msg.Type == tea.KeyPgDown || msg.String() == "pgdown":
		m.syncDetailViewport()
		m.detail.SetYOffset(m.detail.YOffset + max(1, m.detail.Height))
		return m, true
	case msg.Type == tea.KeyPgUp || msg.String() == "pgup":
		m.syncDetailViewport()
		m.detail.SetYOffset(m.detail.YOffset - max(1, m.detail.Height))
		return m, true
	case msg.Type == tea.KeyCtrlD || msg.String() == "ctrl+d":
		m.syncDetailViewport()
		m.detail.SetYOffset(m.detail.YOffset + max(1, m.detail.Height/2))
		return m, true
	case msg.Type == tea.KeyCtrlU || msg.String() == "ctrl+u":
		m.syncDetailViewport()
		m.detail.SetYOffset(m.detail.YOffset - max(1, m.detail.Height/2))
		return m, true
	default:
		return m, false
	}
}

func (m model) updateCommandInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.commandMode = false
		m.commandInput.Blur()
		m.commandInput.Reset()
		m.commandErr = nil
		return m, nil
	case "enter":
		if m.commandRunning {
			return m, nil
		}
		query := m.commandInput.Value()
		if query == "" {
			m.commandErr = nil
			m.commandMode = false
			m.commandInput.Blur()
			return m, nil
		}
		issueID := m.currentIssueID()
		if issueID == "" {
			m.commandMode = false
			m.commandInput.Blur()
			return m, nil
		}
		m.commandMode = false
		m.commandInput.Blur()
		m.commandRunning = true
		m.commandErr = nil
		m.status = "Applying " + query + "..."
		return m, m.applyCommand(issueID, query)
	}
	var cmd tea.Cmd
	m.commandInput, cmd = m.commandInput.Update(msg)
	return m, cmd
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

func (m model) withSelectedAttachmentsLoading() (tea.Model, tea.Cmd) {
	issueID := m.currentIssueID()
	if issueID == "" {
		return m, nil
	}
	if _, ok := m.attachments[issueID]; ok {
		m.attachmentsLoading = false
		m.attachmentsErr = nil
		return m, nil
	}
	m.attachmentsLoading = true
	m.attachmentsErr = nil
	return m, m.loadAttachments(issueID)
}

func (m model) withSelectedPaneLoading() (tea.Model, tea.Cmd) {
	switch m.pane {
	case commentsPane:
		return m.withSelectedCommentsLoading()
	case linksPane:
		return m.withSelectedLinksLoading()
	case activitiesPane:
		return m.withSelectedActivitiesLoading()
	case attachmentsPane:
		return m.withSelectedAttachmentsLoading()
	default:
		return m, nil
	}
}

func (m model) withSelectedActivitiesLoading() (tea.Model, tea.Cmd) {
	issueID := m.currentIssueID()
	if issueID == "" {
		return m, nil
	}
	if _, ok := m.activities[issueID]; ok {
		m.activitiesLoading = false
		m.activitiesErr = nil
		return m, nil
	}
	m.activitiesLoading = true
	m.activitiesErr = nil
	return m, m.loadActivities(issueID)
}

func (m model) withSelectedLinksLoading() (tea.Model, tea.Cmd) {
	issueID := m.currentIssueID()
	if issueID == "" {
		return m, nil
	}
	if _, ok := m.links[issueID]; ok {
		m.linksLoading = false
		m.linksErr = nil
		return m, nil
	}
	m.linksLoading = true
	m.linksErr = nil
	return m, m.loadLinks(issueID)
}

func (m model) loadComments(issueID string) tea.Cmd {
	return func() tea.Msg {
		comments, err := m.client.Comments(m.ctx, issueID)
		return commentsMsg{issueID: issueID, comments: comments, err: err}
	}
}

func (m model) loadAttachments(issueID string) tea.Cmd {
	return func() tea.Msg {
		attachments, err := m.client.Attachments(m.ctx, youtrack.AttachmentListOptions{IssueID: issueID, Top: 42})
		return attachmentsMsg{issueID: issueID, attachments: attachments, err: err}
	}
}

func (m model) loadActivities(issueID string) tea.Cmd {
	return func() tea.Msg {
		activities, err := m.client.Activities(m.ctx, youtrack.ActivityListOptions{
			IssueID:    issueID,
			Categories: youtrack.DefaultActivityCategories(),
			Top:        42,
			Reverse:    true,
		})
		return activitiesMsg{issueID: issueID, activities: activities, err: err}
	}
}

func (m model) loadLinks(issueID string) tea.Cmd {
	return func() tea.Msg {
		links, err := m.client.IssueLinks(m.ctx, youtrack.IssueLinkListOptions{IssueID: issueID, Top: 42})
		return linksMsg{issueID: issueID, links: links, err: err}
	}
}

func (m model) applyCommand(issueID, query string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.ApplyCommand(m.ctx, youtrack.ApplyCommandRequest{
			IssueID: issueID,
			Query:   query,
		})
		return commandMsg{issueID: issueID, query: query, err: err}
	}
}

func (m model) currentIssueID() string {
	if len(m.issues) == 0 || m.selected < 0 || m.selected >= len(m.issues) {
		return ""
	}
	return m.issues[m.selected].IDReadable
}

func (p pane) next() pane {
	switch p {
	case detailsPane:
		return commentsPane
	case commentsPane:
		return linksPane
	case linksPane:
		return activitiesPane
	case activitiesPane:
		return attachmentsPane
	default:
		return detailsPane
	}
}
