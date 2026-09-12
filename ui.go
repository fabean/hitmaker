package main

import (
	"fmt"
	"image/color"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var (
	bg     = lipgloss.Color("#111519")
	fg     = lipgloss.Color("#DAE3DE")
	muted  = lipgloss.Color("#88958F")
	green  = lipgloss.Color("#A4EF83")
	cyan   = lipgloss.Color("#83CEC7")
	purple = lipgloss.Color("#B6A0EA")
	amber  = lipgloss.Color("#E8BC78")
	edge   = lipgloss.Color("#33423F")
	ink    = lipgloss.NewStyle().Foreground(fg)
	dim    = lipgloss.NewStyle().Foreground(muted)
	accent = lipgloss.NewStyle().Foreground(green).Bold(true)
)

type autosaveMsg struct{ revision int }
type choice struct{ title, detail, value string }
type model struct {
	store         Store
	configDir     string
	song          Song
	title         textinput.Model
	style         textarea.Model
	sections      []textarea.Model
	arrangement   textarea.Model
	scratch       textarea.Model
	output        viewport.Model
	help          viewport.Model
	focus         int // title, style, sections..., arrangement, output
	width, height int
	dirty         bool
	revision      int
	status        string
	statusError   bool
	statusNotice  bool
	modal         string
	input         textinput.Model
	choices       []choice
	selection     int
	leftOffset    int
	clearBackups  map[string]string
	modalError    string
	insert        bool
	searching     bool
	leftFocus     int
}

func editor(placeholder string) textarea.Model {
	t := textarea.New()
	t.Prompt = ""
	t.Placeholder = placeholder
	t.ShowLineNumbers = false
	t.EndOfBufferCharacter = ' '
	t.KeyMap.SelectAll.SetKeys("ctrl+a")
	t.KeyMap.LineStart.SetKeys("home")
	t.CharLimit = 0
	t.MaxHeight = 0
	t.MaxWidth = 0
	t.SetVirtualCursor(true)
	s := textarea.DefaultDarkStyles()
	s.Focused.Text = ink
	s.Blurred.Text = ink
	s.Focused.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("#1E3029"))
	s.Blurred.CursorLine = lipgloss.NewStyle()
	s.Focused.Placeholder = dim
	s.Blurred.Placeholder = dim
	s.Focused.Base = lipgloss.NewStyle()
	s.Blurred.Base = lipgloss.NewStyle()
	s.Focused.EndOfBuffer = dim
	s.Blurred.EndOfBuffer = dim
	t.SetStyles(s)
	return t
}
func inputField() textinput.Model {
	t := textinput.New()
	t.Prompt = ""
	t.CharLimit = 120
	t.SetVirtualCursor(true)
	t.SetStyles(textinput.DefaultDarkStyles())
	return t
}
func newModel(store Store, s Song) model {
	m := model{store: store, width: 120, height: 36, status: "Your next song starts here · i to edit · ? for shortcuts"}
	m.help = viewport.New()
	m.help.SoftWrap = true
	m.help.MouseWheelEnabled = false
	m.help.SetContent(helpText)
	m.output = viewport.New()
	m.output.SoftWrap = true
	m.output.MouseWheelEnabled = false
	m.input = inputField()
	m.load(s)
	return m
}
func (m *model) load(s Song) {
	m.song = s
	m.insert = false
	m.searching = false
	m.leftFocus = 0
	m.leftOffset = 0
	m.clearBackups = make(map[string]string)
	m.statusNotice = false
	m.title = inputField()
	m.title.SetValue(s.Title)
	m.style = editor("Genre, mood, instruments, vocals…  Esc then s for presets")
	m.style.DynamicHeight = true
	m.style.MinHeight = 3
	m.style.SetValue(s.Style)
	m.sections = nil
	for _, s := range s.Sections {
		e := editor("Write " + strings.ToLower(s.Name) + " here…")
		e.SetValue(s.Lyrics)
		m.sections = append(m.sections, e)
	}
	m.arrangement = editor("[Intro]\nVerse 1\nChorus x2\n[Outro]")
	m.arrangement.SetValue(s.Arrangement)
	m.scratch = editor("Loose lines, hook ideas, alternate lyrics…")
	m.scratch.SetValue(s.Notes)
	m.focus = 0
	m.dirty = false
	m.revision++
	m.output.GotoTop()
	m.resize()
	m.setFocus(0)
	m.refresh()
}
func (m model) Init() tea.Cmd         { return textinput.Blink }
func (m model) arrangementFocus() int { return 2 + len(m.sections) }
func (m model) outputFocus() int      { return 3 + len(m.sections) }
func (m *model) setFocus(n int) tea.Cmd {
	m.title.Blur()
	m.style.Blur()
	m.arrangement.Blur()
	for i := range m.sections {
		m.sections[i].Blur()
	}
	m.focus = (n + m.outputFocus() + 1) % (m.outputFocus() + 1)
	m.resize()
	if m.focus < m.arrangementFocus() {
		m.leftFocus = m.focus
	}
	if !m.insert {
		return nil
	}
	switch {
	case m.focus == 0:
		return m.title.Focus()
	case m.focus == 1:
		return m.style.Focus()
	case m.focus < m.arrangementFocus():
		return m.sections[m.focus-2].Focus()
	case m.focus == m.arrangementFocus():
		return m.arrangement.Focus()
	}
	return nil
}
func (m *model) syncSong() {
	m.song.Title = m.title.Value()
	m.song.Style = m.style.Value()
	m.song.Arrangement = m.arrangement.Value()
	m.song.Notes = m.scratch.Value()
	for i := range m.sections {
		m.song.Sections[i].Lyrics = m.sections[i].Value()
	}
}
func (m *model) refresh() {
	m.syncSong()
	out, _ := m.song.Render()
	if out == "" {
		out = "Your song will appear here.\n\nWrite a section, then reference its name in the arrangement.\n\nY  copy all lyrics\ny  copy focused field\nb  open Suno"
	}
	m.output.SetContent(out)
}
func (m *model) changed() tea.Cmd {
	m.statusNotice = false
	m.dirty = true
	m.revision++
	m.resize()
	m.refresh()
	r := m.revision
	return tea.Tick(800*time.Millisecond, func(time.Time) tea.Msg { return autosaveMsg{r} })
}
func (m *model) save() bool {
	m.syncSong()
	if err := m.store.Save(&m.song); err != nil {
		m.status = "Save failed: " + err.Error()
		m.statusError = true
		return false
	}
	m.dirty = false
	if !m.statusNotice || m.statusError {
		m.status = "Saved locally · " + m.song.Updated.Format("15:04:05")
		m.statusError = false
		m.statusNotice = false
	}
	return true
}
func (m *model) resize() {
	w := max(20, m.width-4)
	h := max(8, m.height-8)
	if m.width >= 110 {
		w = (m.width-4)/3 - 4
	}
	m.title.SetWidth(w)
	m.style.SetWidth(w)
	// Recalculate the wrapped height, then cap only the viewport. MaxHeight
	// stays unlimited so a small terminal never limits how much text can be entered.
	m.style.SetHeight(min(m.style.Height(), max(3, h-4)))
	for i := range m.sections {
		m.sections[i].SetWidth(w)
		m.sections[i].SetHeight(5)
	}
	m.arrangement.SetWidth(w)
	m.arrangement.SetHeight(max(3, h-6))
	m.output.SetWidth(w)
	m.output.SetHeight(max(3, h-6))
	if m.width < 110 {
		for i := range m.sections {
			m.sections[i].SetHeight(max(3, h-5))
		}
	}
	m.help.SetWidth(max(20, min(78, m.width-10)))
	m.help.SetHeight(max(3, min(27, m.height-12)))
	m.scratch.SetWidth(max(20, min(76, m.width-12)))
	m.scratch.SetHeight(max(3, min(18, m.height-12)))
	m.input.SetWidth(max(12, min(64, m.width-14)))
	if m.width >= 110 {
		total := 6 + m.style.Height() + 4 + 10*len(m.sections)
		m.leftOffset = min(m.leftOffset, max(0, total-h))
		if m.focus < m.arrangementFocus() {
			start, end := m.leftBounds(m.focus)
			if start < m.leftOffset {
				m.leftOffset = start
			}
			if end > m.leftOffset+h {
				m.leftOffset = end - h
			}
		}
	}

}
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()
		m.refresh()
		return m, nil
	case autosaveMsg:
		if msg.revision == m.revision && m.dirty {
			m.save()
		}
		return m, nil
	case actionMsg:
		m.statusNotice = true
		if msg.err != nil {
			m.status = msg.err.Error()
			m.statusError = true
		} else {
			m.status = msg.text
			m.statusError = false
		}
		return m, nil
	case clipboardFallback:
		m.statusNotice = true
		m.status = "Sent " + msg.label + " to terminal clipboard (OSC 52)"
		m.statusError = false
		return m, tea.SetClipboard(msg.value)
	case tea.KeyPressMsg:
		key := msg.String()
		if key == "ctrl+q" || key == "ctrl+c" {
			if !m.dirty || m.save() {
				return m, tea.Quit
			}
			return m, nil
		}
		if m.modal != "" {
			return m.updateModal(msg)
		}
		if key == "esc" {
			return m, m.enterNormal()
		}
		if !m.insert && !strings.HasPrefix(key, "ctrl+") && !(len(key) > 1 && key[0] == 'f') && key != "tab" && key != "shift+tab" {
			return m, m.normalKey(key)
		}

		switch key {
		case "f8":
			return m, m.clearField(false)
		case "f9":
			return m, m.clearField(true)
		case "f1":
			m.modal = "help"
			m.help.GotoTop()
			return m, nil
		case "tab":
			return m, m.setFocus(m.focus + 1)
		case "shift+tab":
			return m, m.setFocus(m.focus - 1)
		case "f2":
			return m, m.setFocus(0)
		case "f3":
			return m, m.setFocus(1)
		case "f4":
			if len(m.sections) > 0 {
				return m, m.setFocus(2)
			}
		case "f5":
			return m, m.setFocus(m.arrangementFocus())
		case "f6":
			return m, m.setFocus(m.outputFocus())
		case "ctrl+s":
			m.save()
			return m, nil
		case "ctrl+n":
			return m, m.openInput("new", "", "Song title or a working idea…")
		case "ctrl+t":
			return m, m.openInput("section", "", "Verse 3, Pre-Chorus, Hook…")
		case "ctrl+e":
			if m.dirty && !m.save() {
				return m, nil
			}
			songs, err := m.store.List()
			if err != nil {
				m.status = err.Error()
				m.statusError = true
			}
			m.choices = nil
			for _, s := range songs {
				m.choices = append(m.choices, choice{s.Title, s.Updated.Format("Jan 02 · 15:04") + "  " + s.Style, s.ID})
			}
			return m, m.openPicker("library")
		case "ctrl+g":
			return m, m.openPresetPicker("genres", "styles.json", genres)
		case "ctrl+l":
			return m, m.openPresetPicker("structures", "arrangements.json", structures)
		case "ctrl+p":
			m.modal = "scratch"
			m.enterNormal()
			return m, nil
		case "ctrl+o":
			out, warnings := m.song.Render()

			if strings.TrimSpace(out) == "" {
				m.status = "Write some lyrics or directions first"
				m.statusError = true
				return m, nil
			}
			label := "lyrics"
			if len(warnings) > 0 {
				label += " · skipped: " + strings.Join(warnings, "; ")
			}
			return m, copyText(out, label)
		case "ctrl+y":
			if strings.TrimSpace(m.song.Style) == "" {
				m.status = "Add a style first · Esc then s for presets"
				return m, nil
			}
			return m, copyText(m.song.Style, "style")
		case "ctrl+b":
			return m, openSuno()
		case "ctrl+x":
			if !m.save() {
				return m, nil
			}
			dir, err := m.store.Export(m.song)
			if err != nil {
				m.status = err.Error()
				m.statusError = true
			} else {
				m.status = "Exported to " + dir
				m.statusError = false
			}
			return m, nil
		}
	}
	if m.modal != "" {
		return m.updateModal(msg)
	}
	if !m.insert {
		return m, nil
	}
	before := m.contentSignature()
	var cmd tea.Cmd
	switch {
	case m.focus == 0:
		m.title, cmd = m.title.Update(msg)
	case m.focus == 1:
		m.style, cmd = m.style.Update(msg)
		m.resize()
	case m.focus < m.arrangementFocus():
		i := m.focus - 2
		m.sections[i], cmd = m.sections[i].Update(msg)
	case m.focus == m.arrangementFocus():
		m.arrangement, cmd = m.arrangement.Update(msg)
	default:
		m.output, cmd = m.output.Update(msg)
	}
	if before != m.contentSignature() {
		return m, tea.Batch(cmd, m.changed())
	}
	return m, cmd
}
func (m model) contentSignature() string {
	v := []string{m.title.Value(), m.style.Value(), m.arrangement.Value(), m.scratch.Value()}
	for _, s := range m.sections {
		v = append(v, s.Value())
	}
	return strings.Join(v, "\x00")
}

