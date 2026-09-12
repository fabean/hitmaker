package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestSaveCreatesMarkdownSnapshots(t *testing.T) {
	st := Store{t.TempDir()}
	s := demoSong()
	s.Title = "What a Time to Be Alive"
	s.Notes = "Private scratchpad"
	if err := st.Save(&s); err != nil {
		t.Fatal(err)
	}
	first := st.MarkdownPath(s)
	name := filepath.Base(first)
	if !strings.HasPrefix(name, s.Updated.Format("2006-01-02_15-04-05")) || !strings.Contains(name, "what-a-time-to-be-alive") || filepath.Ext(name) != ".md" {
		t.Fatal("missing timestamp or title", name)
	}
	b, err := os.ReadFile(first)
	if err != nil {
		t.Fatal(err)
	}
	lyrics, _ := s.Render()
	if !strings.Contains(string(b), fencedText(lyrics)) || !strings.Contains(string(b), s.Style) || !strings.Contains(string(b), "# What a Time to Be Alive") {
		t.Fatal("snapshot does not contain title, style, and rendered output")
	}
	if strings.Contains(string(b), s.Notes) {
		t.Fatal("snapshot includes private notes")
	}
	s.Sections[0].Lyrics = "Changed hook"
	if err := st.Save(&s); err != nil {
		t.Fatal(err)
	}
	second := st.MarkdownPath(s)
	if first == second {
		t.Fatal("successive saves reused a snapshot filename")
	}
	old, err := os.ReadFile(first)
	if err != nil || string(old) != string(b) {
		t.Fatal("previous snapshot changed", err)
	}
	current, err := os.ReadFile(second)
	if err != nil || !strings.Contains(string(current), "Changed hook") {
		t.Fatal("new snapshot not written", err)
	}
	info, err := os.Stat(second)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("snapshot must be private", err)
	}
}

func TestMarkdownForUnfinishedAndUnusualText(t *testing.T) {
	st := Store{t.TempDir()}
	s := newSong("../夜 / # song\n<script>")
	s.Style = "a ``` fence"
	s.Sections[0].Lyrics = "[literal]\n````\n# lyric heading"
	s.Arrangement = "Chorus x2\nVerse 2"
	if err := st.Save(&s); err != nil {
		t.Fatal("unfinished draft must still save", err)
	}
	p := st.MarkdownPath(s)
	if filepath.Dir(p) != filepath.Join(st.Dir, "exports") {
		t.Fatal("title escaped export directory", p)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !strings.Contains(text, "## Skipped sections\n\n- Verse 2 is empty") || strings.Count(text, "# lyric heading") != 2 {
		t.Fatal("missing output or warnings", text)
	}
	if !strings.Contains(text, "`````text\n") || strings.Contains(text, "<script>") {
		t.Fatal("Markdown formatting not escaped", text)
	}
	for _, title := range []string{"", "../../", strings.Repeat("夜", 200)} {
		slug := songSlug(title)
		if slug == "" || len(slug) > 80 || !utf8.ValidString(slug) || strings.ContainsAny(slug, "/\\.") {
			t.Fatal("unsafe filename", slug)
		}
	}
	s.Updated = time.Date(2026, 9, 12, 15, 4, 5, 42, time.FixedZone("EDT", -4*60*60))
	if !strings.Contains(st.MarkdownPath(s), "2026-09-12_15-04-05.000000042-0400") {
		t.Fatal("missing precise local timestamp")
	}
}

func TestMarkdownFailureKeepsJSONAndReportsError(t *testing.T) {
	st := Store{t.TempDir()}
	s := demoSong()
	if err := os.WriteFile(filepath.Join(st.Dir, "exports"), []byte("blocked"), 0600); err != nil {
		t.Fatal(err)
	}
	err := st.Save(&s)
	if err == nil || !strings.Contains(err.Error(), "Markdown snapshot failed") {
		t.Fatal("snapshot error not reported", err)
	}
	loaded, err := st.Load(s.ID)
	if err != nil || loaded.Title != s.Title {
		t.Fatal("editable song was not preserved", err)
	}
}
