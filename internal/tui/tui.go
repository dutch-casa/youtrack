package tui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dutch-casa/youtrack/internal/youtrack"
)

type Options struct {
	Query   string
	Top     int
	Skip    int
	BaseURL string
	OpenURL func(string) error
}

type Client interface {
	Issues(ctx context.Context, opts youtrack.IssueListOptions) ([]youtrack.Issue, error)
	Projects(ctx context.Context, opts youtrack.PageOptions) ([]youtrack.Project, error)
	Users(ctx context.Context, opts youtrack.PageOptions) ([]youtrack.User, error)
	Articles(ctx context.Context, opts youtrack.ArticleListOptions) ([]youtrack.Article, error)
	Agiles(ctx context.Context, opts youtrack.PageOptions) ([]youtrack.Agile, error)
	HelpdeskProjects(ctx context.Context, opts youtrack.PageOptions) ([]youtrack.Project, error)
	Comments(ctx context.Context, issueID string) ([]youtrack.Comment, error)
	Attachments(ctx context.Context, opts youtrack.AttachmentListOptions) ([]youtrack.Attachment, error)
	Activities(ctx context.Context, opts youtrack.ActivityListOptions) ([]youtrack.Activity, error)
	IssueLinks(ctx context.Context, opts youtrack.IssueLinkListOptions) ([]youtrack.IssueLink, error)
	WorkItems(ctx context.Context, opts youtrack.WorkItemListOptions) ([]youtrack.WorkItem, error)
	AddWorkItem(ctx context.Context, req youtrack.AddWorkItemRequest) (youtrack.WorkItem, error)
	AddComment(ctx context.Context, issueID, text string) (youtrack.Comment, error)
	ApplyCommand(ctx context.Context, req youtrack.ApplyCommandRequest) (youtrack.CommandResult, error)
}

type pane int

const (
	detailsPane pane = iota
	commentsPane
	linksPane
	activitiesPane
	workItemsPane
	attachmentsPane
)

type inputMode int

const (
	modeNavigation inputMode = iota
	modeCommand
	modeComment
	modeWorkItem
	modeQuery
	modeProject
	modeIssue
)

