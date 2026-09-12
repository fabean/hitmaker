package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func defaultConfigDir() string {
	if p := os.Getenv("XDG_CONFIG_HOME"); filepath.IsAbs(p) {
		return filepath.Join(p, "hitmaker")
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "./hitmaker-config"
	}
	return filepath.Join(h, ".config", "hitmaker")
}

// Custom names replace matching defaults; new names are appended in file order.
// Rebuild from defaults on every read so removed presets and overrides disappear.
func loadPresets(path string, defaults []Preset) ([]Preset, error) {
	result := append([]Preset(nil), defaults...)
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(b, &entries); err != nil {
		return result, fmt.Errorf("%s: %w", path, err)
	}
	if entries == nil {
		return result, fmt.Errorf("%s: expected a JSON array", path)
	}
	var problems []error
	for i, entry := range entries {
		var p Preset
		decoder := json.NewDecoder(bytes.NewReader(entry))
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&p)
		p.Name = strings.TrimSpace(p.Name)
		p.Detail = strings.TrimSpace(p.Detail)
		if err == nil && (p.Name == "" || strings.TrimSpace(p.Value) == "") {
			err = fmt.Errorf("name and value must be nonempty strings")
		}
		if err != nil {
			problems = append(problems, fmt.Errorf("%s: entry %d: %w", path, i+1, err))
			continue
		}
		index := -1
		for j, existing := range result {
			if strings.EqualFold(existing.Name, p.Name) {
				index = j
				break
			}
		}
		if index >= 0 {
			result[index] = p
		} else {
			result = append(result, p)
		}
	}
	return result, errors.Join(problems...)
}

func (m *model) openPresetPicker(modal, filename string, defaults []Preset) tea.Cmd {
	presets := defaults
	var err error
	if m.configDir != "" {
		presets, err = loadPresets(filepath.Join(m.configDir, filename), defaults)
	}
	m.choices = nil
	for _, p := range presets {
		m.choices = append(m.choices, choice{p.Name, p.Detail, p.Value})
	}
	cmd := m.openPicker(modal)
	if err != nil {
		m.modalError = "Custom presets: " + err.Error()
	}
	return cmd
}
