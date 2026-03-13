package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zhenninglang/mantis/internal/session"
	"github.com/zhenninglang/mantis/internal/tui"
)

func main() {
	sessions, err := session.LoadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading sessions: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found in ~/.factory/sessions/")
		os.Exit(0)
	}

	m := tui.New(sessions)
	p := tea.NewProgram(m, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	model := result.(tui.Model)
	if id := model.ResumeID(); id != "" {
		// exec droid -r to replace this process
		droid, err := exec.LookPath("droid")
		if err != nil {
			fmt.Fprintf(os.Stderr, "droid not found: %v\n", err)
			os.Exit(1)
		}
		cmd := exec.Command(droid, "-r", id)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			os.Exit(1)
		}
	}
}