func Run(ctx context.Context, client Client, opts Options, out io.Writer) error {
	model := newModel(ctx, client, opts)
	program := tea.NewProgram(model, tea.WithOutput(out), tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := program.Run()
	return err
}

type section int

const (
	sectionIssues section = iota
	sectionKnowledge
	sectionHelpdesk
	sectionAgile
	sectionProjects
	sectionUsers
)

var sections = []section{
	sectionIssues,
	sectionKnowledge,
	sectionHelpdesk,
	sectionAgile,
	sectionProjects,
	sectionUsers,
}

type resourceItem struct {
	ID           string
	Title        string
	Subtitle     string
	Body         string
	BodyMarkdown bool
}

type projectOption struct {
	ID       string
	Name     string
	Subtitle string
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
	section  section
	pane     pane
	status   string
	detail   viewport.Model
	help     help.Model

	resources         []resourceItem
	allResources      []resourceItem
	resourceSelected  int
	resourcesLoading  bool
	resourcesErr      error
	resourceFilter    string
	resourceListStart int

	inputMode      inputMode
	commandInput   textinput.Model
	commandRunning bool
	commandErr     error

	commentInput   textinput.Model
	commentRunning bool
	commentErr     error

	workItemInput   textinput.Model
	workItemRunning bool
	workItemErr     error

	queryInput    textinput.Model
	projectInput  textinput.Model
	issueInput    textinput.Model
	projectFilter string

	projectOptions        []projectOption
	allProjectOptions     []projectOption
	projectOptionSelected int
	projectOptionsLoading bool
	projectOptionsErr     error

	comments        map[string][]youtrack.Comment
	commentsLoading bool
	commentsErr     error

	attachments        map[string][]youtrack.Attachment
	attachmentsLoading bool
	attachmentsErr     error

	activities        map[string][]youtrack.Activity
	activitiesLoading bool
	activitiesErr     error

	workItems        map[string][]youtrack.WorkItem
	workItemsLoading bool
	workItemsErr     error

	links        map[string][]youtrack.IssueLink
	linksLoading bool
	linksErr     error
}

type issuesMsg struct {
	issues []youtrack.Issue
	err    error
}

type resourcesMsg struct {
	section   section
	resources []resourceItem
	err       error
}

type projectOptionsMsg struct {
	projects []projectOption
	err      error
}

type openURLMsg struct {
	url string
	err error
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

type workItemsMsg struct {
	issueID   string
	workItems []youtrack.WorkItem
	err       error
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

type addCommentMsg struct {
	issueID string
	text    string
	err     error
}

type addWorkItemMsg struct {
	issueID string
	draft   workItemDraft
	err     error
}

func newModel(ctx context.Context, client Client, opts Options) model {
	if opts.Top <= 0 {
		opts.Top = 50
	}
	commandInput := newPrompt(": ", "State Fixed", 512)
	commentInput := newPrompt("comment> ", "Add a quick comment", 2048)
	workItemInput := newPrompt("work> ", "45m implementation", 512)
	queryInput := newPrompt("/ ", "project: ABC #Unresolved", 512)
	projectInput := newPrompt("project> ", "ABC", 128)
	issueInput := newPrompt("issue> ", "ABC-123", 128)
	detail := viewport.New(0, 0)
	helpView := help.New()
	return model{
		ctx:           ctx,
		client:        client,
		opts:          opts,
		loading:       true,
		section:       sectionIssues,
		detail:        detail,
		help:          helpView,
		commandInput:  commandInput,
		commentInput:  commentInput,
		workItemInput: workItemInput,
		queryInput:    queryInput,
		projectInput:  projectInput,
		issueInput:    issueInput,
		comments:      make(map[string][]youtrack.Comment),
		attachments:   make(map[string][]youtrack.Attachment),
		activities:    make(map[string][]youtrack.Activity),
		workItems:     make(map[string][]youtrack.WorkItem),
		links:         make(map[string][]youtrack.IssueLink),
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
		m.help.Width = msg.Width
	case tea.MouseMsg:
		return m.updateMouse(msg)
	case tea.KeyMsg:
		switch m.inputMode {
		case modeCommand:
			return m.updateCommandInput(msg)
		case modeComment:
			return m.updateCommentInput(msg)
		case modeWorkItem:
			return m.updateWorkItemInput(msg)
		case modeQuery:
			return m.updateQueryInput(msg)
		case modeProject:
			return m.updateProjectInput(msg)
		case modeIssue:
			return m.updateIssueInput(msg)
		}
		if updated, ok := m.scrollDetail(msg); ok {
			return updated, nil
		}
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "1":
			return m.switchSection(sectionIssues)
		case "2":
			return m.switchSection(sectionKnowledge)
		case "3":
			return m.switchSection(sectionHelpdesk)
		case "4":
			return m.switchSection(sectionAgile)
		case "5":
			return m.switchSection(sectionProjects)
		case "6":
			return m.switchSection(sectionUsers)
		case ":":
			if m.section == sectionIssues && m.currentIssueID() != "" && !m.commandRunning {
				return m, m.openCommandPrompt()
			}
		case "c":
			if m.section == sectionIssues && m.currentIssueID() != "" && !m.commentRunning {
				return m, m.openCommentPrompt()
			}
		case "w":
			if m.section == sectionIssues && m.currentIssueID() != "" && !m.workItemRunning {
				return m, m.openWorkItemPrompt()
			}
		case "/":
			if m.section == sectionIssues && !m.loading {
				return m, m.openQueryPrompt()
			}
			if m.section != sectionIssues && !m.resourcesLoading {
				return m, m.openQueryPrompt()
			}
		case "P":
			if m.section == sectionIssues && !m.loading {
				return m, m.openProjectPrompt()
			}
		case "o":
			return m, m.openCurrentInBrowser
		case "i":
			if m.section == sectionIssues && !m.loading {
				return m, m.openIssuePrompt()
			}
		case "tab":
			if m.section != sectionIssues {
				return m, nil
			}
			m.pane = m.pane.next()
			m.detail.GotoTop()
			return m.withSelectedPaneLoading()
		case "j", "down":
			return m.moveSelection(1)
		case "k", "up":
			return m.moveSelection(-1)
		case "g", "home":
			m.selectFirst()
			m.detail.GotoTop()
			return m.withCurrentSelectionLoading()
		case "G", "end":
			m.selectLast()
			m.detail.GotoTop()
			return m.withCurrentSelectionLoading()
		case "n":
			if m.section == sectionIssues && !m.loading && m.canLoadNextIssuePage() {
				m.opts.Skip += m.opts.Top
				m.selected = 0
				return m.withIssueListLoading("Loading next page...")
			}
		case "p":
			if m.section == sectionIssues && !m.loading && m.opts.Skip > 0 {
				m.opts.Skip = max(0, m.opts.Skip-m.opts.Top)
				m.selected = 0
				return m.withIssueListLoading("Loading previous page...")
			}
		case "r":
			m.clearActionErrors()
			m.commandRunning = false
			m.commentRunning = false
			m.workItemRunning = false
			m.clearPrompts()
			return m.withSectionLoading("")
		}
	case issuesMsg:
		m.loading = false
		m.err = msg.err
		m.issues = msg.issues
		if m.selected >= len(m.issues) {
			m.selected = max(0, len(m.issues)-1)
		}
	case resourcesMsg:
		if msg.section != m.section {
			return m, nil
		}
		m.resourcesLoading = false
		m.resourcesErr = msg.err
		m.allResources = msg.resources
		m.resources = filterResources(msg.resources, m.resourceFilter)
		if m.resourceSelected >= len(m.resources) {
			m.resourceSelected = max(0, len(m.resources)-1)
		}
		m.detail.GotoTop()
	case projectOptionsMsg:
		m.projectOptionsLoading = false
		m.projectOptionsErr = msg.err
		m.allProjectOptions = msg.projects
		m.projectOptions = filterProjectOptions(msg.projects, m.projectInput.Value())
		if m.projectOptionSelected >= len(m.projectOptions) {
			m.projectOptionSelected = max(0, len(m.projectOptions)-1)
		}
	case openURLMsg:
		if msg.err != nil {
			m.status = "Open failed: " + msg.err.Error()
			return m, nil
		}
		m.status = "Opened " + msg.url
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
	case workItemsMsg:
		if m.currentIssueID() == msg.issueID {
			m.workItemsLoading = false
			m.workItemsErr = msg.err
			m.detail.GotoTop()
		}
		if msg.err == nil {
			m.workItems[msg.issueID] = msg.workItems
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
		clearInput(&m.commandInput)
		delete(m.comments, msg.issueID)
		delete(m.attachments, msg.issueID)
		delete(m.activities, msg.issueID)
		delete(m.links, msg.issueID)
		return m, m.loadIssues
	case addCommentMsg:
		m.commentRunning = false
		m.commentErr = msg.err
		if msg.err != nil {
			m.status = ""
			return m, nil
		}
		m.status = "Commented on " + msg.issueID
		clearInput(&m.commentInput)
		delete(m.comments, msg.issueID)
		delete(m.activities, msg.issueID)
		if m.currentIssueID() != msg.issueID {
			return m, nil
		}
		m.pane = commentsPane
		m.detail.GotoTop()
		return m.withSelectedCommentsLoading()
	case addWorkItemMsg:
		m.workItemRunning = false
		m.workItemErr = msg.err
		if msg.err != nil {
			m.status = ""
			return m, nil
		}
		m.status = "Added " + formatMinutes(msg.draft.Minutes) + " to " + msg.issueID
		clearInput(&m.workItemInput)
		delete(m.workItems, msg.issueID)
		delete(m.activities, msg.issueID)
		if m.currentIssueID() != msg.issueID {
			return m, nil
		}
		m.pane = workItemsPane
		m.detail.GotoTop()
		return m.withSelectedWorkItemsLoading()
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
		m.closeCommandPrompt()
		m.commandErr = nil
		return m, nil
	case "enter":
		if m.commandRunning {
			return m, nil
		}
		query := m.commandInput.Value()
		if query == "" {
			m.commandErr = nil
			m.closeCommandPrompt()
			return m, nil
		}
		issueID := m.currentIssueID()
		if issueID == "" {
			m.closeCommandPrompt()
			return m, nil
		}
		m.closeCommandPrompt()
		m.commandRunning = true
		m.commandErr = nil
		m.status = "Applying " + query + "..."
		return m, m.applyCommand(issueID, query)
	}
	var cmd tea.Cmd
	m.commandInput, cmd = m.commandInput.Update(msg)
	return m, cmd
}

func (m model) updateCommentInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.closeCommentPrompt()
		m.commentErr = nil
		return m, nil
	case "enter":
		if m.commentRunning {
			return m, nil
		}
		text := strings.TrimSpace(m.commentInput.Value())
		if text == "" {
			m.commentErr = nil
			m.closeCommentPrompt()
			return m, nil
		}
		issueID := m.currentIssueID()
		if issueID == "" {
			m.closeCommentPrompt()
			return m, nil
		}
		m.closeCommentPrompt()
		m.commentRunning = true
		m.commentErr = nil
		m.status = "Adding comment..."
		return m, m.addComment(issueID, text)
	}
	var cmd tea.Cmd
	m.commentInput, cmd = m.commentInput.Update(msg)
	return m, cmd
}

func (m model) updateWorkItemInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.closeWorkItemPrompt()
		m.workItemErr = nil
		return m, nil
	case "enter":
		if m.workItemRunning {
			return m, nil
		}
		draft, err := parseWorkItemInput(m.workItemInput.Value())
		if err != nil {
			m.workItemErr = err
			return m, nil
		}
		issueID := m.currentIssueID()
		if issueID == "" {
			m.closeWorkItemPrompt()
			return m, nil
		}
		m.closeWorkItemPrompt()
		m.workItemRunning = true
		m.workItemErr = nil
		m.status = "Adding work item..."
		return m, m.addWorkItem(issueID, draft)
	}
	var cmd tea.Cmd
	m.workItemInput, cmd = m.workItemInput.Update(msg)
	return m, cmd
}

