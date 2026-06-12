package tui

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/dutch-casa/youtrack/internal/youtrack"
)

func articleResources(articles []youtrack.Article) []resourceItem {
	resources := make([]resourceItem, 0, len(articles))
	for _, article := range articles {
		id := firstNonEmpty(article.IDReadable, article.ID)
		project := article.Project.ShortName
		resources = append(resources, resourceItem{
			ID:           id,
			Title:        firstNonEmpty(article.Summary, id),
			Subtitle:     strings.TrimSpace(strings.Join(nonEmpty(project, article.Reporter.Login), "  ")),
			Body:         articleBody(article),
			BodyMarkdown: true,
		})
	}
	return resources
}

func articleBody(article youtrack.Article) string {
	id := firstNonEmpty(article.IDReadable, article.ID)
	lines := []string{
		"# " + strings.TrimSpace(strings.Join(nonEmpty(id, firstNonEmpty(article.Summary, "Untitled article")), " ")),
		strings.TrimSpace(strings.Join(nonEmpty("Article: "+id, "Project: "+article.Project.ShortName, "Author: "+article.Reporter.Login), "  ")),
		"",
		trimBlank(article.Content),
	}
	return strings.Join(lines, "\n")
}

func agileResources(agiles []youtrack.Agile) []resourceItem {
	resources := make([]resourceItem, 0, len(agiles))
	for _, agile := range agiles {
		projectNames := make([]string, 0, len(agile.Projects))
		for _, project := range agile.Projects {
			projectNames = append(projectNames, project.ShortName)
		}
		resources = append(resources, resourceItem{
			ID:       agile.ID,
			Title:    firstNonEmpty(agile.Name, agile.ID),
			Subtitle: strings.TrimSpace(strings.Join(nonEmpty(agile.Owner.Login, strings.Join(projectNames, ", ")), "  ")),
			Body:     agileBody(agile, projectNames),
		})
	}
	return resources
}

func agileBody(agile youtrack.Agile, projectNames []string) string {
	lines := []string{
		titleStyle.Render(firstNonEmpty(agile.Name, agile.ID)),
		"ID: " + agile.ID,
	}
	if agile.Owner.Login != "" {
		lines = append(lines, "Owner: "+agile.Owner.Login)
	}
	if len(projectNames) > 0 {
		lines = append(lines, "Projects: "+strings.Join(projectNames, ", "))
	}
	if agile.CurrentSprint.ID != "" || agile.CurrentSprint.Name != "" {
		lines = append(lines, "", titleStyle.Render("Current Sprint"))
		lines = append(lines, sprintLine(agile.CurrentSprint))
	}
	return strings.Join(lines, "\n")
}

func projectResources(projects []youtrack.Project) []resourceItem {
	resources := make([]resourceItem, 0, len(projects))
	for _, project := range projects {
		kind := firstNonEmpty(project.ProjectType.Name, "standard")
		hasDescription := strings.TrimSpace(project.Description) != ""
		body := projectBody(project, kind)
		if hasDescription {
			body = projectMarkdownBody(project, kind)
		}
		resources = append(resources, resourceItem{
			ID:           firstNonEmpty(project.ShortName, project.ID),
			Title:        firstNonEmpty(project.Name, project.ShortName, project.ID),
			Subtitle:     strings.TrimSpace(strings.Join(nonEmpty(project.ShortName, kind, project.Leader.Login), "  ")),
			Body:         body,
			BodyMarkdown: hasDescription,
		})
	}
	return resources
}

func projectOptions(projects []youtrack.Project) []projectOption {
	options := make([]projectOption, 0, len(projects))
	for _, project := range projects {
		id := firstNonEmpty(project.ShortName, project.ID)
		kind := firstNonEmpty(project.ProjectType.Name, "standard")
		options = append(options, projectOption{
			ID:       id,
			Name:     firstNonEmpty(project.Name, id),
			Subtitle: strings.TrimSpace(strings.Join(nonEmpty(kind, project.Leader.Login), "  ")),
		})
	}
	return options
}

func filterProjectOptions(options []projectOption, query string) []projectOption {
	query = strings.TrimSpace(query)
	if query == "" {
		return append([]projectOption(nil), options...)
	}
	matches := make([]projectOptionMatch, 0, len(options))
	for _, option := range options {
		score, ok := fuzzyScore(projectOptionSearchText(option), query)
		if ok {
			matches = append(matches, projectOptionMatch{option: option, score: score})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score == matches[j].score {
			return matches[i].option.ID < matches[j].option.ID
		}
		return matches[i].score > matches[j].score
	})
	filtered := make([]projectOption, 0, len(matches))
	for _, match := range matches {
		filtered = append(filtered, match.option)
	}
	return filtered
}

type projectOptionMatch struct {
	option projectOption
	score  int
}

func projectOptionSearchText(option projectOption) string {
	return strings.Join(nonEmpty(option.ID, option.Name, option.Subtitle), " ")
}

