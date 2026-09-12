package main

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (m model) isPicker() bool {
	return m.modal == "library" || m.modal == "genres" || m.modal == "structures"
}

func (m model) modeName() string {
	if m.isPicker() && m.searching {
		return "SEARCH"
	}
	if m.insert || m.modal == "new" || m.modal == "section" {
		return "INSERT"
	}
	return "NORMAL"
}

func (m *model) enterInsert() tea.Cmd {
	if m.modal == "scratch" {
		m.insert = true
		return m.scratch.Focus()
	}
	if m.modal != "" || m.focus == m.outputFocus() {
		return nil
	}
	m.insert = true
	return m.setFocus(m.focus)
}
func (m *model) enterNormal() tea.Cmd {
	m.insert = false
	m.scratch.Blur()
	return m.setFocus(m.focus)
}

// Normal-mode operations act on whole fields. Text-editing keys only reach
// Bubbles in insert mode, so lyrics containing d/y/q cannot trigger actions.
func (m *model) normalKey(key string) tea.Cmd {
	switch key {
	case "i", "a", "enter":
		return m.enterInsert()
	case "c":
		cmd := m.clearField(false)
		return tea.Batch(cmd, m.enterInsert())
	case "d":
		return m.clearField(false)
	case "u":
		return m.clearField(true)
	case "y":
		return m.yankField()
	case "j", "down":
		if m.focus == m.outputFocus() {
			m.output.ScrollDown(1)
			return nil
		}
		if m.focus == m.arrangementFocus() {
			m.arrangement.CursorDown()
			return nil
		}
		return m.setFocus(min(m.arrangementFocus()-1, m.focus+1))
	case "k", "up":
		if m.focus == m.outputFocus() {
			m.output.ScrollUp(1)
			return nil
		}
		if m.focus == m.arrangementFocus() {
			m.arrangement.CursorUp()
			return nil
		}
		return m.setFocus(max(0, m.focus-1))
	case "h", "left":
		if m.focus == m.outputFocus() {
			return m.setFocus(m.arrangementFocus())
		}
		if m.focus == m.arrangementFocus() {
			return m.setFocus(m.leftFocus)
		}
	case "l", "right":
		if m.focus < m.arrangementFocus() {
			return m.setFocus(m.arrangementFocus())
		}
		return m.setFocus(m.outputFocus())
	case "pgdown":
		if m.focus == m.outputFocus() {
			m.output.PageDown()
		}
	case "pgup":
		if m.focus == m.outputFocus() {
			m.output.PageUp()
		}
	case "?":
		m.modal = "help"
		m.help.GotoTop()
	case "n", "e", "t", "s", "r", "m", "w", "Y", "b", "x", "q":
		aliases := map[string]rune{"n": 'n', "e": 'e', "t": 't', "s": 'g', "r": 'l', "m": 'p', "w": 's', "Y": 'o', "b": 'b', "x": 'x', "q": 'q'}
		_, cmd := m.Update(tea.KeyPressMsg{Code: aliases[key], Mod: tea.ModCtrl})
		return cmd
	}
	return nil
}

func (m *model) yankField() tea.Cmd {
	m.syncSong()
	name, value, _ := m.editableField()
	if m.modal == "" && m.focus == m.outputFocus() {
		_, cmd := m.Update(tea.KeyPressMsg{Code: 'o', Mod: tea.ModCtrl})
		return cmd
	}
	if strings.TrimSpace(value) == "" {
		m.status = "Nothing to copy in this field"
		m.statusError = false
		m.statusNotice = true
		return nil
	}
	return copyText(value, name)
}