func (m model) updateQueryInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.closeQueryPrompt()
		return m, nil
	case "enter":
		query := strings.TrimSpace(m.queryInput.Value())
		m.closeQueryPrompt()
		if m.section != sectionIssues {
			m.resourceFilter = query
			m.resources = filterResources(m.allResources, query)
			m.resourceSelected = 0
			m.detail.GotoTop()
			if query == "" {
				m.status = ""
			} else {
				m.status = fmt.Sprintf("Filtered %s to %d result(s)", m.section.title(), len(m.resources))
			}
			return m, nil
		}
		m.opts.Query = query
		m.opts.Skip = 0
		m.selected = 0
		m.clearActionErrors()
		return m.withIssueListLoading("Loading query...")
	}
	var cmd tea.Cmd
	m.queryInput, cmd = m.queryInput.Update(msg)
	if m.section != sectionIssues {
		m.resourceFilter = strings.TrimSpace(m.queryInput.Value())
		m.resources = filterResources(m.allResources, m.resourceFilter)
		m.resourceSelected = min(m.resourceSelected, max(0, len(m.resources)-1))
		m.detail.GotoTop()
	}
	return m, cmd
}

func (m model) updateProjectInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.closeProjectPrompt()
		return m, nil
	case "up", "k":
		if m.projectOptionSelected > 0 {
			m.projectOptionSelected--
		}
		return m, nil
	case "down", "j":
		if m.projectOptionSelected < len(m.projectOptions)-1 {
			m.projectOptionSelected++
		}
		return m, nil
	case "g", "home":
		m.projectOptionSelected = 0
		return m, nil
	case "G", "end":
		if len(m.projectOptions) > 0 {
			m.projectOptionSelected = len(m.projectOptions) - 1
		}
		return m, nil
	case "enter":
		project := m.selectedProjectFilter()
		m.closeProjectPrompt()
		m.projectFilter = project
		m.opts.Skip = 0
		m.selected = 0
		m.clearActionErrors()
		if project == "" {
			return m.withIssueListLoading("Loading issues...")
		}
		return m.withIssueListLoading("Loading project " + project + "...")
	}
	var cmd tea.Cmd
	m.projectInput, cmd = m.projectInput.Update(msg)
	m.projectOptions = filterProjectOptions(m.allProjectOptions, m.projectInput.Value())
	m.projectOptionSelected = min(m.projectOptionSelected, max(0, len(m.projectOptions)-1))
	return m, cmd
}

