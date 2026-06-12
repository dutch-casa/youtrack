package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if m.width == 0 {
		m.width = 100
	}
	if m.height == 0 {
		m.height = 30
	}

	bodyHeight := max(3, m.height-3)
	if m.loading {
		return panelStyle.Width(m.width).Height(bodyHeight).Render("Loading issues...") + "\n" + m.footer()
	}
	if m.err != nil {
		return panelStyle.Width(m.width).Height(bodyHeight).Render("Error: "+m.err.Error()) + "\n" + m.footer()
	}

	listWidth := max(28, m.width/3)
	detailWidth := max(40, m.width-listWidth-4)
	left := m.issueList(listWidth, bodyHeight)
	right := m.issuePane(detailWidth, bodyHeight)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + m.footer()
}

func (m model) issueList(width, height int) string {
	start, end := visibleIssueRange(m.selected, len(m.issues), height)
	rows := make([]string, 0, end-start+1)
	position := 0
	if len(m.issues) > 0 {
		position = min(max(m.selected, 0), len(m.issues)-1) + 1
	}
	titleText := fmt.Sprintf("Issues %d/%d", position, len(m.issues))
	if page := m.issuePageNumber(); page > 1 {
		titleText = fmt.Sprintf("Issues %d/%d page %d", position, len(m.issues), page)
	}
	title := titleStyle.Render(titleText)
	rows = append(rows, title)
	for i, issue := range m.issues[start:end] {
		index := start + i
		line := issueListLine(issue)
		line = truncate(line, width-4)
		if index == m.selected {
			line = selectedStyle.Render(line)
		}
		rows = append(rows, line)
	}
	return panelStyle.Width(width).Height(height).Render(strings.Join(rows, "\n"))
}

func (m model) issuePageNumber() int {
	if m.opts.Top <= 0 {
		return 1
	}
	return m.opts.Skip/m.opts.Top + 1
}

func visibleIssueRange(selected, total, height int) (int, int) {
	if total <= 0 || height <= 0 {
		return 0, 0
	}
	selected = min(max(selected, 0), total-1)
	slots := max(1, height-3)
	if total <= slots {
		return 0, total
	}

	start := selected - slots/2
	if start < 0 {
		start = 0
	}
	if start+slots > total {
		start = total - slots
	}
	return start, start + slots
}

func (m model) issuePane(width, height int) string {
	content := m.issuePaneContent()
	return m.renderDetailViewport(content, width, height)
}

func (m model) issuePaneContent() string {
	if len(m.issues) == 0 {
		return "No issues"
	}
	switch m.pane {
	case commentsPane:
		return m.issueComments()
	case linksPane:
		return m.issueLinks()
	case activitiesPane:
		return m.issueActivities()
	case attachmentsPane:
		return m.issueAttachments()
	default:
		return m.issueDetail()
	}
}

func (m model) renderDetailViewport(content string, width, height int) string {
	viewportWidth := max(1, width-4)
	viewportHeight := max(1, height-2)
	m.detail.Width = viewportWidth
	m.detail.Height = viewportHeight
	m.detail.SetContent(content)
	return panelStyle.Width(width).Height(height).Render(m.detail.View())
}

func (m *model) syncDetailViewport() {
	listWidth := max(28, m.width/3)
	detailWidth := max(40, m.width-listWidth-4)
	bodyHeight := max(3, m.height-3)
	m.detail.Width = max(1, detailWidth-4)
	m.detail.Height = max(1, bodyHeight-2)
	m.detail.SetContent(m.issuePaneContent())
}

func (m model) issueDetail() string {
	issue := m.issues[m.selected]
	lines := []string{
		titleStyle.Render(issue.IDReadable),
		issue.Summary,
		issueMetadataLine(issue),
		"",
		trimBlank(issue.Description),
	}
	fields := issueFields(issue)
	if len(fields) > 0 {
		lines = append(lines, "", titleStyle.Render("Fields"))
		lines = append(lines, fields...)
	}
	return strings.Join(lines, "\n")
}

