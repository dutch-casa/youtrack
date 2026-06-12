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
	switch m.pane {
	case commentsPane:
		return m.issueComments(width, height)
	case linksPane:
		return m.issueLinks(width, height)
	case activitiesPane:
		return m.issueActivities(width, height)
	case attachmentsPane:
		return m.issueAttachments(width, height)
	default:
		return m.issueDetail(width, height)
	}
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

func (m model) issueLinks(width, height int) string {
	issueID := m.currentIssueID()
	if m.linksLoading {
		return panelStyle.Width(width).Height(height).Render("Loading links...")
	}
	if m.linksErr != nil {
		return panelStyle.Width(width).Height(height).Render("Error: " + m.linksErr.Error())
	}
	links := m.links[issueID]
	if len(links) == 0 {
		return panelStyle.Width(width).Height(height).Render(titleStyle.Render(issueID) + "\n\nNo links")
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
	return panelStyle.Width(width).Height(height).Render(truncateBlock(strings.Join(lines, "\n"), width-4, height-2))
}

func (m model) issueActivities(width, height int) string {
	issueID := m.currentIssueID()
	if m.activitiesLoading {
		return panelStyle.Width(width).Height(height).Render("Loading activity...")
	}
	if m.activitiesErr != nil {
		return panelStyle.Width(width).Height(height).Render("Error: " + m.activitiesErr.Error())
	}
	activities := m.activities[issueID]
	if len(activities) == 0 {
		return panelStyle.Width(width).Height(height).Render(titleStyle.Render(issueID) + "\n\nNo activity")
	}

	lines := []string{titleStyle.Render(issueID), titleStyle.Render("Activity"), ""}
	for _, activity := range activities {
		author := firstNonEmpty(activity.Author.FullName, activity.Author.Name, activity.Author.Login)
		lines = append(lines, strings.TrimSpace(strings.Join(nonEmpty(author, activity.Summary()), "  ")))
	}
	return panelStyle.Width(width).Height(height).Render(truncateBlock(strings.Join(lines, "\n"), width-4, height-2))
}

func (m model) issueAttachments(width, height int) string {
	issueID := m.currentIssueID()
	if m.attachmentsLoading {
		return panelStyle.Width(width).Height(height).Render("Loading attachments...")
	}
	if m.attachmentsErr != nil {
		return panelStyle.Width(width).Height(height).Render("Error: " + m.attachmentsErr.Error())
	}
	attachments := m.attachments[issueID]
	if len(attachments) == 0 {
		return panelStyle.Width(width).Height(height).Render(titleStyle.Render(issueID) + "\n\nNo attachments")
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
	return panelStyle.Width(width).Height(height).Render(truncateBlock(strings.Join(lines, "\n"), width-4, height-2))
}

func footer() string {
	return helpStyle.Render("j/k move  tab details/comments/links/activity/attachments  g/G top/bottom  r refresh  q quit")
}