func (m model) selectedProjectFilter() string {
	if len(m.projectOptions) == 0 {
		return strings.TrimSpace(m.projectInput.Value())
	}
	selected := min(max(m.projectOptionSelected, 0), len(m.projectOptions)-1)
	return m.projectOptions[selected].ID
}

func (m model) loadProjectOptions() tea.Msg {
	projects, err := m.client.Projects(m.ctx, youtrack.PageOptions{Top: 200})
	if err != nil {
		return projectOptionsMsg{err: err}
	}
	return projectOptionsMsg{projects: projectOptions(projects)}
}

func (m model) updateIssueInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.closeIssuePrompt()
		return m, nil
	case "enter":
		issueID := strings.TrimSpace(m.issueInput.Value())
		m.closeIssuePrompt()
		if issueID == "" {
			return m, nil
		}
		m.projectFilter = ""
		m.opts.Query = issueID
		m.opts.Skip = 0
		m.selected = 0
		m.clearActionErrors()
		return m.withIssueListLoading("Opening " + issueID + "...")
	}
	var cmd tea.Cmd
	m.issueInput, cmd = m.issueInput.Update(msg)
	return m, cmd
}

func (m model) loadIssues() tea.Msg {
	issues, err := m.client.Issues(m.ctx, youtrack.IssueListOptions{Query: m.issueSearchQuery(), Top: m.opts.Top, Skip: m.opts.Skip})
	return issuesMsg{issues: issues, err: err}
}

