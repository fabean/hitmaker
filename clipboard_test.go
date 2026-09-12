package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestCopyIncompleteSong(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX clipboard stub")
	}
	dir := t.TempDir()
	helper := "wl-copy"
	if runtime.GOOS == "darwin" {
		helper = "pbcopy"
	}
	if err := os.WriteFile(filepath.Join(dir, helper), []byte("#!/bin/sh\n/bin/cat > \"$HITMAKER_TEST_CLIPBOARD\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, "copied.txt")
	t.Setenv("PATH", dir)
	t.Setenv("HITMAKER_TEST_CLIPBOARD", dest)
	song := demoSong()
	song.Sections[2].Lyrics = ""
	m := newModel(Store{t.TempDir()}, song)
	m.changed()
	_, cmd := m.Update(ctrl('o'))
	if cmd == nil {
		t.Fatal("empty referenced verse blocked copying")
	}
	result := cmd()
	if _, ok := result.(actionMsg); !ok {
		t.Fatalf("clipboard command failed: %#v", result)
	}
	m.Update(result)
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := song.Render()
	if string(got) != want || !strings.Contains(string(got), song.Sections[0].Lyrics) {
		t.Fatal("did not copy visible output")
	}
	if !strings.Contains(m.status, "Copied lyrics") || !strings.Contains(m.status, "Verse 2 is empty") {
		t.Fatal("missing copy confirmation or skipped-section warning", m.status)
	}
	m.Update(autosaveMsg{m.revision})
	if !strings.Contains(ansi.Strip(m.View().Content), "Copied lyrics") {
		t.Fatal("copy confirmation hidden by warnings or autosave")
	}
	if m.song.Sections[2].Lyrics != "" {
		t.Fatal("copy changed the draft")
	}
}

func TestCopyEmptyOutput(t *testing.T) {
	song := newSong("empty")
	song.Arrangement = "Chorus"
	m := newModel(Store{t.TempDir()}, song)
	_, cmd := m.Update(ctrl('o'))
	if cmd != nil || !m.statusError {
		t.Fatal("attempted to replace clipboard with empty output")
	}
}

func TestCopyIncompleteSongFallback(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	song := demoSong()
	song.Sections[2].Lyrics = ""
	m := newModel(Store{t.TempDir()}, song)
	_, cmd := m.Update(ctrl('o'))
	if cmd == nil {
		t.Fatal("copy blocked")
	}
	result, ok := cmd().(clipboardFallback)
	if !ok {
		t.Fatal("expected terminal clipboard fallback")
	}
	want, _ := song.Render()
	if result.value != want {
		t.Fatal("fallback did not copy preview")
	}
	_, fallback := m.Update(result)
	if fallback == nil || !m.statusNotice {
		t.Fatal("missing terminal clipboard command or visible status")
	}
}
