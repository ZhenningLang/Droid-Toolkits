package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zhenninglang/mantis/internal/action"
	"github.com/zhenninglang/mantis/internal/completion"
	"github.com/zhenninglang/mantis/internal/config"
	"github.com/zhenninglang/mantis/internal/inspect"
	"github.com/zhenninglang/mantis/internal/provider"
	"github.com/zhenninglang/mantis/internal/session"
	"github.com/zhenninglang/mantis/internal/status"
	"github.com/zhenninglang/mantis/internal/summary"
	"github.com/zhenninglang/mantis/internal/tui"
)

var version = "dev"

var currentProvider = provider.Droid

var resolveForkSession = func(agent provider.ID, prefix string) (session.Session, error) {
	sessions, _, err := provider.Discover(agent)
	if err != nil {
		return session.Session{}, fmt.Errorf("load sessions: %w", err)
	}
	return provider.ResolveByPrefix(sessions, prefix)
}

var forkResolvedSession = provider.Fork

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "--agent" {
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: usage: mantis --agent <droid|claude|opencode|kilo|all>")
			os.Exit(1)
		}
		agent, ok := provider.ParseID(args[1])
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: unknown agent %q\n", args[1])
			os.Exit(1)
		}
		currentProvider = agent
		runTUI(agent)
		return
	}

	if len(args) > 0 {
		if agent, ok := provider.ParseID(args[0]); ok {
			currentProvider = agent
			runTUI(agent)
			return
		}

		switch args[0] {
		case "config":
			if err := config.RunSetup(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "version":
			fmt.Printf("mantis %s\n", version)
			return
		case "status":
			if err := status.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "index":
			if err := runIndex(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "help", "-h", "--help":
			printHelp()
			return
		case "clean":
			if err := runClean(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "inspect":
			cfg := config.Load()
			if !cfg.HasLLM() {
				fmt.Fprintln(os.Stderr, "LLM not configured. Run `mantis config` first.")
				os.Exit(1)
			}
			if err := inspect.Run(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "fork":
			if err := runFork(args[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "completion":
			if err := runCompletion(args[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		default:
			if strings.Contains(args[0], "compress") {
				fmt.Fprintln(os.Stderr, "compress has been removed")
			} else {
				fmt.Fprintf(os.Stderr, "Unknown command: %s\nRun `mantis help` for usage.\n", args[0])
			}
			os.Exit(1)
		}
	}

	agent, err := chooseProvider()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	runTUI(agent)
}

func runTUI(agent provider.ID) {
	cfg := config.Load()
	sessions, diagnostics, err := provider.Discover(agent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading sessions: %v\n", err)
		os.Exit(1)
	}
	if len(sessions) == 0 {
		for _, diagnostic := range diagnostics {
			fmt.Fprintf(os.Stderr, "%s: %s\n", diagnostic.Provider, diagnostic.Message)
		}
		fmt.Println("No sessions found.")
		os.Exit(0)
	}

	cwd, _ := os.Getwd()
	m := tui.New(sessions, version, cfg, cwd)
	p := tea.NewProgram(m, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	model := result.(*tui.Model)
	if id := model.ResumeID(); id != "" {
		selected, err := provider.ResolveByPrefix(sessions, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := provider.Resume(selected); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func printHelp() {
	fmt.Printf(`mantis %s — Browse and search agent chat sessions

Usage: mantis [agent|command]

Commands:
  (none)     Choose an agent, then launch interactive TUI
  droid      Browse Droid sessions
  claude     Browse Claude Code sessions
  opencode   Browse OpenCode sessions
  kilo       Browse Kilo sessions
  all        Browse all supported sessions
  inspect    Context Health Inspector — analyze sessions for optimization
  fork       Fork a session by ID prefix and resume the fork
  completion Print shell completion script for bash/zsh/fish
  config     Configure LLM for smart search and inspect
  index      Generate AI summaries for all sessions (--force to regenerate all, --retry to redo empty ones)
  status     Show indexing status and statistics
  clean      Remove all empty sessions (no user messages)
  version    Print version
  help       Show this help

Keybindings (TUI):
  Type       Fuzzy search
  ↑/↓        Navigate
  Enter      Resume session
  Tab        Toggle project path
  Ctrl+P     Filter by project
  Ctrl+D     Delete session
  Ctrl+X     Batch delete
  Ctrl+R     Rename session
  Ctrl+S     Statistics panel
  Esc        Clear search / Clear filter / Quit
`, version)
}

func runIndex() error {
	cfg := config.Load()
	if !cfg.HasLLM() {
		return fmt.Errorf("LLM not configured. Run `mantis config` first")
	}

	flag := ""
	if len(os.Args) > 2 {
		flag = os.Args[2]
	}

	sessions, err := session.LoadAll()
	if err != nil {
		return err
	}

	switch flag {
	case "--force", "-f":
		os.RemoveAll(summary.Dir())
		fmt.Println("Cleared all existing summaries.")
	case "--retry", "-r":
		removed := summary.RemoveEmpty(sessions)
		fmt.Printf("Removed %d empty summaries for retry.\n", removed)
	}

	ch, total := summary.GenerateMissing(context.Background(), cfg.LLM, sessions)
	if total == 0 {
		fmt.Println("All sessions already indexed.")
		return nil
	}

	fmt.Printf("Indexing %d sessions...\n", total)
	errors := 0
	for p := range ch {
		if p.Err != nil {
			errors++
			fmt.Printf("  [%d/%d] ERROR %s: %v\n", p.Done, total, p.Current, p.Err)
		} else if p.Summary != nil && p.Summary.Title != "" {
			fmt.Printf("  [%d/%d] %s\n", p.Done, total, p.Summary.Title)
		} else {
			fmt.Printf("  [%d/%d] %s (skipped, no messages)\n", p.Done, total, p.Current)
		}
	}

	fmt.Printf("\nDone. Indexed %d/%d sessions", total-errors, total)
	if errors > 0 {
		fmt.Printf(" (%d errors)", errors)
	}
	fmt.Println()
	return nil
}

func runClean() error {
	sessions, err := session.LoadAll()
	if err != nil {
		return err
	}

	var empty []int
	for i := range sessions {
		hasUser := false
		for _, msg := range sessions[i].Messages {
			if msg.Role == "user" {
				hasUser = true
				break
			}
		}
		if !hasUser {
			empty = append(empty, i)
		}
	}

	if len(empty) == 0 {
		fmt.Println("No empty sessions found.")
		return nil
	}

	fmt.Printf("Found %d empty sessions (no user messages). Delete all? [y/N] ", len(empty))
	var answer string
	fmt.Scanln(&answer)
	if answer != "y" && answer != "Y" {
		fmt.Println("Cancelled.")
		return nil
	}

	deleted := 0
	for _, idx := range empty {
		if err := action.Delete(&sessions[idx]); err != nil {
			fmt.Printf("  Failed to delete %s: %v\n", sessions[idx].Meta.ID, err)
		} else {
			deleted++
		}
	}
	fmt.Printf("Deleted %d empty sessions.\n", deleted)
	return nil
}

func runFork(args []string) error {
	if len(args) != 1 || args[0] == "" {
		return fmt.Errorf("usage: mantis fork <session-id-prefix>")
	}
	s, err := resolveForkSession(currentProvider, args[0])
	if err != nil {
		return err
	}
	fmt.Printf("[fork] Forking %s session %s...\n", s.ProviderName, s.Meta.ID)
	return forkResolvedSession(s)
}

func runCompletion(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: mantis completion <bash|zsh|fish>")
	}
	script, err := completion.Generate(args[0])
	if err != nil {
		return err
	}
	fmt.Print(script)
	return nil
}

func chooseProvider() (provider.ID, error) {
	options := []provider.ID{provider.Droid, provider.ClaudeCode, provider.OpenCode, provider.Kilo, provider.All}
	fmt.Println("Choose agent platform:")
	for i, id := range options {
		name := string(id)
		if adapter, ok := provider.AdapterFor(id); ok {
			name = adapter.DisplayName()
		} else if id == provider.All {
			name = "All"
		}
		fmt.Printf("  %d. %s\n", i+1, name)
	}
	fmt.Print("Select [1-5]: ")
	var choice int
	if _, err := fmt.Scanln(&choice); err != nil {
		return "", err
	}
	if choice < 1 || choice > len(options) {
		return "", fmt.Errorf("invalid platform selection %d", choice)
	}
	return options[choice-1], nil
}
