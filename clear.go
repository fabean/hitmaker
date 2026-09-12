package main

import tea "charm.land/bubbletea/v2"

// Field names remain stable when a new section changes the tab order.
func (m *model) editableField() (string, string, func(string)) {
	if m.modal == "scratch" {
		return "scratchpad", m.scratch.Value(), m.scratch.SetValue
	}
	if m.modal != "" {
		return "", "", nil
	}
	switch {
	case m.focus == 0:
		return "title", m.title.Value(), m.title.SetValue
	case m.focus == 1:
		return "style", m.style.Value(), m.style.SetValue
	case m.focus < m.arrangementFocus():
		i := m.focus - 2
		return "section: " + m.song.Sections[i].Name, m.sections[i].Value(), m.sections[i].SetValue
	case m.focus == m.arrangementFocus():
		return "arrangement", m.arrangement.Value(), m.arrangement.SetValue
	default:
		return "", "", nil
	}
}

func (m *model) clearField(restore bool) tea.Cmd {
	name, value, set := m.editableField()
	if set == nil {
		m.status = "Output is generated · focus a field or lyric section to clear it"
		m.statusError = false
		m.statusNotice = true
		return nil
	}
	if restore {
		old, ok := m.clearBackups[name]
		if !ok || value != "" {
			m.status = "Restore is available while this field is empty after d"
			m.statusError = false
			m.statusNotice = true
			return nil
		}
		set(old)
		delete(m.clearBackups, name)
	} else {
		if value == "" {
			return nil
		} // A second clear must not discard the backup.
		m.clearBackups[name] = value
		set("")
	}
	cmd := m.changed()
	m.status = "Cleared " + name + " · u restores it"
	if restore {
		m.status = "Restored " + name
	}
	m.statusError = false
	m.statusNotice = true
	return cmd
}
