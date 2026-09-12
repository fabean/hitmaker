package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestNormalInsertAndPaneNavigation(t *testing.T) {
	m := newModel(Store{t.TempDir()}, demoSong())
	if m.insert || m.title.Focused() {
		t.Fatal("must start in normal mode without an editing cursor")
	}
	before := m.contentSignature()
	m.Update(tea.PasteMsg{Content: "do not insert"})
	if m.contentSignature() != before {
		t.Fatal("paste edited text in normal mode")
	}
	m.Update(runeKey('j'))
	if m.focus != 1 {
		t.Fatal("j did not select style")
	}
	m.Update(runeKey('l'))
	if m.focus != m.arrangementFocus() {
		t.Fatal("l did not select arrangement")
	}
	m.Update(runeKey('l'))
	if m.focus != m.outputFocus() {
		t.Fatal("l did not select output")
	}
	m.Update(runeKey('h'))
	m.Update(runeKey('h'))
	if m.focus != 1 {
		t.Fatal("h lost the left field selection")
	}
	m.Update(runeKey('i'))
	if !m.insert || !m.style.Focused() {
		t.Fatal("i did not enable typing")
	}
	m.Update(ctrl('a'))
	m.Update(tea.PasteMsg{Content: ""})
	m.style.SetValue("")
	for _, r := range "dyuqhjkl" {
		m.Update(runeKey(r))
	}
	if m.style.Value() != "dyuqhjkl" {
		t.Fatal("insert-mode text was handled as commands", m.style.Value())
	}
	m.Update(key(tea.KeyEscape))
	if m.insert || m.style.Focused() {
		t.Fatal("escape did not leave insert mode")
	}
	m.Update(runeKey('d'))
	if m.style.Value() != "" {
		t.Fatal("d did not clear field")
	}
	m.Update(runeKey('u'))
	if m.style.Value() != "dyuqhjkl" {
		t.Fatal("u did not restore field")
	}
	m.Update(runeKey('c'))
	if !m.insert || m.style.Value() != "" {
		t.Fatal("c did not clear and enter insert mode")
	}
}

func TestNormalYanksFocusedFieldAndOutput(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // Capture OSC 52 requests without touching the real clipboard.
	m := newModel(Store{t.TempDir()}, demoSong())
	for _, f := range []int{0, 1, 2, m.arrangementFocus(), m.outputFocus()} {
		m.setFocus(f)
		_, cmd := m.Update(runeKey('y'))
		if cmd == nil {
			t.Fatal("no copy command", f)
		}
		msg, ok := cmd().(clipboardFallback)
		if !ok {
			t.Fatal("unexpected copy result")
		}
		_, want, _ := m.editableField()
		if f == m.outputFocus() {
			want, _ = m.song.Render()
		}
		if msg.value != want {
			t.Fatal("y copied the wrong field", f)
		}
	}
	m.setFocus(1)
	_, cmd := m.Update(runeKey('Y'))
	msg := cmd().(clipboardFallback)
	want, _ := m.song.Render()
	if msg.value != want {
		t.Fatal("Y did not copy full song")
	}
}

func TestVimPickerSearchAndScratchpad(t *testing.T) {
	m := newModel(Store{t.TempDir()}, demoSong())
	m.Update(runeKey('s'))
	if m.modal != "genres" || m.searching {
		t.Fatal("style picker should start in browse mode")
	}
	m.Update(runeKey('j'))
	if m.selection != 1 || m.input.Value() != "" {
		t.Fatal("j typed instead of moving selection")
	}
	m.Update(runeKey('/'))
	m.Update(tea.PasteMsg{Content: "jazz"})
	m.Update(runeKey('k'))
	if m.input.Value() != "jazzk" {
		t.Fatal("search consumed k as navigation")
	}
	m.input.SetValue("Cathedral garage")
	m.Update(key(tea.KeyEscape))
	if m.searching || m.modal != "genres" {
		t.Fatal("escape should stop searching before closing")
	}
	m.Update(key(tea.KeyEnter))
	if m.modal != "" || m.insert || !strings.Contains(m.style.Value(), "pipe-organ") {
		t.Fatal("picker did not apply in normal mode")
	}
	m.Update(runeKey('m'))
	if m.modal != "scratch" || m.insert {
		t.Fatal("scratchpad should start normal")
	}
	old := m.scratch.Value()
	m.Update(runeKey('d'))
	if m.scratch.Value() != "" {
		t.Fatal("scratch d failed")
	}
	m.Update(runeKey('u'))
	if m.scratch.Value() != old {
		t.Fatal("scratch u failed")
	}
	m.Update(runeKey('i'))
	m.Update(tea.PasteMsg{Content: "new notes"})
	m.Update(key(tea.KeyEscape))
	if m.insert || m.modal != "scratch" {
		t.Fatal("first escape should exit insert")
	}
	m.Update(runeKey('q'))
	if m.modal != "" {
		t.Fatal("q should close scratchpad")
	}
}

func TestVimNewSaveAndQuit(t *testing.T) {
	m := newModel(Store{t.TempDir()}, demoSong())
	m.Update(runeKey('n'))
	m.Update(tea.PasteMsg{Content: "new song"})
	m.Update(key(tea.KeyEnter))
	if m.song.Title != "new song" || !m.insert || m.focus != 2 {
		t.Fatal("new song should start editing chorus")
	}
	m.Update(tea.PasteMsg{Content: "a chorus"})
	m.Update(key(tea.KeyEscape))
	m.Update(runeKey('w'))
	loaded, err := m.store.Load(m.song.ID)
	if err != nil || loaded.Sections[0].Lyrics != "a chorus" {
		t.Fatal("w did not save", err)
	}
	_, cmd := m.Update(runeKey('q'))
	if cmd == nil {
		t.Fatal("q did not quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("wrong quit command")
	}
}

func TestVimHints(t *testing.T) {
	for _, width := range []int{120, 80, 48} {
		m := newModel(Store{t.TempDir()}, demoSong())
		m.Update(tea.WindowSizeMsg{Width: width, Height: 24})
		view := ansi.Strip(m.View().Content)
		if !strings.Contains(view, "NORMAL") || strings.Contains(view, "F1") || strings.Contains(view, "F8") || strings.Contains(view, "^Q") {
			t.Fatal("stale shortcuts in normal view", width)
		}
		m.Update(runeKey('i'))
		view = ansi.Strip(m.View().Content)
		if !strings.Contains(view, "INSERT") || strings.Contains(view, "d clear") {
			t.Fatal("normal actions advertised in insert mode", width)
		}
	}
}

func TestArtistArrangementFromNormalMode(t *testing.T) {
	m := newModel(Store{t.TempDir()}, demoSong())
	old, style := m.song.Arrangement, m.song.Style
	m.Update(runeKey('r'))
	m.Update(runeKey('/'))
	m.Update(tea.PasteMsg{Content: "mewithoutYou crescendo"})
	if len(m.filteredChoices()) != 1 {
		t.Fatal("artist arrangement not searchable")
	}
	m.Update(key(tea.KeyEnter))
	if m.modal != "" || m.insert || m.focus != m.arrangementFocus() {
		t.Fatal("arrangement should be focused in normal mode")
	}
	out, warnings := m.song.Render()
	if len(warnings) != 0 || !strings.Contains(out, "[Chorus]") || !strings.Contains(m.song.Arrangement, "Conversational") {
		t.Fatal("artist arrangement failed to render", warnings)
	}
	if m.song.Style != style || !strings.Contains(m.song.Notes, old) {
		t.Fatal("arrangement lost existing style or previous arrangement")
	}
}
