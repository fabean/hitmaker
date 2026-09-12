package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func writePresetFile(t *testing.T, dir, name string, presets []Preset) {
	t.Helper()
	b, err := json.Marshal(presets)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), b, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestCustomPresetsMergeAndValidate(t *testing.T) {
	defaults := []Preset{{Name: "Original", Value: "default"}}
	path := filepath.Join(t.TempDir(), "styles.json")
	got, err := loadPresets(path, defaults)
	if err != nil || !reflect.DeepEqual(got, defaults) {
		t.Fatalf("missing optional file: %v, %v", got, err)
	}
	body := `[
		{"name":" original ","value":"override"},
		{"name":"","value":"invalid"},
		{"name":"New","detail":"searchable","value":"first"},
		{"name":"Bad key","value":"unused","typo":true},
		{"name":"NEW","value":"last wins"},
		{"name":"Blank","value":"  "}
	]`
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	got, err = loadPresets(path, defaults)
	if err == nil || !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "entry 4") {
		t.Fatalf("missing actionable warning: %v", err)
	}
	if len(got) != 2 || got[0].Value != "override" || got[1].Value != "last wins" {
		t.Fatalf("bad merge: %#v", got)
	}
	if defaults[0].Value != "default" {
		t.Fatal("custom presets mutated built-ins")
	}
	for _, invalid := range []string{`{`, `{}`, `null`, `[false]`} {
		if err := os.WriteFile(path, []byte(invalid), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := loadPresets(path, defaults)
		if err == nil || !reflect.DeepEqual(got, defaults) {
			t.Fatalf("invalid file should preserve defaults: %q, %v, %v", invalid, got, err)
		}
	}
}

func TestCustomPresetPickerReloadAndApply(t *testing.T) {
	m := newModel(Store{t.TempDir()}, demoSong())
	m.configDir = t.TempDir()
	style := Preset{Name: "Personal style", Detail: "copper bells", Value: "bright custom prompt"}
	writePresetFile(t, m.configDir, "styles.json", []Preset{style})
	m.Update(runeKey('s'))
	pasteText(&m, "copper bells")
	if got := m.filteredChoices(); len(got) != 1 || got[0].value != style.Value {
		t.Fatalf("custom style unavailable in search: %#v", got)
	}
	m.Update(key(tea.KeyEnter))
	if !strings.Contains(m.song.Style, style.Value) {
		t.Fatal("custom style was not applied")
	}
	style.Value = "changed on disk"
	writePresetFile(t, m.configDir, "styles.json", []Preset{style})
	m.Update(runeKey('s'))
	pasteText(&m, "personal style")
	if got := m.filteredChoices(); len(got) != 1 || got[0].value != style.Value {
		t.Fatalf("picker did not reload: %#v", got)
	}
	m.Update(key(tea.KeyEscape))
	m.Update(key(tea.KeyEscape))
	if err := os.Remove(filepath.Join(m.configDir, "styles.json")); err != nil {
		t.Fatal(err)
	}
	m.Update(runeKey('s'))
	if len(m.choices) != len(genres) {
		t.Fatal("removed presets remained in picker")
	}
	m.Update(key(tea.KeyEscape))
	before := m.song.Arrangement
	arrangement := Preset{Name: "Personal arrangement", Value: "[Custom intro]\nChorus x2\n[End]"}
	writePresetFile(t, m.configDir, "arrangements.json", []Preset{arrangement})
	m.Update(runeKey('r'))
	pasteText(&m, "personal arrangement")
	m.Update(key(tea.KeyEnter))
	if m.song.Arrangement != arrangement.Value || !strings.Contains(m.song.Notes, before) {
		t.Fatal("custom arrangement not applied or previous arrangement lost")
	}
	output, warnings := m.song.Render()
	if len(warnings) != 0 || !strings.Contains(output, "[Custom intro]") || strings.Count(output, m.song.Sections[0].Lyrics) != 2 {
		t.Fatalf("custom arrangement did not render: %q, %v", output, warnings)
	}
	if err := os.WriteFile(filepath.Join(m.configDir, "styles.json"), []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	m.Update(runeKey('s'))
	if m.modalError == "" || len(m.choices) != len(genres) {
		t.Fatal("broken configuration should show a warning and retain built-ins")
	}
}

func TestConfigDirectoryAndExamples(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if got := defaultConfigDir(); got != filepath.Join(dir, "hitmaker") {
		t.Fatal("XDG config directory ignored", got)
	}
	t.Setenv("XDG_CONFIG_HOME", "relative-path")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if got := defaultConfigDir(); got != filepath.Join(home, ".config", "hitmaker") {
		t.Fatal("relative XDG path must fall back to home", got)
	}
	for _, file := range []string{"styles.json", "arrangements.json"} {
		presets, err := loadPresets(filepath.Join("examples", "config", file), nil)
		if err != nil || len(presets) == 0 {
			t.Fatalf("invalid example %s: %v", file, err)
		}
		if file == "arrangements.json" {
			song := demoSong()
			for _, preset := range presets {
				song.Arrangement = preset.Value
				if _, warnings := song.Render(); len(warnings) != 0 {
					t.Fatal("example references missing sections", warnings)
				}
			}
		}
	}
}
