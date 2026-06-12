package tui

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dutchcaz/youtrack/internal/youtrack"
)

type fakeClient struct {
	issues        []youtrack.Issue
	issueRequests *[]youtrack.IssueListOptions
	comments      []youtrack.Comment
	attachments   []youtrack.Attachment
	activities    []youtrack.Activity
	links         []youtrack.IssueLink
	commands      *[]youtrack.ApplyCommandRequest
	err           error
}

func (f fakeClient) Issues(ctx context.Context, opts youtrack.IssueListOptions) ([]youtrack.Issue, error) {
	if f.issueRequests != nil {
		*f.issueRequests = append(*f.issueRequests, opts)
	}
	return f.issues, f.err
}

func (f fakeClient) Comments(ctx context.Context, issueID string) ([]youtrack.Comment, error) {
	return f.comments, f.err
}

func (f fakeClient) Attachments(ctx context.Context, opts youtrack.AttachmentListOptions) ([]youtrack.Attachment, error) {
	return f.attachments, f.err
}

func (f fakeClient) Activities(ctx context.Context, opts youtrack.ActivityListOptions) ([]youtrack.Activity, error) {
	return f.activities, f.err
}

func (f fakeClient) IssueLinks(ctx context.Context, opts youtrack.IssueLinkListOptions) ([]youtrack.IssueLink, error) {
	return f.links, f.err
}

func (f fakeClient) ApplyCommand(ctx context.Context, req youtrack.ApplyCommandRequest) (youtrack.CommandResult, error) {
	if f.commands != nil {
		*f.commands = append(*f.commands, req)
	}
	return youtrack.CommandResult{Query: req.Query}, f.err
}

func TestModelMovementClamps(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{
		{IDReadable: "ABC-1", Summary: "One"},
		{IDReadable: "ABC-2", Summary: "Two"},
	}})
	m = updated.(model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	if m.selected != 1 {
		t.Fatalf("selected = %d, want 1", m.selected)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(model)
	if m.selected != 0 {
		t.Fatalf("selected = %d, want 0", m.selected)
	}
}

func TestModelRefreshSetsLoading(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updated.(model)
	if !m.loading {
		t.Fatal("loading = false, want true")
	}
	if cmd == nil {
		t.Fatal("refresh command = nil")
	}
}

func TestTabLoadsCommentsPane(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(model)
	if m.pane != commentsPane {
		t.Fatalf("pane = %v, want commentsPane", m.pane)
	}
	if !m.commentsLoading {
		t.Fatal("commentsLoading = false, want true")
	}
	if cmd == nil {
		t.Fatal("comment load command = nil")
	}
}

func TestTabCyclesToAttachmentsPane(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(model)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(model)
	if m.pane != attachmentsPane {
		t.Fatalf("pane = %v, want attachmentsPane", m.pane)
	}
	if !m.attachmentsLoading {
		t.Fatal("attachmentsLoading = false, want true")
	}
	if cmd == nil {
		t.Fatal("attachment load command = nil")
	}
}

func TestCommentsPaneRendersComments(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)
	m.pane = commentsPane
	updated, _ = m.Update(commentsMsg{
		issueID: "ABC-1",
		comments: []youtrack.Comment{{
			Text:   "Looks fixed",
			Author: youtrack.User{Login: "jane"},
		}},
	})
	m = updated.(model)

	view := m.View()
	if !strings.Contains(view, "Comments") || !strings.Contains(view, "jane: Looks fixed") {
		t.Fatalf("View() = %q, want comment", view)
	}
}

func TestAttachmentsPaneRendersAttachments(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)
	m.pane = attachmentsPane
	updated, _ = m.Update(attachmentsMsg{
		issueID: "ABC-1",
		attachments: []youtrack.Attachment{{
			Name:     "screenshot.png",
			Size:     2048,
			MimeType: "image/png",
			Author:   youtrack.User{Login: "jane"},
		}},
	})
	m = updated.(model)

	view := m.View()
	if !strings.Contains(view, "Attachments") || !strings.Contains(view, "screenshot.png") {
		t.Fatalf("View() = %q, want attachment", view)
	}
	if !strings.Contains(view, "2.0 KiB") || !strings.Contains(view, "image/png") {
		t.Fatalf("View() = %q, want attachment metadata", view)
	}
}