func (m model) issueComments() string {
	issueID := m.currentIssueID()
	if m.commentsLoading {
		return "Loading comments..."
	}
	if m.commentsErr != nil {
		return "Error: " + m.commentsErr.Error()
	}
	comments := m.comments[issueID]
	if len(comments) == 0 {
		return titleStyle.Render(issueID) + "\n\nNo comments"
	}

	lines := []string{titleStyle.Render(issueID), titleStyle.Render("Comments"), ""}
	for _, comment := range comments {
		author := firstNonEmpty(comment.Author.FullName, comment.Author.Name, comment.Author.Login)
		lines = append(lines, author+": "+strings.TrimSpace(comment.Text), "")
	}
	return strings.Join(lines, "\n")
}

func (m model) issueLinks() string {
	issueID := m.currentIssueID()
	if m.linksLoading {
		return "Loading links..."
	}
	if m.linksErr != nil {
		return "Error: " + m.linksErr.Error()
	}
	links := m.links[issueID]
	if len(links) == 0 {
		return titleStyle.Render(issueID) + "\n\nNo links"
	}

	lines := []string{titleStyle.Render(issueID), titleStyle.Render("Links"), ""}
	for _, link := range links {
		issues := linkedIssues(link)
		if len(issues) == 0 {
			lines = append(lines, firstNonEmpty(link.LinkType.Name, "Link")+"  "+linkDirection(link))
			continue
		}
		for _, issue := range issues {
			lines = append(lines, fmt.Sprintf("%s  %s  %s  %s", firstNonEmpty(link.LinkType.Name, "Link"), linkDirection(link), issue.IDReadable, issue.Summary))
		}
	}
	return strings.Join(lines, "\n")
}

func (m model) issueActivities() string {
	issueID := m.currentIssueID()
	if m.activitiesLoading {
		return "Loading activity..."
	}
	if m.activitiesErr != nil {
		return "Error: " + m.activitiesErr.Error()
	}
	activities := m.activities[issueID]
	if len(activities) == 0 {
		return titleStyle.Render(issueID) + "\n\nNo activity"
	}

	lines := []string{titleStyle.Render(issueID), titleStyle.Render("Activity"), ""}
	for _, activity := range activities {
		author := firstNonEmpty(activity.Author.FullName, activity.Author.Name, activity.Author.Login)
		lines = append(lines, strings.TrimSpace(strings.Join(nonEmpty(author, activity.Summary()), "  ")))
	}
	return strings.Join(lines, "\n")
}

func (m model) issueAttachments() string {
	issueID := m.currentIssueID()
	if m.attachmentsLoading {
		return "Loading attachments..."
	}
	if m.attachmentsErr != nil {
		return "Error: " + m.attachmentsErr.Error()
	}
	attachments := m.attachments[issueID]
	if len(attachments) == 0 {
		return titleStyle.Render(issueID) + "\n\nNo attachments"
	}

	lines := []string{titleStyle.Render(issueID), titleStyle.Render("Attachments"), ""}
	for _, attachment := range attachments {
		author := firstNonEmpty(attachment.Author.FullName, attachment.Author.Name, attachment.Author.Login)
		size := formatBytes(attachment.Size)
		kind := firstNonEmpty(attachment.MimeType, attachment.Extension)
		details := strings.TrimSpace(strings.Join(nonEmpty(size, kind, author), "  "))
		if details != "" {
			lines = append(lines, attachment.Name+"  "+details)
			continue
		}
		lines = append(lines, attachment.Name)
	}
	return strings.Join(lines, "\n")
}

func (m model) footer() string {
	if m.inputMode == modeCommand {
		return commandStyle.Render(m.commandInput.View())
	}
	if m.inputMode == modeQuery {
		return commandStyle.Render(m.queryInput.View())
	}
	if m.commandErr != nil {
		return errorStyle.Render("command failed: "+m.commandErr.Error()) + "  " + helpStyle.Render("/ query  n/p page  : command  esc cancel  q quit")
	}
	if m.status != "" {
		return statusStyle.Render(m.status) + "  " + helpStyle.Render("/ query  n/p page  : command  tab panes  r refresh  q quit")
	}
	return helpStyle.Render("j/k move  / query  n/p page  tab details/comments/links/activity/attachments  pgup/pgdn scroll  : command  g/G top/bottom  r refresh  q quit")
}