func (m *model) openInput(modal, value, placeholder string) tea.Cmd {
	m.enterNormal()
	m.modal = modal
	m.modalError = ""
	m.input.SetValue(value)
	m.input.Placeholder = placeholder
	return m.input.Focus()
}
func (m *model) openPicker(modal string) tea.Cmd {
	m.selection = 0
	placeholder := "Type to filter…"
	if modal == "genres" {
		placeholder = "Search genres, instruments, moods…"
	}
	if modal == "structures" {
		placeholder = "Search genres, bands, arrangements…"
	}
	m.openInput(modal, "", placeholder)
	m.searching = false
	m.input.Blur()
	return nil
}
func (m model) filteredChoices() []choice {
	var out []choice
	q := strings.ToLower(m.input.Value())
	for _, c := range m.choices {
		searchable := strings.ToLower(c.title + " " + c.detail)
		if m.modal == "genres" || m.modal == "structures" {
			searchable += " " + strings.ToLower(c.value)
		}
		match := strings.Contains(searchable, q)
		if m.modal == "genres" || m.modal == "structures" {
			match = true
			for _, word := range strings.Fields(q) {
				if !strings.Contains(searchable, word) {
					match = false
					break
				}
			}
		}
		if match {
			out = append(out, c)
		}
	}
	return out
}
func (m *model) closeModal() tea.Cmd {
	m.modal = ""
	m.insert = false
	m.searching = false
	m.modalError = ""
	m.input.Blur()
	m.scratch.Blur()
	return m.setFocus(m.focus)
}
func (m *model) updateModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		k := key.String()
		if m.modal == "scratch" {
			if k == "esc" {
				if m.insert {
					return m, m.enterNormal()
				}
				return m, m.closeModal()
			}
			if !m.insert {
				switch k {
				case "i", "a", "enter":
					return m, m.enterInsert()
				case "d", "f8":
					return m, m.clearField(false)
				case "u", "f9":
					return m, m.clearField(true)
				case "c":
					cmd := m.clearField(false)
					return m, tea.Batch(cmd, m.enterInsert())
				case "y":
					return m, m.yankField()
				case "q", "m":
					return m, m.closeModal()
				case "w":
					m.save()
					return m, nil
				case "j", "down":
					m.scratch.CursorDown()
				case "k", "up":
					m.scratch.CursorUp()
				case "pgdown":
					m.scratch.PageDown()
				case "pgup":
					m.scratch.PageUp()
				}
				return m, nil
			}
		}
		if m.isPicker() {
			if k == "esc" && m.searching {
				m.searching = false
				m.input.Blur()
				return m, nil
			}
			if !m.searching {
				switch k {
				case "/", "i":
					m.searching = true
					return m, m.input.Focus()
				case "q":
					return m, m.closeModal()
				case "j":
					msg = tea.KeyPressMsg{Code: tea.KeyDown}
				case "k":
					msg = tea.KeyPressMsg{Code: tea.KeyUp}
				case "enter", "esc", "up", "down", "pgup", "pgdown", "ctrl+s":
				default:
					return m, nil
				}
			}
		}
		if m.modal == "help" && (k == "q" || k == "?") {
			return m, m.closeModal()
		}
	}
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "f8", "f9":
			if m.modal == "scratch" {
				return m, m.clearField(key.String() == "f9")
			}
			return m, nil
		case "esc":
			return m, m.closeModal()
		case "f1":
			if m.modal == "help" {
				return m, m.closeModal()
			}
		case "ctrl+s":
			m.save()
			return m, nil
		case "ctrl+p":
			if m.modal == "scratch" {
				return m, m.closeModal()
			}
		case "up", "down", "pgup", "pgdown":
			if m.modal == "library" || m.modal == "genres" || m.modal == "structures" {
				delta := 1
				if key.String() == "pgup" || key.String() == "pgdown" {
					delta = max(1, (m.height-13)/3)
				}
				if key.String() == "up" || key.String() == "pgup" {
					delta = -delta
				}
				m.selection = max(0, min(len(m.filteredChoices())-1, m.selection+delta))
				return m, nil
			}
		case "enter":
			switch m.modal {
			case "help":
				return m, m.closeModal()
			case "new":
				title := strings.TrimSpace(m.input.Value())
				if title == "" {
					m.modalError = "Give your song a working title."
					return m, nil
				}
				if m.dirty && !m.save() {
					m.modalError = m.status
					return m, nil
				}
				m.load(newSong(title))
				m.closeModal()
				m.setFocus(2)
				cmd := m.enterInsert()
				return m, tea.Batch(cmd, m.changed())
			case "section":
				name := strings.TrimSpace(m.input.Value())
				if err := validSectionName(name, m.song.Sections); err != nil {
					m.modalError = err.Error()
					return m, nil
				}
				if len(m.sections) >= 100 {
					m.modalError = "This song already has 100 sections."
					return m, nil
				}
				m.song.Sections = append(m.song.Sections, Section{Name: name})
				m.sections = append(m.sections, editor("Write "+strings.ToLower(name)+" here…"))
				m.closeModal()
				m.setFocus(len(m.sections) + 1)
				cmd := m.enterInsert()
				m.status = "Added " + name + " · use its name in the arrangement"
				return m, tea.Batch(cmd, m.changed())
			case "library", "genres", "structures":
				choices := m.filteredChoices()
				if len(choices) == 0 {
					return m, nil
				}
				c := choices[min(m.selection, len(choices)-1)]
				switch m.modal {
				case "library":
					if m.dirty && !m.save() {
						m.modalError = m.status
						return m, nil
					}
					s, err := m.store.Load(c.value)
					if err != nil {
						m.modalError = err.Error()
						return m, nil
					}
					m.load(s)
					m.status = "Opened " + s.Title
					return m, m.closeModal()
				case "genres":
					value := strings.TrimSpace(m.style.Value())
					if value != "" {
						value += ", "
					}
					m.style.SetValue(value + c.value)
					m.closeModal()
					cmd := m.setFocus(1)
					return m, tea.Batch(cmd, m.changed())
				case "structures":
					// Keep the previous arrangement in notes so a preset never loses work.
					old := m.arrangement.Value()
					if old != "" && old != c.value {
						m.scratch.SetValue(strings.TrimSpace(m.scratch.Value() + "\n\nPrevious arrangement:\n" + old))
					}
					m.arrangement.SetValue(c.value)
					m.closeModal()
					cmd := m.setFocus(m.arrangementFocus())
					return m, tea.Batch(cmd, m.changed())
				}
			}
		}
	}
	if m.isPicker() && !m.searching {
		return m, nil
	}
	if m.modal == "scratch" && !m.insert {
		return m, nil
	}
	var cmd tea.Cmd
	if m.modal == "scratch" {
		before := m.scratch.Value()
		m.scratch, cmd = m.scratch.Update(msg)
		if before != m.scratch.Value() {
			return m, tea.Batch(cmd, m.changed())
		}
	} else if m.modal == "help" {
		m.help, cmd = m.help.Update(msg)
	} else {
		before := m.input.Value()
		m.input, cmd = m.input.Update(msg)
		if before != m.input.Value() {
			m.selection = 0
			m.modalError = ""
		}
	}
	return m, cmd
}

