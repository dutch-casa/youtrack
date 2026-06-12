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