func TestLinksPaneRendersLinks(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)
	m.pane = linksPane
	updated, _ = m.Update(linksMsg{
		issueID: "ABC-1",
		links: []youtrack.IssueLink{{
			Direction: "OUTWARD",
			LinkType:  youtrack.LinkType{Name: "Depend", SourceToTarget: "is required for"},
			Issues:    []youtrack.Issue{{IDReadable: "ABC-2", Summary: "Two"}},
		}},
	})
	m = updated.(model)

	view := m.View()
	if !strings.Contains(view, "Links") || !strings.Contains(view, "ABC-2") || !strings.Contains(view, "is required for") {
		t.Fatalf("View() = %q, want link", view)
	}
}

func TestActivitiesPaneRendersActivity(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)
	m.pane = activitiesPane
	updated, _ = m.Update(activitiesMsg{
		issueID: "ABC-1",
		activities: []youtrack.Activity{{
			Type:   "CommentActivityItem",
			Author: youtrack.User{Login: "jane"},
			Target: []byte(`{"text":"Looks fixed"}`),
		}},
	})
	m = updated.(model)

	view := m.View()
	if !strings.Contains(view, "Activity") || !strings.Contains(view, "jane") || !strings.Contains(view, "Looks fixed") {
		t.Fatalf("View() = %q, want activity", view)
	}
}

func TestCommandModeAppliesYouTrackCommand(t *testing.T) {
	var commands []youtrack.ApplyCommandRequest
	m := newModel(context.Background(), fakeClient{commands: &commands}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	m = updated.(model)
	if m.inputMode != modeCommand {
		t.Fatalf("inputMode = %v, want modeCommand", m.inputMode)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("State Fixed")})
	m = updated.(model)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if !m.commandRunning {
		t.Fatal("commandRunning = false, want true")
	}
	if cmd == nil {
		t.Fatal("apply command = nil")
	}

	msg := cmd().(commandMsg)
	if msg.issueID != "ABC-1" || msg.query != "State Fixed" {
		t.Fatalf("commandMsg = %#v", msg)
	}
	if len(commands) != 1 || commands[0].IssueID != "ABC-1" || commands[0].Query != "State Fixed" {
		t.Fatalf("commands = %#v", commands)
	}

	updated, reload := m.Update(msg)
	m = updated.(model)
	if m.commandRunning {
		t.Fatal("commandRunning = true, want false")
	}
	if !strings.Contains(m.status, "Applied State Fixed") {
		t.Fatalf("status = %q, want applied status", m.status)
	}
	if reload == nil {
		t.Fatal("reload command = nil")
	}
}

func TestCommandModeCancels(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("State Fixed")})
	m = updated.(model)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)

	if m.inputMode != modeNavigation || m.commandInput.Value() != "" {
		t.Fatalf("inputMode = %v, command input = %q; want canceled", m.inputMode, m.commandInput.Value())
	}
	if cmd != nil {
		t.Fatal("cancel command != nil")
	}
}

func TestCommandModeSupportsCursorEditing(t *testing.T) {
	var commands []youtrack.ApplyCommandRequest
	m := newModel(context.Background(), fakeClient{commands: &commands}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("State Fxed")})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = updated.(model)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if cmd == nil {
		t.Fatal("apply command = nil")
	}

	msg := cmd().(commandMsg)
	if msg.query != "State Fixed" {
		t.Fatalf("query = %q, want cursor-edited command", msg.query)
	}
}

