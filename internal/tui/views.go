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
	header := m.sectionBar() + "\n"
	if m.inputMode == modeProject {
		return header + m.projectSelector(m.width, bodyHeight-1) + "\n" + m.footer()
	}
	if m.section == sectionIssues && m.loading {
		return header + panelStyle.Width(m.width).Height(bodyHeight-1).Render("Loading issues...") + "\n" + m.footer()
	}
	if m.section == sectionIssues && m.err != nil {
		return header + panelStyle.Width(m.width).Height(bodyHeight-1).Render("Error: "+m.err.Error()) + "\n" + m.footer()
	}
	if m.section != sectionIssues && m.resourcesLoading {
		return header + panelStyle.Width(m.width).Height(bodyHeight-1).Render("Loading "+m.section.title()+"...") + "\n" + m.footer()
	}
	if m.section != sectionIssues && m.resourcesErr != nil {
		return header + panelStyle.Width(m.width).Height(bodyHeight-1).Render("Error: "+m.resourcesErr.Error()) + "\n" + m.footer()
	}

	listWidth := max(28, m.width/3)
	detailWidth := max(40, m.width-listWidth-4)
	contentHeight := max(3, bodyHeight-1)
	left := m.leftList(listWidth, contentHeight)
	right := m.rightPane(detailWidth, contentHeight)
	return header + lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + m.footer()
}

func (m model) sectionBar() string {
	labels := make([]string, 0, len(sections))
	for _, s := range sections {
		label := fmt.Sprintf(" %s %s ", s.key(), s.title())
		if s == m.section {
			label = sectionSelectedStyle.Render(label)
		} else {
			label = sectionStyle.Render(label)
		}
		labels = append(labels, label)
	}
	return strings.Join(labels, " ")
}

func (m model) leftList(width, height int) string {
	if m.section == sectionIssues {
		return m.issueList(width, height)
	}
	return m.resourceList(width, height)
}

func (m model) rightPane(width, height int) string {
	if m.section == sectionIssues {
		return m.issuePane(width, height)
	}
	return m.resourcePane(width, height)
}

