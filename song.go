package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Section struct {
	Name   string `json:"name"`
	Lyrics string `json:"lyrics"`
}
type Song struct {
	Version     int       `json:"version"`
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Style       string    `json:"style"`
	Sections    []Section `json:"sections"`
	Arrangement string    `json:"arrangement"`
	Notes       string    `json:"notes"`
	Updated     time.Time `json:"updated"`
}

type Preset struct {
	Name   string `json:"name"`
	Detail string `json:"detail,omitempty"`
	Value  string `json:"value"`
}

func newSong(title string) Song {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return Song{Version: 1, ID: hex.EncodeToString(b), Title: title, Sections: []Section{{Name: "Chorus"}, {Name: "Verse 1"}, {Name: "Verse 2"}, {Name: "Bridge"}}, Arrangement: structures[0].Value}
}

func demoSong() Song {
	s := newSong("Last light on the radio")
	s.Style = genres[1].Value
	s.Sections[0].Lyrics = "We are the last light on the radio\nA little louder when the city goes\nIf all we have is one more song\nTurn it up and take the long way home"
	s.Sections[1].Lyrics = "Coffee cold at a quarter to two\nDashboard stars and a borrowed moon\nYou draw a map on the window steam\nTo somewhere we have never been"
	s.Sections[2].Lyrics = "Every exit is a chance to stay\nEvery mile takes the noise away\nYou find a station through the static snow\nA melody we used to know"
	s.Sections[3].Lyrics = "Let the morning wait outside\nWe have still got road tonight"
	s.Notes = "Try a stripped-back second verse.\nFinal chorus: lift the harmony, add a counter-melody."
	return s
}

var repeatRE = regexp.MustCompile(`(?i)^(.+?)\s+x(\d+)$`)
var sectionRE = regexp.MustCompile(`(?i)^(verse|chorus|bridge|pre-chorus|post-chorus|hook)(\s+\d+)?$`)

// Render expands named sections once, preserving free text and bracketed directions.
// References are case-insensitive; repetition is bounded to avoid runaway output.
func (s Song) Render() (string, []string) {
	lookup := map[string]Section{}
	for _, sec := range s.Sections {
		lookup[strings.ToLower(strings.TrimSpace(sec.Name))] = sec
	}
	var blocks, warnings []string
	for i, raw := range strings.Split(strings.ReplaceAll(s.Arrangement, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		name, count := line, 1
		if match := repeatRE.FindStringSubmatch(line); match != nil {
			name = strings.TrimSpace(match[1])
			n, err := strconv.Atoi(match[2])
			if err != nil || n < 1 || n > 16 {
				warnings = append(warnings, fmt.Sprintf("Line %d: repeat must be 1–16", i+1))
				continue
			}
			count = n
		}
		if sec, ok := lookup[strings.ToLower(name)]; ok {
			if strings.TrimSpace(sec.Lyrics) == "" {
				warnings = append(warnings, sec.Name+" is empty")
				continue
			}
			for j := 0; j < count; j++ {
				blocks = append(blocks, "["+sec.Name+"]\n"+strings.TrimSpace(sec.Lyrics))
			}
		} else if name != line || sectionRE.MatchString(name) {
			warnings = append(warnings, fmt.Sprintf("Line %d: unknown section %q", i+1, name))
		} else {
			blocks = append(blocks, line)
		}
	}
	return strings.Join(blocks, "\n\n"), unique(warnings)
}

func unique(in []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, v := range in {
		if !seen[v] {
			out = append(out, v)
			seen[v] = true
		}
	}
	return out
}

func validSectionName(name string, sections []Section) error {
	if name == "" {
		return fmt.Errorf("give the section a name")
	}
	if len([]rune(name)) > 40 || strings.ContainsAny(name, "[]\n\r\t") || repeatRE.MatchString(name) {
		return fmt.Errorf("use a short name without brackets or an x2 suffix")
	}
	for _, s := range sections {
		if strings.EqualFold(s.Name, name) {
			return fmt.Errorf("a section named %q already exists", name)
		}
	}
	return nil
}
