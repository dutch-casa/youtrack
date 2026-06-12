package tui

import (
	"errors"
	"net/url"
	"path"
	"strings"
)

func (m model) currentBrowserURL() (string, error) {
	baseURL := strings.TrimSpace(m.opts.BaseURL)
	if baseURL == "" {
		return "", errors.New("browser URL unavailable: YouTrack base URL was not provided to the TUI")
	}
	switch m.section {
	case sectionIssues:
		issueID := m.currentIssueID()
		if issueID == "" {
			return "", errors.New("no selected issue to open")
		}
		return browserURL(baseURL, "issue", issueID)
	default:
		resource, ok := m.currentResource()
		if !ok {
			return "", errors.New("no selected item to open")
		}
		return m.resourceBrowserURL(baseURL, resource)
	}
}

func (m model) resourceBrowserURL(baseURL string, resource resourceItem) (string, error) {
	id := firstNonEmpty(resource.ID, resource.Title)
	if id == "" {
		return "", errors.New("selected item has no browser identifier")
	}
	switch m.section {
	case sectionKnowledge:
		return browserURL(baseURL, "articles", id)
	case sectionHelpdesk:
		return browserURL(baseURL, "projects", id)
	case sectionAgile:
		return browserURL(baseURL, "agiles", id)
	case sectionProjects:
		return browserURL(baseURL, "projects", id)
	case sectionUsers:
		return browserURL(baseURL, "users", id)
	default:
		return "", errors.New("current section has no browser URL")
	}
}

func (m model) currentResource() (resourceItem, bool) {
	if len(m.resources) == 0 {
		return resourceItem{}, false
	}
	selected := min(max(m.resourceSelected, 0), len(m.resources)-1)
	return m.resources[selected], true
}

func browserURL(baseURL string, elements ...string) (string, error) {
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("YouTrack base URL must include scheme and host")
	}
	parts := []string{strings.Trim(parsed.Path, "/")}
	for _, element := range elements {
		element = strings.Trim(element, "/")
		if element != "" {
			parts = append(parts, url.PathEscape(element))
		}
	}
	parsed.Path = path.Join(parts...)
	return parsed.String(), nil
}