func TestInputModeIsExclusive(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	m = updated.(model)
	if m.inputMode != modeCommand {
		t.Fatalf("inputMode = %v, want modeCommand", m.inputMode)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(model)
	if m.inputMode != modeCommand {
		t.Fatalf("inputMode = %v, want command prompt to keep focus", m.inputMode)
	}
	if m.commandInput.Value() != "/" {
		t.Fatalf("command input = %q, want slash typed into command prompt", m.commandInput.Value())
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(model)
	if m.inputMode != modeQuery {
		t.Fatalf("inputMode = %v, want modeQuery", m.inputMode)
	}
}

func TestQueryModeReloadsIssues(t *testing.T) {
	var issueRequests []youtrack.IssueListOptions
	m := newModel(context.Background(), fakeClient{
		issues:        []youtrack.Issue{{IDReadable: "ABC-3", Summary: "Three"}},
		issueRequests: &issueRequests,
	}, Options{Query: "project: ABC", Top: 25, Skip: 50})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{
		{IDReadable: "ABC-1", Summary: "One"},
		{IDReadable: "ABC-2", Summary: "Two"},
	}})
	m = updated.(model)
	m.selected = 1
	m.comments["ABC-2"] = []youtrack.Comment{{ID: "c-1", Text: "stale"}}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(model)
	if m.inputMode != modeQuery {
		t.Fatalf("inputMode = %v, want modeQuery", m.inputMode)
	}
	if m.queryInput.Value() != "project: ABC" {
		t.Fatalf("query input = %q, want current query", m.queryInput.Value())
	}

	m.queryInput.SetValue("project: DEF #Unresolved")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if cmd == nil {
		t.Fatal("query reload command = nil")
	}
	if m.inputMode != modeNavigation {
		t.Fatalf("inputMode = %v, want modeNavigation", m.inputMode)
	}
	if m.opts.Query != "project: DEF #Unresolved" {
		t.Fatalf("query = %q, want updated query", m.opts.Query)
	}
	if m.opts.Skip != 0 {
		t.Fatalf("skip = %d, want reset to first page", m.opts.Skip)
	}
	if m.selected != 0 {
		t.Fatalf("selected = %d, want reset to first issue", m.selected)
	}
	if len(m.comments) != 0 {
		t.Fatalf("comments cache = %#v, want cleared cache", m.comments)
	}

	msg := cmd().(issuesMsg)
	if msg.err != nil {
		t.Fatalf("reload error = %v", msg.err)
	}
	if len(issueRequests) != 1 || issueRequests[0].Query != "project: DEF #Unresolved" || issueRequests[0].Top != 25 || issueRequests[0].Skip != 0 {
		t.Fatalf("issue requests = %#v, want updated query and top", issueRequests)
	}
	if len(msg.issues) != 1 || msg.issues[0].IDReadable != "ABC-3" {
		t.Fatalf("reload issues = %#v, want fake client issues", msg.issues)
	}
}

func TestIssuePagingUsesSkip(t *testing.T) {
	var issueRequests []youtrack.IssueListOptions
	m := newModel(context.Background(), fakeClient{
		issues:        []youtrack.Issue{{IDReadable: "ABC-11", Summary: "Eleven"}},
		issueRequests: &issueRequests,
	}, Options{Query: "project: ABC", Top: 10})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{
		{IDReadable: "ABC-1", Summary: "One"},
		{IDReadable: "ABC-2", Summary: "Two"},
	}})
	m = updated.(model)
	m.selected = 1
	m.comments["ABC-2"] = []youtrack.Comment{{ID: "c-1", Text: "stale"}}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(model)
	if cmd == nil {
		t.Fatal("next page command = nil")
	}
	if m.opts.Skip != 10 {
		t.Fatalf("skip = %d, want next page skip", m.opts.Skip)
	}
	if m.selected != 0 {
		t.Fatalf("selected = %d, want reset to first issue", m.selected)
	}
	if len(m.comments) != 0 {
		t.Fatalf("comments cache = %#v, want cleared cache", m.comments)
	}

	msg := cmd().(issuesMsg)
	if msg.err != nil {
		t.Fatalf("next page error = %v", msg.err)
	}
	if len(issueRequests) != 1 || issueRequests[0].Skip != 10 || issueRequests[0].Top != 10 || issueRequests[0].Query != "project: ABC" {
		t.Fatalf("issue requests = %#v, want next page request", issueRequests)
	}
	if len(msg.issues) != 1 || msg.issues[0].IDReadable != "ABC-11" {
		t.Fatalf("next issues = %#v, want fake next page issues", msg.issues)
	}

	updated, _ = m.Update(issuesMsg{issues: msg.issues})
	m = updated.(model)
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updated.(model)
	if cmd == nil {
		t.Fatal("previous page command = nil")
	}
	if m.opts.Skip != 0 {
		t.Fatalf("skip = %d, want previous page skip", m.opts.Skip)
	}
	_ = cmd()
	if len(issueRequests) != 2 || issueRequests[1].Skip != 0 {
		t.Fatalf("issue requests = %#v, want previous page request", issueRequests)
	}
}

