package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func ctrl(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Mod: tea.ModCtrl} }
func key(r rune) tea.KeyPressMsg  { return tea.KeyPressMsg{Code: r} }
func TestEditingWorkflow(t *testing.T) {
	m := newModel(Store{t.TempDir()}, newSong("first"))
	m.enterInsert()
	m.Update(key(tea.KeyTab))
	if m.focus != 1 || !m.style.Focused() {
		t.Fatal("tab focus")
	}
	m.Update(key(tea.KeyF4))
	pasteText(&m, "A new chorus\nSecond line")
	if m.song.Sections[0].Lyrics != "A new chorus\nSecond line" {
		t.Fatal("paste failed")
	}
	m.Update(key(tea.KeyF5))
	m.Update(ctrl('a'))
	pasteText(&m, "Chorus x4")
	if m.arrangement.Value() != "Chorus x4" {
		t.Fatal("select-all paste did not replace arrangement", m.arrangement.Value())
	}
	if strings.Count(m.output.GetContent(), "A new chorus") != 4 {
		t.Fatal("live output not updated")
	}
	m.Update(ctrl('p'))
	before := m.song.Title
	pasteText(&m, "Remember this idea")
	if !strings.Contains(m.song.Notes, "Remember this idea") || m.song.Title != before {
		t.Fatal("modal paste routed to wrong editor")
	}
	m.Update(key(tea.KeyEscape))
	m.Update(key(tea.KeyEscape))
	m.Update(ctrl('t'))
	pasteText(&m, "Pre-Chorus")
	m.Update(key(tea.KeyEnter))
	if len(m.sections) != 5 || m.focus != 6 || m.modal != "" {
		t.Fatal("new section not focused", m.focus, m.modal)
	}
	m.Update(ctrl('s'))
	if m.dirty {
		t.Fatal(m.status)
	}
	saved, err := m.store.Load(m.song.ID)
	if err != nil || !strings.Contains(saved.Notes, "Remember this idea") {
		t.Fatal("save missing data", err)
	}
	firstID := m.song.ID
	m.Update(ctrl('n'))
	pasteText(&m, "second")
	m.Update(key(tea.KeyEnter))
	if m.song.Title != "second" || m.song.ID == firstID || m.focus != 2 {
		t.Fatal("new song failed")
	}
	m.Update(ctrl('s'))
	m.Update(ctrl('e'))
	pasteText(&m, "first")
	m.Update(key(tea.KeyEnter))
	if m.song.ID != firstID {
		t.Fatal("library selection failed", m.modalError)
	}
}
func TestPresetAndStaleAutosave(t *testing.T) {
	m := newModel(Store{t.TempDir()}, demoSong())
	old := m.song.Arrangement
	m.Update(ctrl('l'))
	m.Update(key(tea.KeyDown))
	m.Update(key(tea.KeyEnter))
	if m.song.Arrangement != structures[1].Value || !strings.Contains(m.song.Notes, old) {
		t.Fatal("preset lost previous arrangement")
	}
	stale := m.revision
	m.title.SetValue("newer")
	m.changed()
	m.Update(autosaveMsg{stale})
	if !m.dirty {
		t.Fatal("stale timer saved")
	}
	m.Update(autosaveMsg{m.revision})
	if m.dirty {
		t.Fatal("autosave did not save")
	}
	m.Update(ctrl('g'))
	pasteText(&m, "funk")
	m.Update(key(tea.KeyEnter))
	if !strings.Contains(m.song.Style, genres[7].Value) {
		t.Fatal("style preset not applied")
	}
}
func TestQuitPreservesUnsavedOnFailure(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	os.WriteFile(blocker, []byte("x"), 0600)
	m := newModel(Store{filepath.Join(blocker, "songs")}, demoSong())
	m.changed()
	_, cmd := m.Update(ctrl('q'))
	if cmd != nil || !m.dirty || !m.statusError {
		t.Fatal("quit despite save failure")
	}
}
func TestViewsFitTerminal(t *testing.T) {
	for _, size := range [][2]int{{150, 44}, {120, 36}, {110, 24}, {109, 24}, {80, 24}, {48, 18}, {40, 12}} {
		m := newModel(Store{t.TempDir()}, demoSong())
		m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		for focus := 0; focus <= m.outputFocus(); focus++ {
			m.setFocus(focus)
			v := m.View().Content
			if lipgloss.Width(v) > size[0] || lipgloss.Height(v) > size[1] {
				t.Errorf("size %v focus %d overflow %dx%d", size, focus, lipgloss.Width(v), lipgloss.Height(v))
			}
			if size[0] >= 48 && size[1] >= 18 && !strings.Contains(ansi.Strip(v), "quit") {
				t.Errorf("footer missing at %v", size)
			}
		}
		for _, modal := range []string{"help", "scratch", "new", "section", "genres", "structures", "library"} {
			m.modal = modal
			v := m.View().Content
			if lipgloss.Width(v) > size[0] || lipgloss.Height(v) > size[1] {
				t.Errorf("%s overflows at %v", modal, size)
			}
		}
	}
}
func TestPanelDimensions(t *testing.T) {
	v := panel("TITLE", "body", 40, 10, true, green)
	if lipgloss.Width(v) != 40 || lipgloss.Height(v) != 10 {
		t.Fatalf("panel dimensions %dx%d", lipgloss.Width(v), lipgloss.Height(v))
	}
}

func TestStyleSearchByPromptAndAppend(t *testing.T) {
	m := newModel(Store{t.TempDir()}, newSong("test"))
	m.style.SetValue("warm recording")
	m.Update(ctrl('g'))
	// Words can match separate parts of the full prompt, beyond the short description.
	m.input.SetValue("  CATHEDRAL   pipe-organ  ")
	choices := m.filteredChoices()
	if len(choices) != 1 || choices[0].title != "Cathedral garage" {
		t.Fatalf("instrument search returned %v", choices)
	}
	expected := "warm recording, " + choices[0].value
	m.Update(key(tea.KeyEnter))
	if m.style.Value() != expected || m.modal != "" {
		t.Fatal("search selection did not append the full prompt")
	}
}

func TestLargeStylePickerPaging(t *testing.T) {
	m := newModel(Store{t.TempDir()}, demoSong())
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.Update(ctrl('g'))
	m.Update(key(tea.KeyPgDown))
	if m.selection <= 1 {
		t.Fatal("page down did not advance a page")
	}
	for i := 0; i < len(m.choices); i++ {
		m.Update(key(tea.KeyPgDown))
	}
	if m.selection != len(m.choices)-1 {
		t.Fatal("cannot reach last preset")
	}
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, m.choices[m.selection].title) || !strings.Contains(view, "Enter appends") {
		t.Fatal("selected preset or controls clipped", view)
	}
	m.Update(key(tea.KeyPgUp))
	if m.selection >= len(m.choices)-2 {
		t.Fatal("page up did not advance a page")
	}
}

func runeKey(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Text: string(r)} }
func pasteText(m *model, text string) {
	if m.isPicker() && !m.searching {
		m.Update(runeKey('/'))
	}
	if (m.modal == "" || m.modal == "scratch") && !m.insert {
		m.Update(runeKey('i'))
	}
	m.Update(tea.PasteMsg{Content: text})
}
