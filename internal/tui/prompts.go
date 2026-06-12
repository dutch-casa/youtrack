package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newPrompt(prompt, placeholder string, charLimit int) textinput.Model {
	input := textinput.New()
	input.Prompt = prompt
	input.Placeholder = placeholder
	input.CharLimit = charLimit
	return input
}

func (m *model) openCommandPrompt() tea.Cmd {
	m.inputMode = modeCommand
	resetAndFocus(&m.commandInput)
	m.clearActionErrors()
	m.status = ""
	return textinput.Blink
}

func (m *model) openCommentPrompt() tea.Cmd {
	m.inputMode = modeComment
	resetAndFocus(&m.commentInput)
	m.clearActionErrors()
	m.status = ""
	return textinput.Blink
}

func (m *model) openWorkItemPrompt() tea.Cmd {
	m.inputMode = modeWorkItem
	resetAndFocus(&m.workItemInput)
	m.clearActionErrors()
	m.status = ""
	return textinput.Blink
}

func (m *model) openQueryPrompt() tea.Cmd {
	m.inputMode = modeQuery
	resetAndFocus(&m.queryInput)
	m.queryInput.SetValue(m.opts.Query)
	m.clearActionErrors()
	m.status = ""
	return textinput.Blink
}

func (m *model) closeCommandPrompt() {
	m.inputMode = modeNavigation
	clearInput(&m.commandInput)
}

func (m *model) closeCommentPrompt() {
	m.inputMode = modeNavigation
	clearInput(&m.commentInput)
}

func (m *model) closeWorkItemPrompt() {
	m.inputMode = modeNavigation
	clearInput(&m.workItemInput)
}

func (m *model) closeQueryPrompt() {
	m.inputMode = modeNavigation
	clearInput(&m.queryInput)
}

func (m *model) clearPrompts() {
	m.inputMode = modeNavigation
	clearInput(&m.commandInput)
	clearInput(&m.commentInput)
	clearInput(&m.workItemInput)
	clearInput(&m.queryInput)
}

func (m *model) clearActionErrors() {
	m.commandErr = nil
	m.commentErr = nil
	m.workItemErr = nil
}

func resetAndFocus(input *textinput.Model) {
	input.Reset()
	input.Focus()
}

func clearInput(input *textinput.Model) {
	input.Blur()
	input.Reset()
	input.SetValue("")
}
