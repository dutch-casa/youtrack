package tui

import "github.com/charmbracelet/bubbles/key"

type tuiKeyMap struct {
	Sections key.Binding
	Move     key.Binding
	Search   key.Binding
	Project  key.Binding
	Issue    key.Binding
	Browser  key.Binding
	Action   key.Binding
	Comment  key.Binding
	Work     key.Binding
	Pages    key.Binding
	Panes    key.Binding
	Scroll   key.Binding
	Refresh  key.Binding
	Quit     key.Binding
}

func issueKeyMap() tuiKeyMap {
	return tuiKeyMap{
		Sections: key.NewBinding(key.WithKeys("1", "2", "3", "4", "5", "6"), key.WithHelp("1-6", "sections")),
		Move:     key.NewBinding(key.WithKeys("j", "k", "up", "down"), key.WithHelp("j/k", "move")),
		Search:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "query")),
		Project:  key.NewBinding(key.WithKeys("P"), key.WithHelp("P", "project")),
		Issue:    key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "issue")),
		Browser:  key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "browser")),
		Action:   key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "command")),
		Comment:  key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "comment")),
		Work:     key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "work")),
		Pages:    key.NewBinding(key.WithKeys("n", "p"), key.WithHelp("n/p", "page")),
		Panes:    key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "panes")),
		Scroll:   key.NewBinding(key.WithKeys("pgup", "pgdown", "ctrl+u", "ctrl+d"), key.WithHelp("pg", "scroll")),
		Refresh:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Quit:     key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

func resourceKeyMap() tuiKeyMap {
	return tuiKeyMap{
		Sections: key.NewBinding(key.WithKeys("1", "2", "3", "4", "5", "6"), key.WithHelp("1-6", "sections")),
		Move:     key.NewBinding(key.WithKeys("j", "k", "up", "down"), key.WithHelp("j/k", "move")),
		Search:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "live fuzzy")),
		Browser:  key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "browser")),
		Pages:    key.NewBinding(key.WithKeys("n", "p"), key.WithHelp("n/p", "page")),
		Scroll:   key.NewBinding(key.WithKeys("pgup", "pgdown", "ctrl+u", "ctrl+d"), key.WithHelp("pg", "scroll")),
		Refresh:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Quit:     key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

func helpdeskKeyMap() tuiKeyMap {
	keys := resourceKeyMap()
	keys.Search = key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "projects"))
	keys.Action = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "tickets"))
	return keys
}

func knowledgeKeyMap() tuiKeyMap {
	keys := resourceKeyMap()
	keys.Project = key.NewBinding(key.WithKeys("P"), key.WithHelp("P", "project"))
	return keys
}

func agileKeyMap() tuiKeyMap {
	keys := resourceKeyMap()
	keys.Action = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "sprints"))
	return keys
}

func (k tuiKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Sections, k.Move, k.Search, k.Project, k.Issue, k.Browser, k.Action, k.Comment, k.Work, k.Pages, k.Panes, k.Scroll, k.Refresh, k.Quit}
}

func (k tuiKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Sections, k.Move, k.Search, k.Project, k.Issue, k.Browser},
		{k.Action, k.Comment, k.Work, k.Pages, k.Panes, k.Scroll},
		{k.Refresh, k.Quit},
	}
}
