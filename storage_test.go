package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSaveLoadExport(t *testing.T) {
	st := Store{t.TempDir()}
	s := demoSong()
	s.Title = "夜 / ../ song"
	s.Notes = "private draft"
	if err := st.Save(&s); err != nil {
		t.Fatal(err)
	}
	got, err := st.Load(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != s.Title || got.Notes != s.Notes || !reflect.DeepEqual(got.Sections, s.Sections) || !got.Updated.Equal(s.Updated) {
		t.Fatal("roundtrip mismatch")
	}
	// A rename updates the same file, not a new copy.
	s.Title = "renamed"
	if err := st.Save(&s); err != nil {
		t.Fatal(err)
	}
	all, err := st.List()
	if err != nil || len(all) != 1 || all[0].Title != "renamed" {
		t.Fatal(all, err)
	}
	dir, err := st.Export(s)
	if err != nil {
		t.Fatal(err)
	}
	lyrics, err := os.ReadFile(filepath.Join(dir, "lyrics.txt"))
	if err != nil {
		t.Fatal(err)
	}
	want, _ := s.Render()
	if string(lyrics) != want+"\n" {
		t.Fatal("wrong export")
	}
	if info, err := os.Stat(filepath.Join(st.Dir, "songs", s.ID+".json")); err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("song must be private", err)
	}
}
func TestStorageFailures(t *testing.T) {
	st := Store{t.TempDir()}
	if _, err := st.Load("../../etc/passwd"); err == nil {
		t.Fatal("accepted traversal")
	}
	bad := demoSong()
	bad.ID = "../outside"
	if err := st.Save(&bad); err == nil {
		t.Fatal("accepted unsafe id")
	}
	s := newSong("empty")
	if _, err := st.Export(s); err == nil {
		t.Fatal("exported unresolved sections")
	}
	s = demoSong()
	if err := st.Save(&s); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(st.Dir, "songs", "broken.json"), []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	songs, err := st.List()
	if len(songs) != 1 || err == nil {
		t.Fatal("corrupt file should be reported without hiding healthy songs")
	}
	if err := os.WriteFile(filepath.Join(st.Dir, "blocker"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	blocked := Store{filepath.Join(st.Dir, "blocker", "child")}
	if err := blocked.Save(&s); err == nil {
		t.Fatal("save failure swallowed")
	}
}
func TestExportDoesNotIncludeNotes(t *testing.T) {
	st := Store{t.TempDir()}
	s := demoSong()
	s.Notes = "PRIVATE"
	dir, err := st.Export(s)
	if err != nil {
		t.Fatal(err)
	}
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		b, _ := os.ReadFile(filepath.Join(dir, f.Name()))
		if strings.Contains(string(b), "PRIVATE") {
			t.Fatal("notes leaked")
		}
	}
}