func (m model) issueList(width, height int) string {
	start, end := visibleIssueRange(m.selected, len(m.issues), height)
	rows := make([]string, 0, end-start+1)
	position := 0
	if len(m.issues) > 0 {
		position = min(max(m.selected, 0), len(m.issues)-1) + 1
	}
	titleText := fmt.Sprintf("Issues %d/%d", position, len(m.issues))
	if m.projectFilter != "" {
		titleText = fmt.Sprintf("Issues %s %d/%d", m.projectFilter, position, len(m.issues))
	}
	if page := m.issuePageNumber(); page > 1 {
		titleText = fmt.Sprintf("%s page %d", titleText, page)
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

func (m model) resourceList(width, height int) string {
	start, end := visibleResourceRange(m.resourceSelected, len(m.resources), height)
	rows := make([]string, 0, end-start+1)
	position := 0
	if len(m.resources) > 0 {
		position = min(max(m.resourceSelected, 0), len(m.resources)-1) + 1
	}
	rows = append(rows, titleStyle.Render(fmt.Sprintf("%s %d/%d", m.section.title(), position, len(m.resources))))
	for i, resource := range m.resources[start:end] {
		index := start + i
		line := truncate(resourceLine(resource), width-4)
		if index == m.resourceSelected {
			line = selectedStyle.Render(line)
		}
		rows = append(rows, line)
	}
	return panelStyle.Width(width).Height(height).Render(strings.Join(rows, "\n"))
}

func (m model) projectSelector(width, height int) string {
	if m.projectOptionsLoading {
		return panelStyle.Width(width).Height(height).Render("Loading projects...")
	}
	if m.projectOptionsErr != nil {
		return panelStyle.Width(width).Height(height).Render("Error: " + m.projectOptionsErr.Error())
	}
	rows := []string{
		titleStyle.Render("Project Selector"),
		helpStyle.Render("type to filter, enter to apply, esc to cancel"),
		"",
	}
	start, end := visibleResourceRange(m.projectOptionSelected, len(m.projectOptions), height-3)
	for i, option := range m.projectOptions[start:end] {
		index := start + i
		line := truncate(projectOptionLine(option), width-4)
		if index == m.projectOptionSelected {
			line = selectedStyle.Render(line)
		}
		rows = append(rows, line)
	}
	if len(m.projectOptions) == 0 {
		rows = append(rows, "No matching projects")
	}
	return panelStyle.Width(width).Height(height).Render(strings.Join(rows, "\n"))
}

func projectOptionLine(option projectOption) string {
	if option.Subtitle == "" {
		return firstNonEmpty(option.ID, option.Name)
	}
	return firstNonEmpty(option.ID, option.Name) + "  " + option.Name + "  " + option.Subtitle
}

func visibleResourceRange(selected, total, height int) (int, int) {
	return visibleIssueRange(selected, total, height)
}

func resourceLine(resource resourceItem) string {
	if resource.Subtitle == "" {
		return firstNonEmpty(resource.ID, resource.Title)
	}
	return strings.TrimSpace(firstNonEmpty(resource.ID, resource.Title) + "  " + resource.Subtitle)
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
	content := m.issuePaneContent(max(20, width-4))
	return m.renderDetailViewport(content, width, height)
}

func (m model) resourcePane(width, height int) string {
	if len(m.resources) == 0 {
		return m.renderDetailViewport("No "+strings.ToLower(m.section.title()), width, height)
	}
	selected := min(max(m.resourceSelected, 0), len(m.resources)-1)
	resource := m.resources[selected]
	content := resource.Body
	if resource.BodyMarkdown {
		content = renderMarkdown(content, max(20, width-4))
	}
	if strings.TrimSpace(content) == "" {
		content = titleStyle.Render(resource.Title)
	}
	return m.renderDetailViewport(content, width, height)
}

func (m model) issuePaneContent(width int) string {
	if len(m.issues) == 0 {
		return "No issues"
	}
	switch m.pane {
	case commentsPane:
		return m.issueComments(width)
	case linksPane:
		return m.issueLinks()
	case activitiesPane:
		return m.issueActivities()
	case workItemsPane:
		return m.issueWorkItems(width)
	case attachmentsPane:
		return m.issueAttachments(width)
	default:
		return m.issueDetail(width)
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
	if m.section == sectionIssues {
		m.detail.SetContent(m.issuePaneContent(max(20, detailWidth-4)))
		return
	}
	if len(m.resources) == 0 {
		m.detail.SetContent("")
		return
	}
	selected := min(max(m.resourceSelected, 0), len(m.resources)-1)
	resource := m.resources[selected]
	content := resource.Body
	if resource.BodyMarkdown {
		content = renderMarkdown(content, max(20, detailWidth-4))
	}
	m.detail.SetContent(content)
}

func (m model) issueDetail(width int) string {
	issue := m.issues[m.selected]
	lines := []string{
		titleStyle.Render(issue.IDReadable),
		inlineText(issue.Summary),
	}
	if fields := issueFieldPanel(issue, width); fields != "" {
		lines = append(lines, "", fields)
	}
	lines = append(lines, "", m.issueEvidencePanel(issue.IDReadable, width))
	lines = append(lines, "", renderMarkdown(issue.Description, width))
	return strings.Join(lines, "\n")
}

func (m model) issueComments(width int) string {
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
		lines = append(lines, author, renderMarkdown(comment.Text, width), "")
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
			lines = append(lines, fmt.Sprintf("%s  %s  %s  %s", firstNonEmpty(link.LinkType.Name, "Link"), linkDirection(link), issue.IDReadable, inlineText(issue.Summary)))
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

func (m model) issueWorkItems(width int) string {
	issueID := m.currentIssueID()
	if m.workItemsLoading {
		return "Loading work items..."
	}
	if m.workItemsErr != nil {
		return "Error: " + m.workItemsErr.Error()
	}
	workItems := m.workItems[issueID]
	if len(workItems) == 0 {
		return titleStyle.Render(issueID) + "\n\nNo work items"
	}

	lines := []string{titleStyle.Render(issueID), titleStyle.Render("Work Items"), ""}
	for _, item := range workItems {
		lines = append(lines, workItemLine(item))
		if inlineText(item.Text) != "" {
			lines = append(lines, renderMarkdown(item.Text, width))
		}
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func (m model) issueAttachments(width int) string {
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
			lines = append(lines, inlineText(attachment.Name)+"  "+details)
		} else {
			lines = append(lines, inlineText(attachment.Name))
		}
		if preview := m.renderedAttachmentPreview(attachment, width); preview != "" {
			lines = append(lines, preview)
		}
	}
	return strings.Join(lines, "\n")
}

func (m model) footer() string {
	if m.inputMode == modeCommand {
		return commandStyle.Render(m.commandInput.View())
	}
	if m.inputMode == modeComment {
		return commandStyle.Render(m.commentInput.View())
	}
	if m.inputMode == modeWorkItem {
		return commandStyle.Render(m.workItemInput.View())
	}
	if m.inputMode == modeQuery {
		return commandStyle.Render(m.queryInput.View())
	}
	if m.inputMode == modeProject {
		return commandStyle.Render(m.projectInput.View()) + "  " + helpStyle.Render("enter apply  esc cancel")
	}
	if m.inputMode == modeIssue {
		return commandStyle.Render(m.issueInput.View())
	}
	if m.commandErr != nil {
		return errorStyle.Render("command failed: "+m.commandErr.Error()) + "  " + helpStyle.Render("/ query  c comment  w work  n/p page  : command  esc cancel  q quit")
	}
	if m.commentErr != nil {
		return errorStyle.Render("comment failed: "+m.commentErr.Error()) + "  " + helpStyle.Render("/ query  c comment  w work  n/p page  : command  esc cancel  q quit")
	}
	if m.workItemErr != nil {
		return errorStyle.Render("work item failed: "+m.workItemErr.Error()) + "  " + helpStyle.Render("/ query  c comment  w work  n/p page  : command  esc cancel  q quit")
	}
	if m.status != "" {
		return statusStyle.Render(m.status) + "  " + m.help.View(m.keyMap())
	}
	return m.help.View(m.keyMap())
}

func (m model) keyMap() tuiKeyMap {
	if m.section != sectionIssues {
		return resourceKeyMap()
	}
	return issueKeyMap()
}
