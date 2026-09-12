package main

import (
	tea "charm.land/bubbletea/v2"
	"flag"
	"fmt"
	"os"
)

var version = "0.1.0"

func main() {
	data := flag.String("data-dir", defaultDataDir(), "directory for songs and exports")
	config := flag.String("config-dir", defaultConfigDir(), "directory for custom styles.json and arrangements.json")
	demo := flag.Bool("demo", false, "start with an editable example song")
	list := flag.Bool("list", false, "list saved song IDs and titles")
	export := flag.String("export", "", "export a saved song ID to lyrics, style, and title files")
	ver := flag.Bool("version", false, "print version")
	flag.Parse()
	if *ver {
		fmt.Println("hitmaker " + version)
		return
	}
	st := Store{Dir: *data}
	if *export != "" {
		s, err := st.Load(*export)
		if err != nil {
			fatal(err)
		}
		dir, err := st.Export(s)
		if err != nil {
			fatal(err)
		}
		fmt.Println(dir)
		return
	}
	songs, err := st.List()
	if *list {
		if err != nil {
			fatal(err)
		}
		for _, s := range songs {
			fmt.Printf("%s  %s\n", s.ID, s.Title)
		}
		return
	}
	s := newSong("Untitled song")
	if *demo {
		s = demoSong()
	} else if len(songs) > 0 {
		s = songs[0]
	}
	m := newModel(st, s)
	m.configDir = *config
	if err != nil {
		m.status = err.Error()
		m.statusError = true
	}
	if len(songs) == 0 && !*demo {
		m.openInput("new", "", "Song title or a working idea…")
	}
	if _, err = tea.NewProgram(&m).Run(); err != nil {
		fatal(err)
	}
}
func fatal(err error) { fmt.Fprintln(os.Stderr, "hitmaker:", err); os.Exit(1) }