func (m model) openCurrentInBrowser() tea.Msg {
	rawURL, err := m.currentBrowserURL()
	if err != nil {
		return openURLMsg{err: err}
	}
	if m.opts.OpenURL == nil {
		return openURLMsg{url: rawURL}
	}
	if err := m.opts.OpenURL(rawURL); err != nil {
		return openURLMsg{url: rawURL, err: err}
	}
	return openURLMsg{url: rawURL}
}

func (m model) issueSearchQuery() string {
	query := strings.TrimSpace(m.opts.Query)
	project := strings.TrimSpace(m.projectFilter)
	if project == "" {
		return query
	}
	if query == "" {
		return "project: " + project
	}
	return "project: " + project + " " + query
}

func (m model) loadResources(s section) tea.Cmd {
	return func() tea.Msg {
		resources, err := m.resourcesForSection(s)
		return resourcesMsg{section: s, resources: resources, err: err}
	}
}

func (m model) resourcesForSection(s section) ([]resourceItem, error) {
	switch s {
	case sectionKnowledge:
		articles, err := m.client.Articles(m.ctx, youtrack.ArticleListOptions{Top: m.opts.Top})
		if err != nil {
			return nil, err
		}
		return articleResources(articles), nil
	case sectionHelpdesk:
		projects, err := m.client.HelpdeskProjects(m.ctx, youtrack.PageOptions{Top: m.opts.Top})
		if err != nil {
			return nil, err
		}
		return projectResources(projects), nil
	case sectionAgile:
		agiles, err := m.client.Agiles(m.ctx, youtrack.PageOptions{Top: m.opts.Top})
		if err != nil {
			return nil, err
		}
		return agileResources(agiles), nil
	case sectionProjects:
		projects, err := m.client.Projects(m.ctx, youtrack.PageOptions{Top: m.opts.Top})
		if err != nil {
			return nil, err
		}
		return projectResources(projects), nil
	case sectionUsers:
		users, err := m.client.Users(m.ctx, youtrack.PageOptions{Top: m.opts.Top})
		if err != nil {
			return nil, err
		}
		return userResources(users), nil
	default:
		return nil, nil
	}
}

