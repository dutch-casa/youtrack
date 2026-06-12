package tui

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"testing/quick"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dutch-casa/youtrack/internal/youtrack"
)

type fakeClient struct {
	issues        []youtrack.Issue
	issueRequests *[]youtrack.IssueListOptions
	comments      []youtrack.Comment
	commentAdds   *[]commentAdd
	attachments   []youtrack.Attachment
	activities    []youtrack.Activity
	workItems     []youtrack.WorkItem
	workItemAdds  *[]workItemAdd
	links         []youtrack.IssueLink
	commands      *[]youtrack.ApplyCommandRequest
	err           error
}

type commentAdd struct {
	issueID string
	text    string
}

type workItemAdd struct {
	issueID string
	minutes int
	text    string
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

func (f fakeClient) AddComment(ctx context.Context, issueID, text string) (youtrack.Comment, error) {
	if f.commentAdds != nil {
		*f.commentAdds = append(*f.commentAdds, commentAdd{issueID: issueID, text: text})
	}
	return youtrack.Comment{ID: "c-added", Text: text}, f.err
}

func (f fakeClient) Attachments(ctx context.Context, opts youtrack.AttachmentListOptions) ([]youtrack.Attachment, error) {
	return f.attachments, f.err
}

func (f fakeClient) Activities(ctx context.Context, opts youtrack.ActivityListOptions) ([]youtrack.Activity, error) {
	return f.activities, f.err
}

func (f fakeClient) WorkItems(ctx context.Context, opts youtrack.WorkItemListOptions) ([]youtrack.WorkItem, error) {
	return f.workItems, f.err
}

func (f fakeClient) AddWorkItem(ctx context.Context, req youtrack.AddWorkItemRequest) (youtrack.WorkItem, error) {
	if f.workItemAdds != nil {
		*f.workItemAdds = append(*f.workItemAdds, workItemAdd{issueID: req.IssueID, minutes: req.Minutes, text: req.Text})
	}
	return youtrack.WorkItem{ID: "w-added", Text: req.Text, Duration: youtrack.DurationValue{Minutes: req.Minutes}}, f.err
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

func TestTabLoadsWorkItemsPane(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)

	for range 4 {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
		m = updated.(model)
	}
	if m.pane != workItemsPane {
		t.Fatalf("pane = %v, want workItemsPane", m.pane)
	}
	if !m.workItemsLoading {
		t.Fatal("workItemsLoading = false, want true")
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

func TestWorkItemsPaneRendersWorkItems(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)
	m.pane = workItemsPane
	updated, _ = m.Update(workItemsMsg{
		issueID: "ABC-1",
		workItems: []youtrack.WorkItem{{
			Text:     "implementation",
			Date:     1704067200000,
			Duration: youtrack.DurationValue{Minutes: 90},
			Type:     youtrack.WorkItemType{Name: "Development"},
			Author:   youtrack.User{Login: "jane"},
		}},
	})
	m = updated.(model)

	view := m.View()
	for _, want := range []string{"Work Items", "1h 30m", "Development", "jane", "2024-01-01", "implementation"} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() = %q, want work item field %q", view, want)
		}
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

func TestCommentModeAddsComment(t *testing.T) {
	var commentAdds []commentAdd
	m := newModel(context.Background(), fakeClient{
		comments:    []youtrack.Comment{{Text: "new note", Author: youtrack.User{Login: "jane"}}},
		commentAdds: &commentAdds,
	}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)
	m.activities["ABC-1"] = []youtrack.Activity{{Type: "stale"}}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(model)
	if m.inputMode != modeComment {
		t.Fatalf("inputMode = %v, want modeComment", m.inputMode)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("  new note  ")})
	m = updated.(model)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if !m.commentRunning {
		t.Fatal("commentRunning = false, want true")
	}
	if cmd == nil {
		t.Fatal("add comment command = nil")
	}

	msg := cmd().(addCommentMsg)
	if msg.issueID != "ABC-1" || msg.text != "new note" {
		t.Fatalf("addCommentMsg = %#v", msg)
	}
	if len(commentAdds) != 1 || commentAdds[0].issueID != "ABC-1" || commentAdds[0].text != "new note" {
		t.Fatalf("commentAdds = %#v", commentAdds)
	}

	updated, reload := m.Update(msg)
	m = updated.(model)
	if m.commentRunning {
		t.Fatal("commentRunning = true, want false")
	}
	if m.pane != commentsPane {
		t.Fatalf("pane = %v, want commentsPane", m.pane)
	}
	if _, ok := m.activities["ABC-1"]; ok {
		t.Fatalf("activities cache = %#v, want invalidated selected issue", m.activities)
	}
	if !strings.Contains(m.status, "Commented on ABC-1") {
		t.Fatalf("status = %q, want commented status", m.status)
	}
	if reload == nil {
		t.Fatal("comment reload command = nil")
	}

	commentsMsg := reload().(commentsMsg)
	if commentsMsg.issueID != "ABC-1" || len(commentsMsg.comments) != 1 || commentsMsg.comments[0].Text != "new note" {
		t.Fatalf("commentsMsg = %#v, want reloaded comments", commentsMsg)
	}
}

func TestCommentModeCancels(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("new note")})
	m = updated.(model)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)

	if m.inputMode != modeNavigation || m.commentInput.Value() != "" {
		t.Fatalf("inputMode = %v, comment input = %q; want canceled", m.inputMode, m.commentInput.Value())
	}
	if cmd != nil {
		t.Fatal("cancel command != nil")
	}
}

