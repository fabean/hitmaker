package main

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name, arrangement, want string
		warn                    int
	}{
		{"repeats and directions", "[sad cinematic intro]\n\nchorus x2\n[Outro]", "[sad cinematic intro]\n\n[Chorus]\nOne\nTwo\n\n[Chorus]\nOne\nTwo\n\n[Outro]", 0},
		{"numbered section and free lyrics", "verse 1\nA standalone line", "[Verse 1]\nVerse lyric\n\nA standalone line", 0},
		{"unknown reference", "Verse 9\nChorus x2", "[Chorus]\nOne\nTwo\n\n[Chorus]\nOne\nTwo", 1},
		{"empty section skipped", "Bridge\nBridge\n[Outro]", "[Outro]", 1},
		{"bounded repetition", "Chorus x0\nChorus x17\nChorus x99999999999999999999999999", "", 3},
		{"CRLF", "[Intro]\r\nChorus\r\n[Outro]", "[Intro]\n\n[Chorus]\nOne\nTwo\n\n[Outro]", 0},
		{"unicode", "Chorus", "[Chorus]\nOne\nTwo", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Song{Sections: []Section{{"Chorus", "One\nTwo"}, {"Verse 1", "Verse lyric"}, {"Bridge", ""}}, Arrangement: tt.arrangement}
			got, warn := s.Render()
			if got != tt.want || len(warn) != tt.warn {
				t.Fatalf("got %q (%v), want %q (%d warnings)", got, warn, tt.want, tt.warn)
			}
		})
	}
}
func TestRepeatedSectionsStayInSync(t *testing.T) {
	s := newSong("test")
	s.Arrangement = "Chorus x4"
	s.Sections[0].Lyrics = "original"
	s.Sections[0].Lyrics = "new lyric"
	out, w := s.Render()
	if len(w) > 0 || strings.Count(out, "new lyric") != 4 || strings.Contains(out, "original") {
		t.Fatal(out, w)
	}
}
func TestSectionNames(t *testing.T) {
	for _, n := range []string{"", "[Intro]", "Chorus x2", "chorus", "long\nname"} {
		if validSectionName(n, []Section{{Name: "Chorus"}}) == nil {
			t.Errorf("accepted %q", n)
		}
	}
	if err := validSectionName("Pre-Chorus", nil); err != nil {
		t.Fatal(err)
	}
}