func (m model) switchSection(s section) (tea.Model, tea.Cmd) {
	if m.section == s {
		return m, nil
	}
	m.section = s
	m.detail.GotoTop()
	m.clearActionErrors()
	m.clearPrompts()
	if s == sectionIssues {
		return m.withIssueListLoading("Loading issues...")
	}
	m.resources = nil
	m.allResources = nil
	m.resourceSelected = 0
	m.resourcesLoading = true
	m.resourcesErr = nil
	m.resourceFilter = ""
	m.status = "Loading " + s.title() + "..."
	return m, m.loadResources(s)
}

func (m model) withSectionLoading(status string) (tea.Model, tea.Cmd) {
	if m.section == sectionIssues {
		return m.withIssueListLoading(status)
	}
	m.resourcesLoading = true
	m.resourcesErr = nil
	m.resourceSelected = 0
	m.detail.GotoTop()
	m.status = status
	return m, m.loadResources(m.section)
}

func (m model) moveSelection(delta int) (tea.Model, tea.Cmd) {
	if m.section == sectionIssues {
		if len(m.issues) == 0 {
			return m, nil
		}
		next := min(max(m.selected+delta, 0), len(m.issues)-1)
		if next == m.selected {
			return m, nil
		}
		m.selected = next
		m.detail.GotoTop()
		return m.withSelectedPaneLoading()
	}
	if len(m.resources) == 0 {
		return m, nil
	}
	m.resourceSelected = min(max(m.resourceSelected+delta, 0), len(m.resources)-1)
	m.detail.GotoTop()
	return m, nil
}

func (m *model) selectFirst() {
	if m.section == sectionIssues {
		m.selected = 0
		return
	}
	m.resourceSelected = 0
}

func (m *model) selectLast() {
	if m.section == sectionIssues {
		if len(m.issues) > 0 {
			m.selected = len(m.issues) - 1
		}
		return
	}
	if len(m.resources) > 0 {
		m.resourceSelected = len(m.resources) - 1
	}
}

func (m model) withCurrentSelectionLoading() (tea.Model, tea.Cmd) {
	if m.section == sectionIssues {
		return m.withSelectedPaneLoading()
	}
	return m, nil
}

func (m model) updateMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	event := tea.MouseEvent(msg)
	if event.Action != tea.MouseActionPress {
		return m, nil
	}
	if m.inputMode == modeProject {
		return m.updateProjectMouse(event)
	}
	if event.Button == tea.MouseButtonWheelDown {
		if m.mouseInDetailPane(event) {
			return m.scrollDetailBy(1), nil
		}
		return m.moveSelection(1)
	}
	if event.Button == tea.MouseButtonWheelUp {
		if m.mouseInDetailPane(event) {
			return m.scrollDetailBy(-1), nil
		}
		return m.moveSelection(-1)
	}
	if event.Button != tea.MouseButtonLeft {
		return m, nil
	}
	if event.Y == 0 {
		if clicked, ok := sectionAtX(event.X); ok {
			return m.switchSection(clicked)
		}
	}
	if event.Y < 2 || event.Y >= max(3, m.height-3)+1 {
		return m, nil
	}
	listWidth := max(28, m.width/3)
	if event.X >= listWidth {
		return m, nil
	}
	row := event.Y - 3
	if row < 0 {
		return m, nil
	}
	if m.section == sectionIssues {
		start, end := visibleIssueRange(m.selected, len(m.issues), max(3, m.height-4))
		index := start + row
		if index >= start && index < end {
			m.selected = index
			m.detail.GotoTop()
			return m.withSelectedPaneLoading()
		}
		return m, nil
	}
	start, end := visibleResourceRange(m.resourceSelected, len(m.resources), max(3, m.height-4))
	index := start + row
	if index >= start && index < end {
		m.resourceSelected = index
		m.detail.GotoTop()
	}
	return m, nil
}