func TestQueryModeCancels(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{Query: "project: ABC"})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(model)
	m.queryInput.SetValue("project: DEF")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)

	if m.inputMode != modeNavigation || m.queryInput.Value() != "" {
		t.Fatalf("inputMode = %v, query input = %q; want canceled", m.inputMode, m.queryInput.Value())
	}
	if m.opts.Query != "project: ABC" {
		t.Fatalf("query = %q, want unchanged query", m.opts.Query)
	}
	if cmd != nil {
		t.Fatal("cancel command != nil")
	}
}

func TestDetailPaneScrollsAndResetsOnSelectionChange(t *testing.T) {
	description := strings.Join([]string{
		"one",
		"two",
		"three",
		"four",
		"five",
		"six",
	}, "\n")
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 8})
	m = updated.(model)
	updated, _ = m.Update(issuesMsg{issues: []youtrack.Issue{
		{IDReadable: "ABC-1", Summary: "One", Description: description},
		{IDReadable: "ABC-2", Summary: "Two", Description: description},
	}})
	m = updated.(model)
	_ = m.View()

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = updated.(model)
	if m.detail.YOffset == 0 {
		t.Fatal("detail YOffset = 0, want scrolled viewport")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	if m.detail.YOffset != 0 {
		t.Fatalf("detail YOffset = %d, want reset on selection change", m.detail.YOffset)
	}
}

func TestIssueListFollowsSelection(t *testing.T) {
	issues := make([]youtrack.Issue, 20)
	for i := range issues {
		issues[i] = youtrack.Issue{
			IDReadable: "ABC-" + strconv.Itoa(i+1),
			Summary:    "Issue " + strconv.Itoa(i+1),
		}
	}
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: issues})
	m = updated.(model)
	m.selected = 15

	view := m.issueList(40, 8)
	if !strings.Contains(view, "ABC-16") {
		t.Fatalf("issueList() = %q, want selected issue", view)
	}
	if strings.Contains(view, "ABC-1 ") {
		t.Fatalf("issueList() = %q, want early issues outside visible window", view)
	}
	if !strings.Contains(view, "16/20") {
		t.Fatalf("issueList() = %q, want position title", view)
	}
}

func TestVisibleIssueRangeCentersSelectionWithinBounds(t *testing.T) {
	start, end := visibleIssueRange(15, 20, 8)
	if start != 13 || end != 18 {
		t.Fatalf("visibleIssueRange(15, 20, 8) = %d, %d; want 13, 18", start, end)
	}

	start, end = visibleIssueRange(19, 20, 8)
	if start != 15 || end != 20 {
		t.Fatalf("visibleIssueRange(19, 20, 8) = %d, %d; want 15, 20", start, end)
	}

	start, end = visibleIssueRange(-4, 2, 8)
	if start != 0 || end != 2 {
		t.Fatalf("visibleIssueRange(-4, 2, 8) = %d, %d; want 0, 2", start, end)
	}
}

func TestModelErrorView(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{err: errors.New("network down")})
	m = updated.(model)
	view := m.View()
	if !strings.Contains(view, "network down") {
		t.Fatalf("View() = %q, want error", view)
	}
}

func TestIssueDetailRendersCustomFields(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{
		IDReadable: "ABC-1",
		Summary:    "One",
		Project:    youtrack.Project{ShortName: "ABC"},
		Custom:     []byte(`[{"name":"State","value":{"name":"Open"}},{"name":"Assignee","value":{"login":"jane"}}]`),
	}}})
	m = updated.(model)

	view := m.View()
	if !strings.Contains(view, "State: Open") {
		t.Fatalf("View() = %q, want state field", view)
	}
	if !strings.Contains(view, "Assignee: jane") {
		t.Fatalf("View() = %q, want assignee field", view)
	}
}