func panel(title, body string, w, h int, active bool, tint color.Color) string {
	border := edge
	if active {
		border = tint
	}
	heading := lipgloss.NewStyle().Foreground(tint).Bold(true).Render(ansi.Truncate(title, w-4, "…"))
	if active {
		heading = lipgloss.NewStyle().Foreground(tint).Bold(true).Render(ansi.Truncate("● "+title, w-4, "…"))
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border).Padding(0, 1).Width(w).Height(h).MaxHeight(h).Render(heading + "\n\n" + body)
}

// leftBounds measures full cards, including the gap before the next card.
func (m model) leftBounds(focus int) (int, int) {
	if focus == 0 {
		return 0, 5
	}
	styleEnd := 6 + m.style.Height() + 4
	if focus == 1 {
		return 6, styleEnd
	}
	start := styleEnd + 1 + (focus-2)*10
	return start, start + 9
}

func (m model) editPanel(title, body string, w, h int, active bool, tint color.Color) string {
	if active {
		suffix := " · d clear"
		if m.insert {
			suffix = " · INSERT"
		}
		title = ansi.Truncate(title, max(1, w-17), "…") + suffix
	}
	return panel(title, body, w, h, active, tint)
}

func (m model) leftPane(w, h int) string {
	cards := []string{
		m.editPanel("SONG", m.title.View(), w, 5, m.focus == 0, amber),
		m.editPanel("STYLE", m.style.View(), w, m.style.Height()+4, m.focus == 1, cyan),
	}
	for i := range m.sections {
		label := strings.ToUpper(m.song.Sections[i].Name) + fmt.Sprintf("  %d/%d", i+1, len(m.sections))
		cards = append(cards, m.editPanel(label, m.sections[i].View(), w, 9, m.focus == i+2, green))
	}
	lines := strings.Split(strings.Join(cards, "\n\n"), "\n")
	start := min(m.leftOffset, max(0, len(lines)-h))
	return strings.Join(lines[start:min(len(lines), start+h)], "\n")
}