func (m model) updateProjectMouse(event tea.MouseEvent) (tea.Model, tea.Cmd) {
	switch event.Button {
	case tea.MouseButtonWheelDown:
		if m.projectOptionSelected < len(m.projectOptions)-1 {
			m.projectOptionSelected++
		}
		return m, nil
	case tea.MouseButtonWheelUp:
		if m.projectOptionSelected > 0 {
			m.projectOptionSelected--
		}
		return m, nil
	case tea.MouseButtonLeft:
		row := event.Y - 4
		if row < 0 {
			return m, nil
		}
		start, end := visibleResourceRange(m.projectOptionSelected, len(m.projectOptions), max(3, m.height-7))
		index := start + row
		if index < start || index >= end {
			return m, nil
		}
		m.projectOptionSelected = index
		project := m.selectedProjectFilter()
		m.closeProjectPrompt()
		m.projectFilter = project
		m.opts.Skip = 0
		m.selected = 0
		if project == "" {
			return m.withIssueListLoading("Loading issues...")
		}
		return m.withIssueListLoading("Loading project " + project + "...")
	default:
		return m, nil
	}
}

func (m model) mouseInDetailPane(event tea.MouseEvent) bool {
	listWidth := max(28, m.width/3)
	return event.X >= listWidth
}

func (m model) scrollDetailBy(direction int) model {
	m.syncDetailViewport()
	delta := max(1, m.detail.Height/3)
	m.detail.SetYOffset(m.detail.YOffset + direction*delta)
	return m
}

func (m model) withIssueListLoading(status string) (tea.Model, tea.Cmd) {
	m.loading = true
	m.err = nil
	m.section = sectionIssues
	m.detail.GotoTop()
	m.clearPaneCaches()
	m.status = status
	return m, m.loadIssues
}

func (m model) canLoadNextIssuePage() bool {
	return m.opts.Top > 0 && len(m.issues) == m.opts.Top
}

func (m *model) clearPaneCaches() {
	m.comments = make(map[string][]youtrack.Comment)
	m.commentsErr = nil
	m.commentsLoading = false
	m.attachments = make(map[string][]youtrack.Attachment)
	m.attachmentsErr = nil
	m.attachmentsLoading = false
	m.activities = make(map[string][]youtrack.Activity)
	m.activitiesErr = nil
	m.activitiesLoading = false
	m.workItems = make(map[string][]youtrack.WorkItem)
	m.workItemsErr = nil
	m.workItemsLoading = false
	m.links = make(map[string][]youtrack.IssueLink)
	m.linksErr = nil
	m.linksLoading = false
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
	case workItemsPane:
		return m.withSelectedWorkItemsLoading()
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

func (m model) withSelectedWorkItemsLoading() (tea.Model, tea.Cmd) {
	issueID := m.currentIssueID()
	if issueID == "" {
		return m, nil
	}
	if _, ok := m.workItems[issueID]; ok {
		m.workItemsLoading = false
		m.workItemsErr = nil
		return m, nil
	}
	m.workItemsLoading = true
	m.workItemsErr = nil
	return m, m.loadWorkItems(issueID)
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

func (m model) loadWorkItems(issueID string) tea.Cmd {
	return func() tea.Msg {
		workItems, err := m.client.WorkItems(m.ctx, youtrack.WorkItemListOptions{IssueID: issueID, Top: 42})
		return workItemsMsg{issueID: issueID, workItems: workItems, err: err}
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

func (m model) addComment(issueID, text string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.AddComment(m.ctx, issueID, text)
		return addCommentMsg{issueID: issueID, text: text, err: err}
	}
}

func (m model) addWorkItem(issueID string, draft workItemDraft) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.AddWorkItem(m.ctx, youtrack.AddWorkItemRequest{
			IssueID: issueID,
			Minutes: draft.Minutes,
			Text:    draft.Text,
		})
		return addWorkItemMsg{issueID: issueID, draft: draft, err: err}
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
		return workItemsPane
	case workItemsPane:
		return attachmentsPane
	default:
		return detailsPane
	}
}
