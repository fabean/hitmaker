package main

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

type actionMsg struct {
	text string
	err  error
}

func copyText(value, label string) tea.Cmd {
	return func() tea.Msg {
		var candidates [][]string
		switch runtime.GOOS {
		case "darwin":
			candidates = [][]string{{"pbcopy"}}
		case "windows":
			candidates = [][]string{{"clip.exe"}}
		default:
			candidates = [][]string{{"wl-copy"}, {"xclip", "-selection", "clipboard"}, {"xsel", "--clipboard", "--input"}}
		}
		for _, args := range candidates {
			if _, err := exec.LookPath(args[0]); err != nil {
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			cmd := exec.CommandContext(ctx, args[0], args[1:]...)
			cmd.Stdin = strings.NewReader(value)
			err := cmd.Run()
			cancel()
			if err == nil {
				return actionMsg{text: "Copied " + label}
			}
		}
		return clipboardFallback{value: value, label: label}
	}
}

type clipboardFallback struct{ value, label string }

func openSuno() tea.Cmd {
	return func() tea.Msg {
		var args []string
		switch runtime.GOOS {
		case "darwin":
			args = []string{"open", "https://suno.com/create"}
		case "windows":
			args = []string{"rundll32", "url.dll,FileProtocolHandler", "https://suno.com/create"}
		default:
			args = []string{"xdg-open", "https://suno.com/create"}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := exec.CommandContext(ctx, args[0], args[1:]...).Run(); err != nil {
			return actionMsg{err: fmt.Errorf("open Suno: %w", err)}
		}
		return actionMsg{text: "Opened Suno · choose Custom, then paste lyrics and style"}
	}
}