func (m model) View() tea.View {
	w, h := m.width, m.height
	if w < 48 || h < 18 {
		v := tea.NewView("hitmaker\n\nResize to at least 48 × 18.\nEsc then q saves and quits.")
		v.AltScreen = true
		return v
	}
	saved := "saved"
	if m.song.Updated.IsZero() {
		saved = "new draft"
	}
	if m.dirty {
		saved = "saving…"
	}
	brand := lipgloss.NewStyle().Foreground(bg).Background(green).Bold(true).Padding(0, 1).Render("hitmaker")
	top := brand + dim.Render("  /  ") + ink.Bold(true).Render(m.song.Title)
	right := dim.Render(m.modeName() + "  ·  " + saved)
	top = lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(max(1, w-lipgloss.Width(right)-2)).MaxWidth(max(1, w-lipgloss.Width(right)-2)).Render(top), right)
	shortcuts := "n new   e songs   t section   s styles   r structure   m scratchpad"
	if m.insert {
		shortcuts = "INSERT · type lyrics freely · Esc returns to normal mode"
	}
	if w < 80 {
		shortcuts = "NORMAL · n new  s styles  ? help"
		if m.insert {
			shortcuts = "INSERT · Esc returns to normal mode"
		}
	}
	header := top + "\n" + dim.Render(shortcuts) + "\n" + dim.Render(strings.Repeat("─", w)) + "\n"
	bodyH := h - 8
	var body string
	fieldHint := "i edits · j/k fields · h/l panes · y copies · d clears"
	if m.insert {
		fieldHint = "Type your song title. Esc returns to normal mode."
	}
	if w >= 110 {
		col := (w - 4) / 3
		left := lipgloss.NewStyle().Width(col).Height(bodyH).MaxHeight(bodyH).Render(m.leftPane(col, bodyH))
		middle := m.editPanel("ARRANGEMENT", m.arrangement.View()+"\n"+dim.Render(ansi.Truncate("Section name · Chorus x2 · [direction]", col-4, "…")), col, bodyH, m.focus == m.arrangementFocus(), cyan)
		out, _ := m.song.Render()
		right := panel(fmt.Sprintf("SUNO OUTPUT  ·  %d chars", utf8.RuneCountInString(out)), m.output.View()+"\n"+dim.Render(ansi.Truncate(fmt.Sprintf("%d%%  ·  y copy lyrics", int(m.output.ScrollPercent()*100)), col-4, "…")), col, bodyH, m.focus == m.outputFocus(), purple)
		body = lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", middle, "  ", right)
	} else {
		switch {
		case m.focus == 0:
			body = m.editPanel("SONG", m.title.View()+"\n\n"+dim.Render(fieldHint), w, bodyH, true, amber)
		case m.focus == 1:
			body = m.editPanel("STYLE", m.style.View(), w, bodyH, true, cyan)
		case m.focus < m.arrangementFocus():
			i := m.focus - 2
			body = m.editPanel(strings.ToUpper(m.song.Sections[i].Name)+"", m.sections[i].View(), w, bodyH, true, green)
		case m.focus == m.arrangementFocus():
			body = m.editPanel("ARRANGEMENT", m.arrangement.View(), w, bodyH, true, cyan)
		default:
			body = panel("SUNO OUTPUT", m.output.View(), w, bodyH, true, purple)
		}
	}
	_, warnings := m.song.Render()
	status := m.status
	statusStyle := dim
	if m.statusError {
		statusStyle = lipgloss.NewStyle().Foreground(amber)
	} else if len(warnings) > 0 && !m.statusNotice {
		status = strings.Join(warnings, " · ")
		statusStyle = lipgloss.NewStyle().Foreground(amber)
	}
	footer := statusStyle.Width(w).MaxHeight(1).Render(status) + "\n" + dim.Render("h/l panes · j/k fields · i edit · d clear · y copy · u restore · w save · q quit")
	if w < 80 {
		footer = statusStyle.Width(w).MaxHeight(1).Render(status) + "\n" + dim.Render("i edit · d clear · y copy · ? help · q quit")
	}
	if m.insert {
		footer = statusStyle.Width(w).MaxHeight(1).Render(status) + "\n" + dim.Render("INSERT · Esc normal · Tab next field")
	}
	content := header + body + "\n" + footer
	if m.modal != "" {
		content = m.modalView()
	}
	v := tea.NewView(lipgloss.NewStyle().Foreground(fg).Width(w).Height(h).MaxWidth(w).MaxHeight(h).Render(content))
	v.AltScreen = true
	v.BackgroundColor = bg
	v.ForegroundColor = fg
	v.WindowTitle = "hitmaker · " + m.song.Title
	return v
}