func TestCommentResponseDoesNotReloadDifferentSelection(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{
		{IDReadable: "ABC-1", Summary: "One"},
		{IDReadable: "ABC-2", Summary: "Two"},
	}})
	m = updated.(model)
	m.selected = 1
	m.comments["ABC-1"] = []youtrack.Comment{{Text: "stale"}}
	m.comments["ABC-2"] = []youtrack.Comment{{Text: "current"}}

	updated, cmd := m.Update(addCommentMsg{issueID: "ABC-1", text: "new note"})
	m = updated.(model)
	if cmd != nil {
		t.Fatal("reload command != nil for non-selected commented issue")
	}
	if m.pane != detailsPane {
		t.Fatalf("pane = %v, want unchanged detailsPane", m.pane)
	}
	if _, ok := m.comments["ABC-1"]; ok {
		t.Fatalf("comments cache = %#v, want stale commented issue invalidated", m.comments)
	}
	if got := m.comments["ABC-2"]; len(got) != 1 || got[0].Text != "current" {
		t.Fatalf("current issue comments = %#v, want untouched current cache", got)
	}
}

func TestWorkItemModeAddsWorkItem(t *testing.T) {
	var workItemAdds []workItemAdd
	m := newModel(context.Background(), fakeClient{
		workItems: []youtrack.WorkItem{{
			Text:     "implementation",
			Duration: youtrack.DurationValue{Minutes: 90},
			Author:   youtrack.User{Login: "jane"},
		}},
		workItemAdds: &workItemAdds,
	}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)
	m.activities["ABC-1"] = []youtrack.Activity{{Type: "stale"}}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = updated.(model)
	if m.inputMode != modeWorkItem {
		t.Fatalf("inputMode = %v, want modeWorkItem", m.inputMode)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1h 30m implementation")})
	m = updated.(model)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if !m.workItemRunning {
		t.Fatal("workItemRunning = false, want true")
	}
	if cmd == nil {
		t.Fatal("add work item command = nil")
	}

	msg := cmd().(addWorkItemMsg)
	if msg.issueID != "ABC-1" || msg.draft.Minutes != 90 || msg.draft.Text != "implementation" {
		t.Fatalf("addWorkItemMsg = %#v", msg)
	}
	if len(workItemAdds) != 1 || workItemAdds[0].issueID != "ABC-1" || workItemAdds[0].minutes != 90 || workItemAdds[0].text != "implementation" {
		t.Fatalf("workItemAdds = %#v", workItemAdds)
	}

	updated, reload := m.Update(msg)
	m = updated.(model)
	if m.workItemRunning {
		t.Fatal("workItemRunning = true, want false")
	}
	if m.pane != workItemsPane {
		t.Fatalf("pane = %v, want workItemsPane", m.pane)
	}
	if _, ok := m.activities["ABC-1"]; ok {
		t.Fatalf("activities cache = %#v, want invalidated selected issue", m.activities)
	}
	if !strings.Contains(m.status, "Added 1h 30m to ABC-1") {
		t.Fatalf("status = %q, want added work status", m.status)
	}
	if reload == nil {
		t.Fatal("work item reload command = nil")
	}

	workItemsMsg := reload().(workItemsMsg)
	if workItemsMsg.issueID != "ABC-1" || len(workItemsMsg.workItems) != 1 || workItemsMsg.workItems[0].Text != "implementation" {
		t.Fatalf("workItemsMsg = %#v, want reloaded work items", workItemsMsg)
	}
}

func TestWorkItemModeRejectsInvalidDuration(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("soon implementation")})
	m = updated.(model)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if cmd != nil {
		t.Fatal("add work item command != nil for invalid duration")
	}
	if m.inputMode != modeWorkItem {
		t.Fatalf("inputMode = %v, want prompt to stay focused", m.inputMode)
	}
	if m.workItemErr == nil {
		t.Fatal("workItemErr = nil, want duration validation")
	}
}

