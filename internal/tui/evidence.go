package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type issueEvidenceRow struct {
	Label string
	Value string
	Kind  string
}

func (m model) issueEvidencePanel(issueID string, width int) string {
	rows := []issueEvidenceRow{
		m.commentsEvidence(issueID),
		m.linksEvidence(issueID),
		m.activitiesEvidence(issueID),
		m.workItemsEvidence(issueID),
		m.attachmentsEvidence(issueID),
	}
	lines := make([]string, 0, len(rows)+1)
	lines = append(lines, titleStyle.Render("Evidence"))
	for _, row := range rows {
		lines = append(lines, fieldLabelStyle.Render(row.Label+":")+" "+evidenceStyle(row).Render(row.Value))
	}
	panelWidth := max(20, width-2)
	return evidencePanelStyle.Width(panelWidth).Render(strings.Join(lines, "\n"))
}

func (m model) commentsEvidence(issueID string) issueEvidenceRow {
	return evidenceFromCache("Comments", len(m.comments[issueID]), m.commentsLoading, m.commentsErr, hasCommentsCache(m, issueID))
}

func (m model) linksEvidence(issueID string) issueEvidenceRow {
	count := 0
	for _, link := range m.links[issueID] {
		count += len(linkedIssues(link))
	}
	return evidenceFromCache("Links", count, m.linksLoading, m.linksErr, hasLinksCache(m, issueID))
}

func (m model) activitiesEvidence(issueID string) issueEvidenceRow {
	return evidenceFromCache("Activity", len(m.activities[issueID]), m.activitiesLoading, m.activitiesErr, hasActivitiesCache(m, issueID))
}

func (m model) workItemsEvidence(issueID string) issueEvidenceRow {
	return evidenceFromCache("Work", len(m.workItems[issueID]), m.workItemsLoading, m.workItemsErr, hasWorkItemsCache(m, issueID))
}

func (m model) attachmentsEvidence(issueID string) issueEvidenceRow {
	return evidenceFromCache("Attachments", len(m.attachments[issueID]), m.attachmentsLoading, m.attachmentsErr, hasAttachmentsCache(m, issueID))
}

func evidenceFromCache(label string, count int, loading bool, err error, loaded bool) issueEvidenceRow {
	switch {
	case loading:
		return issueEvidenceRow{Label: label, Value: "loading", Kind: "loading"}
	case err != nil:
		return issueEvidenceRow{Label: label, Value: "error", Kind: "error"}
	case loaded:
		return issueEvidenceRow{Label: label, Value: fmt.Sprintf("%d", count), Kind: "count"}
	default:
		return issueEvidenceRow{Label: label, Value: "tab to load", Kind: "hint"}
	}
}

func evidenceStyle(row issueEvidenceRow) lipgloss.Style {
	switch row.Kind {
	case "count":
		return evidenceCountStyle
	case "error":
		return errorStyle
	case "loading":
		return fieldInfoStyle
	default:
		return evidenceHintStyle
	}
}

func hasCommentsCache(m model, issueID string) bool {
	_, ok := m.comments[issueID]
	return ok
}

func hasLinksCache(m model, issueID string) bool {
	_, ok := m.links[issueID]
	return ok
}

func hasActivitiesCache(m model, issueID string) bool {
	_, ok := m.activities[issueID]
	return ok
}

func hasWorkItemsCache(m model, issueID string) bool {
	_, ok := m.workItems[issueID]
	return ok
}

func hasAttachmentsCache(m model, issueID string) bool {
	_, ok := m.attachments[issueID]
	return ok
}
