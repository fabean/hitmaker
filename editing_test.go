package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestStyleGrowsShrinksAndScrolls(t *testing.T) {
	m := newModel(Store{t.TempDir()}, newSong("test"))
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 36})
	m.setFocus(1)
	initial := m.style.Height()
	pasteText(&m, strings.Repeat("warm analog synths, ", 16))
	if m.style.Height() <= initial {
		t.Fatal("wrapped style did not grow")
	}
	start, end := m.leftBounds(1)
	if start < m.leftOffset || end > m.leftOffset+m.height-8 {
		t.Fatal("expanded style is clipped")
	}
	// More content than the terminal can display must still be accepted.
	m.Update(ctrl('a'))
	long := strings.Repeat("full line\n", 80) + "last line"
	pasteText(&m, long)
	if m.style.Value() != long {
		t.Fatal("viewport height limited pasted content")
	}
	if m.style.Height() > m.height-12 {
		t.Fatal("style grew beyond terminal")
	}
	for _, size := range [][2]int{{150, 44}, {110, 18}, {80, 24}, {48, 18}} {
		m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		if m.style.Value() != long {
			t.Fatal("resize changed content")
		}
		for _, focus := range []int{1, 2, 3} {
			m.setFocus(focus)
			v := m.View().Content
			if lipgloss.Width(v) > size[0] || lipgloss.Height(v) > size[1] {
				t.Fatal("layout overflow", size)
			}
			if size[0] >= 110 {
				a, b := m.leftBounds(focus)
				if a < m.leftOffset || b > m.leftOffset+size[1]-8 {
					t.Fatal("active card hidden", size, focus)
				}
			}
		}
	}
	m.setFocus(1)
	m.Update(key(tea.KeyF8))
	if m.style.Height() != 3 {
		t.Fatal("style did not shrink on clear", m.style.Height())
	}
}

func TestAppendingStylePresetGrowsEditor(t *testing.T) {
	m := newModel(Store{t.TempDir()}, newSong("test"))
	before := m.style.Height()
	m.Update(ctrl('g'))
	m.input.SetValue("mewithoutYou")
	choices := m.filteredChoices()
	if len(choices) != 1 {
		t.Fatal("expected one matching preset")
	}
	expected := choices[0].value
	m.Update(key(tea.KeyEnter))
	if m.style.Height() <= before || m.song.Style != expected {
		t.Fatal("preset did not grow style editor")
	}
	if !strings.Contains(ansi.Strip(m.View().Content), "d clear") {
		t.Fatal("clear shortcut hidden")
	}
}

func TestClearAndRestoreFields(t *testing.T) {
	m := newModel(Store{t.TempDir()}, demoSong())
	for focus := 0; focus <= m.arrangementFocus(); focus++ {
		m.setFocus(focus)
		name, before, _ := m.editableField()
		if before == "" {
			t.Fatal("fixture is empty", name)
		}
		m.Update(key(tea.KeyF8))
		_, after, _ := m.editableField()
		if after != "" || !m.dirty {
			t.Fatal("field was not cleared", name)
		}
		m.Update(key(tea.KeyF8))
		m.Update(key(tea.KeyF9))
		_, restored, _ := m.editableField()
		if restored != before {
			t.Fatal("restore failed after repeated clear", name)
		}
	}
	m.setFocus(2)
	sectionName := m.song.Sections[0].Name
	m.Update(key(tea.KeyF8))
	if m.song.Sections[0].Name != sectionName || strings.Contains(m.output.GetContent(), "We are the last light") {
		t.Fatal("clear removed the section name or left stale output")
	}
	m.Update(autosaveMsg{m.revision})
	saved, err := m.store.Load(m.song.ID)
	if err != nil || saved.Sections[0].Lyrics != "" {
		t.Fatal("clear was not saved", err)
	}
	m.Update(key(tea.KeyF9))
	if !strings.Contains(m.output.GetContent(), "We are the last light") {
		t.Fatal("restore left stale output")
	}
	m.setFocus(m.outputFocus())
	before := m.contentSignature()
	m.Update(key(tea.KeyF8))
	if m.contentSignature() != before {
		t.Fatal("clearing generated output modified editors")
	}
}

func TestClearBackupsStayWithTheirFieldAndSong(t *testing.T) {
	m := newModel(Store{t.TempDir()}, demoSong())
	m.setFocus(m.arrangementFocus())
	old := m.arrangement.Value()
	m.Update(key(tea.KeyF8))
	m.Update(ctrl('t'))
	m.input.SetValue("Hook")
	m.Update(key(tea.KeyEnter))
	m.setFocus(m.arrangementFocus())
	m.Update(key(tea.KeyF9))
	if m.arrangement.Value() != old {
		t.Fatal("adding section broke arrangement restore")
	}
	m.setFocus(1)
	m.Update(key(tea.KeyF8))
	pasteText(&m, "replacement")
	m.Update(key(tea.KeyF9))
	if m.style.Value() != "replacement" {
		t.Fatal("restore overwrote new text")
	}
	m.Update(ctrl('p'))
	notes := m.scratch.Value()
	m.Update(key(tea.KeyF8))
	if m.song.Notes != "" {
		t.Fatal("scratchpad not cleared")
	}
	m.Update(key(tea.KeyF9))
	if m.song.Notes != notes {
		t.Fatal("scratchpad restore failed")
	}
	m.Update(key(tea.KeyF8))
	m.closeModal()
	m.load(newSong("other"))
	m.Update(ctrl('p'))
	m.Update(key(tea.KeyF9))
	if m.scratch.Value() != "" {
		t.Fatal("restore leaked another song's notes")
	}
}
