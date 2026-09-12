package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Store struct{ Dir string }

var idRE = regexp.MustCompile(`^[a-f0-9]{24}$`)

func defaultDataDir() string {
	if p := os.Getenv("XDG_DATA_HOME"); filepath.IsAbs(p) {
		return filepath.Join(p, "hitmaker")
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "./hitmaker-data"
	}
	return filepath.Join(h, ".local", "share", "hitmaker")
}
func (st Store) Save(s *Song) error {
	if !idRE.MatchString(s.ID) {
		return fmt.Errorf("invalid song ID")
	}
	copy := *s
	copy.Updated = time.Now()
	copy.Version = 1
	b, err := json.MarshalIndent(copy, "", "  ")
	if err != nil {
		return err
	}
	if err = atomicWrite(filepath.Join(st.Dir, "songs", s.ID+".json"), append(b, '\n')); err != nil {
		return err
	}
	if err = atomicWrite(st.MarkdownPath(copy), []byte(copy.Markdown())); err != nil {
		return fmt.Errorf("song JSON saved, but Markdown snapshot failed: %w", err)
	}
	s.Updated = copy.Updated
	return nil
}
func atomicWrite(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".hitmaker-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
func (st Store) Load(id string) (Song, error) {
	var s Song
	if !idRE.MatchString(id) {
		return s, fmt.Errorf("invalid song ID")
	}
	b, err := os.ReadFile(filepath.Join(st.Dir, "songs", id+".json"))
	if err != nil {
		return s, err
	}
	if err = json.Unmarshal(b, &s); err != nil {
		return s, err
	}
	if s.ID != id || s.Version != 1 {
		return s, fmt.Errorf("unsupported or invalid song file %s", id)
	}
	if len(s.Sections) > 100 {
		return s, fmt.Errorf("too many sections in %s", id)
	}
	for i, sec := range s.Sections {
		if err = validSectionName(sec.Name, s.Sections[:i]); err != nil {
			return s, err
		}
	}
	return s, nil
}
func (st Store) List() ([]Song, error) {
	entries, err := os.ReadDir(filepath.Join(st.Dir, "songs"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Song
	var problems []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		s, err := st.Load(strings.TrimSuffix(e.Name(), ".json"))
		if err != nil {
			problems = append(problems, e.Name()+": "+err.Error())
			continue
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Updated.After(out[j].Updated) })
	if len(problems) > 0 {
		return out, fmt.Errorf("could not read %d song file(s): %s", len(problems), strings.Join(problems, "; "))
	}
	return out, nil
}
func (st Store) Export(s Song) (string, error) {
	lyrics, warnings := s.Render()
	if len(warnings) > 0 {
		return "", fmt.Errorf("fix arrangement before export: %s", strings.Join(warnings, "; "))
	}
	if strings.TrimSpace(lyrics) == "" {
		return "", fmt.Errorf("write some lyrics or directions first")
	}
	if !idRE.MatchString(s.ID) {
		return "", fmt.Errorf("invalid song ID")
	}
	dir := filepath.Join(st.Dir, "exports", s.ID)
	for _, item := range []struct{ name, body string }{{"lyrics.txt", lyrics}, {"style.txt", s.Style}, {"title.txt", s.Title}} {
		if err := atomicWrite(filepath.Join(dir, item.name), []byte(item.body+"\n")); err != nil {
			return "", err
		}
	}
	return dir, nil
}