func projectBody(project youtrack.Project, kind string) string {
	lines := []string{
		titleStyle.Render(firstNonEmpty(project.Name, project.ShortName, project.ID)),
		"Short name: " + project.ShortName,
		"Type: " + kind,
	}
	if project.Leader.Login != "" {
		lines = append(lines, "Leader: "+project.Leader.Login)
	}
	if project.Archived {
		lines = append(lines, statusStyle.Render("Archived"))
	}
	if strings.TrimSpace(project.Description) != "" {
		lines = append(lines, "", terminalText(project.Description))
	}
	return strings.Join(lines, "\n")
}

func projectMarkdownBody(project youtrack.Project, kind string) string {
	lines := []string{
		"# " + firstNonEmpty(project.Name, project.ShortName, project.ID),
		"- Short name: " + project.ShortName,
		"- Type: " + kind,
	}
	if project.Leader.Login != "" {
		lines = append(lines, "- Leader: "+project.Leader.Login)
	}
	if project.Archived {
		lines = append(lines, "- Archived")
	}
	lines = append(lines, "", terminalText(project.Description))
	return strings.Join(lines, "\n")
}

func userResources(users []youtrack.User) []resourceItem {
	resources := make([]resourceItem, 0, len(users))
	for _, user := range users {
		name := firstNonEmpty(user.FullName, user.Name, user.Login, user.ID)
		state := "offline"
		if user.Online {
			state = "online"
		}
		if user.Banned {
			state = "banned"
		}
		resources = append(resources, resourceItem{
			ID:       firstNonEmpty(user.Login, user.ID),
			Title:    name,
			Subtitle: strings.TrimSpace(strings.Join(nonEmpty(user.Login, user.Email, state), "  ")),
			Body:     userBody(user, state),
		})
	}
	return resources
}

func userBody(user youtrack.User, state string) string {
	lines := []string{
		titleStyle.Render(firstNonEmpty(user.FullName, user.Name, user.Login, user.ID)),
		"Login: " + user.Login,
		"State: " + state,
	}
	if user.Email != "" {
		lines = append(lines, "Email: "+user.Email)
	}
	return strings.Join(lines, "\n")
}

func filterResources(resources []resourceItem, query string) []resourceItem {
	query = strings.TrimSpace(query)
	if query == "" {
		return append([]resourceItem(nil), resources...)
	}
	matches := make([]resourceMatch, 0, len(resources))
	for _, resource := range resources {
		score, ok := fuzzyScore(resourceSearchText(resource), query)
		if ok {
			matches = append(matches, resourceMatch{resource: resource, score: score})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score == matches[j].score {
			return matches[i].resource.Title < matches[j].resource.Title
		}
		return matches[i].score > matches[j].score
	})
	filtered := make([]resourceItem, 0, len(matches))
	for _, match := range matches {
		filtered = append(filtered, match.resource)
	}
	return filtered
}

type resourceMatch struct {
	resource resourceItem
	score    int
}

func resourceSearchText(resource resourceItem) string {
	return strings.Join(nonEmpty(resource.ID, resource.Title, resource.Subtitle, resource.Body), " ")
}

func fuzzyScore(text, query string) (int, bool) {
	textRunes := []rune(strings.ToLower(text))
	queryRunes := []rune(strings.ToLower(query))
	if len(queryRunes) == 0 {
		return 0, true
	}
	score := 0
	queryIndex := 0
	lastMatch := -1
	for textIndex, r := range textRunes {
		if queryIndex >= len(queryRunes) {
			break
		}
		if r != queryRunes[queryIndex] {
			continue
		}
		score += 10
		if textIndex == 0 || unicode.IsSpace(textRunes[textIndex-1]) || strings.ContainsRune("-_/.:", textRunes[textIndex-1]) {
			score += 5
		}
		if lastMatch == textIndex-1 {
			score += 3
		}
		lastMatch = textIndex
		queryIndex++
	}
	if queryIndex != len(queryRunes) {
		return 0, false
	}
	score -= max(0, lastMatch-len(queryRunes)+1)
	return score, true
}

func sprintLine(sprint youtrack.Sprint) string {
	parts := nonEmpty(firstNonEmpty(sprint.Name, sprint.ID))
	if sprint.Archived {
		parts = append(parts, "archived")
	}
	if sprint.IsDefault {
		parts = append(parts, "default")
	}
	return strings.Join(parts, "  ")
}

func (s section) title() string {
	switch s {
	case sectionKnowledge:
		return "Knowledge Base"
	case sectionHelpdesk:
		return "Help Desk"
	case sectionAgile:
		return "Agile Boards"
	case sectionProjects:
		return "Projects"
	case sectionUsers:
		return "Users"
	default:
		return "Issues"
	}
}

func (s section) key() string {
	switch s {
	case sectionKnowledge:
		return "2"
	case sectionHelpdesk:
		return "3"
	case sectionAgile:
		return "4"
	case sectionProjects:
		return "5"
	case sectionUsers:
		return "6"
	default:
		return "1"
	}
}

func sectionAtX(x int) (section, bool) {
	cursor := 0
	for _, s := range sections {
		label := fmt.Sprintf(" %s %s ", s.key(), s.title())
		next := cursor + len(label)
		if x >= cursor && x < next {
			return s, true
		}
		cursor = next + 1
	}
	return sectionIssues, false
}