func (m model) modalView() string {
	w := min(86, m.width-4)
	var title, body, hint string
	switch m.modal {
	case "help":
		title = "HITMAKER / SHORTCUTS"
		body = m.help.View()
		hint = "j/k scroll · Esc or q closes"
	case "scratch":
		title = "SCRATCHPAD · " + m.modeName()
		body = m.scratch.View()
		hint = "i edit · d clear · y copy · u restore · q back"
		if m.insert {
			hint = "INSERT · Esc returns to normal mode"
		}
	case "new":
		title = "NEW SONG"
		body = dim.Render("What are you writing about?") + "\n\n" + m.input.View()
		hint = "Enter creates your song · Esc cancels"
	case "section":
		title = "NEW SECTION"
		body = dim.Render("Give this reusable section a name.") + "\n\n" + m.input.View()
		hint = "Enter adds it · Reference its name in the arrangement"
	default:
		switch m.modal {
		case "library":
			title = "YOUR SONGS"
			hint = "j/k choose · / search · Enter opens · q back"
		case "genres":
			title = fmt.Sprintf("STYLE PRESETS · %d / %d", len(m.filteredChoices()), len(m.choices))
			hint = "j/k or PgUp/PgDn · / search · Enter appends"
		case "structures":
			title = fmt.Sprintf("ARRANGEMENTS · %d / %d", len(m.filteredChoices()), len(m.choices))
			hint = "j/k choose · / search · Enter applies · q back"
		}
		if m.searching {
			hint = "SEARCH · type to filter · Enter chooses · Esc stops searching"
		}
		body = m.input.View() + "\n\n"
		choices := m.filteredChoices()
		if len(choices) == 0 {
			body += dim.Render("No songs or presets found.")
		}
		count := max(1, (m.height-13)/3)
		start := max(0, m.selection-count+1)
		for i := start; i < min(len(choices), start+count); i++ {
			c := choices[i]
			prefix := "  "
			style := ink
			if i == m.selection {
				prefix = "› "
				style = accent
			}
			body += style.Render(prefix+ansi.Truncate(c.title, w-8, "…")) + "\n" + dim.Width(w-8).MaxHeight(1).Render("  "+c.detail) + "\n\n"
		}
	}
	if m.modalError != "" {
		body += "\n\n" + lipgloss.NewStyle().Foreground(amber).Render(m.modalError)
	}
	inner := accent.Render(title) + "\n\n" + body + "\n\n" + dim.Render(hint)
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(green).Padding(1, 2).Width(w).MaxHeight(m.height - 2).Render(inner)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

const helpText = `NORMAL MODE
h / l           Previous / next pane
j / k           Next / previous field; scroll output
Tab / Shift+Tab Next / previous field across all panes
i / a / Enter   Edit the focused field
Esc             Leave insert mode
d               Clear the entire focused field
y               Copy the focused field or output
u               Restore a cleared field if still empty
c               Clear the field and start editing
Y               Copy the full Suno lyrics from any pane

SONGS & TOOLS
n / e           New song / saved songs
t               Add a lyric section
s / r           Style presets / song structures
m               Scratchpad
w               Save (also automatic after edits)
b               Open Suno
x               Export lyrics, style, and title
q               Save and quit
?               Help

PICKERS
j / k           Choose a row
/               Start searching
Esc             Stop searching; again to close
Enter           Choose the current result
q               Close while not searching

ARRANGE
Verse 1         Insert a named section
Chorus x4       Repeat a section (1–16)
[Quiet intro]   Keep a performance direction

In INSERT mode, letters are text, not commands.
d clears a whole field; this is not a full Vim editor.`
