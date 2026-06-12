package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dutchcaz/youtrack/internal/youtrack"
)

type fakeClient struct {
	issues      []youtrack.Issue
	comments    []youtrack.Comment
	attachments []youtrack.Attachment
	activities  []youtrack.Activity
	links       []youtrack.IssueLink
	commands    *[]youtrack.ApplyCommandRequest
	err         error
}

func (f fakeClient) Issues(ctx context.Context, opts youtrack.IssueListOptions) ([]youtrack.Issue, error) {
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
	if !m.commandMode {
		t.Fatal("commandMode = false, want true")
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

	if m.commandMode || m.commandInput != "" {
		t.Fatalf("command mode = %v, input = %q; want canceled", m.commandMode, m.commandInput)
	}
	if cmd != nil {
		t.Fatal("cancel command != nil")
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
