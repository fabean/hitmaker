package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// MarkdownPath uses the save's local time and offset, plus a stable song ID.
// Nanosecond precision preserves separate snapshots from rapid successive saves.
func (st Store) MarkdownPath(s Song) string {
	name := s.Updated.Format("2006-01-02_15-04-05.000000000-0700") + "_" + songSlug(s.Title) + "_" + s.ID + ".md"
	return filepath.Join(st.Dir, "exports", name)
}

func songSlug(title string) string {
	var b strings.Builder
	separator := false
	for _, r := range strings.ToLower(title) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if b.Len()+len(string(r))+1 > 80 {
				break
			}
			if separator && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			separator = false
		} else {
			separator = true
		}
	}
	if b.Len() == 0 {
		return "untitled-song"
	}
	return b.String()
}

func markdownText(value string) string {
	// Single-line headings and list items cannot inject Markdown blocks.
	value = strings.Join(strings.Fields(value), " ")
	return strings.NewReplacer("\\", "\\\\", "`", "\\`", "*", "\\*", "_", "\\_", "[", "\\[", "]", "\\]", "<", "&lt;", ">", "&gt;", "#", "\\#").Replace(value)
}

func fencedText(value string) string {
	longest, run := 0, 0
	for _, r := range value {
		if r == '`' {
			run++
			longest = max(longest, run)
		} else {
			run = 0
		}
	}
	fence := strings.Repeat("`", max(3, longest+1))
	return fence + "text\n" + value + "\n" + fence + "\n"
}

// Markdown contains the same rendered lyrics as the output pane, including
// unfinished drafts. Scratchpad notes remain only in the editable JSON file.
func (s Song) Markdown() string {
	title := markdownText(s.Title)
	if title == "" {
		title = "Untitled song"
	}
	lyrics, warnings := s.Render()
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\nSaved: %s\n\n## Style\n\n%s\n## Lyrics\n\n%s", title, s.Updated.Format(time.RFC3339Nano), fencedText(s.Style), fencedText(lyrics))
	if len(warnings) > 0 {
		b.WriteString("\n## Skipped sections\n\n")
		for _, warning := range warnings {
			fmt.Fprintf(&b, "- %s\n", markdownText(warning))
		}
	}
	return b.String()
}