func TestWorkItemModeCancels(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("45m implementation")})
	m = updated.(model)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)

	if m.inputMode != modeNavigation || m.workItemInput.Value() != "" {
		t.Fatalf("inputMode = %v, work item input = %q; want canceled", m.inputMode, m.workItemInput.Value())
	}
	if cmd != nil {
		t.Fatal("cancel command != nil")
	}
}

func TestWorkItemResponseDoesNotReloadDifferentSelection(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{
		{IDReadable: "ABC-1", Summary: "One"},
		{IDReadable: "ABC-2", Summary: "Two"},
	}})
	m = updated.(model)
	m.selected = 1
	m.workItems["ABC-1"] = []youtrack.WorkItem{{Text: "stale"}}
	m.workItems["ABC-2"] = []youtrack.WorkItem{{Text: "current"}}

	updated, cmd := m.Update(addWorkItemMsg{issueID: "ABC-1", draft: workItemDraft{Minutes: 45}})
	m = updated.(model)
	if cmd != nil {
		t.Fatal("reload command != nil for non-selected work item")
	}
	if m.pane != detailsPane {
		t.Fatalf("pane = %v, want unchanged detailsPane", m.pane)
	}
	if _, ok := m.workItems["ABC-1"]; ok {
		t.Fatalf("work items cache = %#v, want stale target issue invalidated", m.workItems)
	}
	if got := m.workItems["ABC-2"]; len(got) != 1 || got[0].Text != "current" {
		t.Fatalf("current issue work items = %#v, want untouched current cache", got)
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
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(model)
	if m.inputMode != modeComment {
		t.Fatalf("inputMode = %v, want modeComment", m.inputMode)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	m = updated.(model)
	if m.inputMode != modeComment {
		t.Fatalf("inputMode = %v, want comment prompt to keep focus", m.inputMode)
	}
	if m.commentInput.Value() != ":" {
		t.Fatalf("comment input = %q, want colon typed into comment prompt", m.commentInput.Value())
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = updated.(model)
	if m.inputMode != modeWorkItem {
		t.Fatalf("inputMode = %v, want modeWorkItem", m.inputMode)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(model)
	if m.inputMode != modeWorkItem {
		t.Fatalf("inputMode = %v, want work item prompt to keep focus", m.inputMode)
	}
	if m.workItemInput.Value() != "/" {
		t.Fatalf("work item input = %q, want slash typed into work item prompt", m.workItemInput.Value())
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(model)
	if m.inputMode != modeQuery {
		t.Fatalf("inputMode = %v, want modeQuery", m.inputMode)
	}
}

func TestRefreshClearsPromptStateFromNavigation(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{IDReadable: "ABC-1", Summary: "One"}}})
	m = updated.(model)

	m.commandInput.SetValue("State Fixed")
	m.commentInput.SetValue("ready")
	m.workItemInput.SetValue("45m implementation")
	m.queryInput.SetValue("project: ABC")
	promptErr := errors.New("prompt failed")
	m.commandErr = promptErr
	m.commentErr = promptErr
	m.workItemErr = promptErr
	m.commandRunning = true
	m.commentRunning = true
	m.workItemRunning = true

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updated.(model)
	if cmd == nil {
		t.Fatal("refresh command = nil")
	}
	if m.inputMode != modeNavigation {
		t.Fatalf("inputMode = %v, want modeNavigation", m.inputMode)
	}
	if m.commandInput.Value() != "" || m.commentInput.Value() != "" || m.workItemInput.Value() != "" || m.queryInput.Value() != "" {
		t.Fatalf("prompt values = command %q comment %q work %q query %q; want all cleared", m.commandInput.Value(), m.commentInput.Value(), m.workItemInput.Value(), m.queryInput.Value())
	}
	if m.commandErr != nil || m.commentErr != nil || m.workItemErr != nil {
		t.Fatalf("prompt errors = command %v comment %v work %v; want all cleared", m.commandErr, m.commentErr, m.workItemErr)
	}
	if m.commandRunning || m.commentRunning || m.workItemRunning {
		t.Fatalf("running flags = command %v comment %v work %v; want all stopped", m.commandRunning, m.commentRunning, m.workItemRunning)
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
	m.workItems["ABC-2"] = []youtrack.WorkItem{{Text: "stale time"}}

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
	if len(m.workItems) != 0 {
		t.Fatalf("work items cache = %#v, want cleared cache", m.workItems)
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
	updated, _ := m.Update(issuesMsg{issues: numberedIssues(10)})
	m = updated.(model)
	m.selected = 9
	m.comments["ABC-10"] = []youtrack.Comment{{ID: "c-1", Text: "stale"}}

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

func TestIssuePagingDoesNotAdvancePastPartialPage(t *testing.T) {
	var issueRequests []youtrack.IssueListOptions
	m := newModel(context.Background(), fakeClient{issueRequests: &issueRequests}, Options{Query: "project: ABC", Top: 10, Skip: 20})
	updated, _ := m.Update(issuesMsg{issues: numberedIssues(3)})
	m = updated.(model)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(model)
	if cmd != nil {
		t.Fatal("next page command != nil on partial page")
	}
	if m.opts.Skip != 20 {
		t.Fatalf("skip = %d, want unchanged skip", m.opts.Skip)
	}
	if len(issueRequests) != 0 {
		t.Fatalf("issue requests = %#v, want no request", issueRequests)
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
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: numberedIssues(20)})
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

func TestIssuePageNumber(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{Top: 25, Skip: 50})
	if page := m.issuePageNumber(); page != 3 {
		t.Fatalf("issuePageNumber() = %d, want 3", page)
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

func numberedIssues(total int) []youtrack.Issue {
	issues := make([]youtrack.Issue, total)
	for i := range issues {
		issues[i] = youtrack.Issue{
			IDReadable: "ABC-" + strconv.Itoa(i+1),
			Summary:    "Issue " + strconv.Itoa(i+1),
		}
	}
	return issues
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
		CustomFields: []youtrack.CustomField{
			{Name: "State", Value: "Open"},
			{Name: "Assignee", Value: "jane"},
		},
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

func TestIssueDetailRendersMetadataStrip(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 30})
	m = updated.(model)
	updated, _ = m.Update(issuesMsg{issues: []youtrack.Issue{{
		IDReadable: "ABC-1",
		Summary:    "One",
		Project:    youtrack.Project{ShortName: "ABC"},
		CustomFields: []youtrack.CustomField{
			{Name: "State", Value: "Open"},
			{Name: "Assignee", Value: "jane"},
			{Name: "Priority", Value: "Major"},
			{Name: "Type", Value: "Bug"},
		},
	}}})
	m = updated.(model)

	view := m.View()
	for _, want := range []string{"Project ABC", "State Open", "Assignee jane", "Priority Major", "Type Bug", "Resolved no"} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() = %q, want metadata %q", view, want)
		}
	}
}

func TestIssueListRendersStateSignal(t *testing.T) {
	m := newModel(context.Background(), fakeClient{}, Options{})
	updated, _ := m.Update(issuesMsg{issues: []youtrack.Issue{{
		IDReadable:   "ABC-1",
		Summary:      "One",
		CustomFields: []youtrack.CustomField{{Name: "State", Value: "Open"}},
	}}})
	m = updated.(model)

	view := m.issueList(40, 8)
	if !strings.Contains(view, "[Open] One") {
		t.Fatalf("issueList() = %q, want state signal", view)
	}
}

func TestParseWorkItemInput(t *testing.T) {
	tests := []struct {
		input       string
		wantMinutes int
		wantText    string
	}{
		{input: "45m implementation", wantMinutes: 45, wantText: "implementation"},
		{input: "1h implementation", wantMinutes: 60, wantText: "implementation"},
		{input: "1h30m implementation", wantMinutes: 90, wantText: "implementation"},
		{input: "1h 30m implementation notes", wantMinutes: 90, wantText: "implementation notes"},
		{input: "2H 5M", wantMinutes: 125, wantText: ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseWorkItemInput(tt.input)
			if err != nil {
				t.Fatalf("parseWorkItemInput() error = %v", err)
			}
			if got.Minutes != tt.wantMinutes || got.Text != tt.wantText {
				t.Fatalf("parseWorkItemInput() = %#v, want minutes %d text %q", got, tt.wantMinutes, tt.wantText)
			}
		})
	}
}

func TestParseWorkItemInputRejectsInvalidDuration(t *testing.T) {
	for _, input := range []string{"", "implementation", "45 implementation", "1d implementation", "0m implementation", "1hsoon"} {
		t.Run(input, func(t *testing.T) {
			if _, err := parseWorkItemInput(input); err == nil {
				t.Fatalf("parseWorkItemInput(%q) error = nil, want validation", input)
			}
		})
	}
}

func TestParseWorkItemInputDurationProperty(t *testing.T) {
	property := func(hours, minutes uint8) bool {
		if hours == 0 && minutes == 0 {
			return true
		}
		input := strconv.Itoa(int(hours)) + "h " + strconv.Itoa(int(minutes)) + "m work"
		draft, err := parseWorkItemInput(input)
		return err == nil && draft.Minutes == int(hours)*60+int(minutes) && draft.Text == "work"
	}
	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}

func FuzzParseWorkItemInput(f *testing.F) {
	for _, seed := range []string{
		"45m implementation",
		"1h implementation",
		"1h 30m implementation",
		"1h30m implementation",
		"implementation",
		"45 implementation",
		"0m",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		draft, err := parseWorkItemInput(input)
		if err != nil {
			return
		}
		if draft.Minutes <= 0 {
			t.Fatalf("parseWorkItemInput(%q) minutes = %d, want positive", input, draft.Minutes)
		}
		if strings.Contains(draft.Text, "\n") {
			t.Fatalf("parseWorkItemInput(%q) text contains newline: %q", input, draft.Text)
		}
	})
}
