package provider

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zhenninglang/mantis/internal/session"
)

type ID string

const (
	Droid      ID = "droid"
	ClaudeCode ID = "claude"
	OpenCode   ID = "opencode"
	Kilo       ID = "kilo"
	All        ID = "all"
)

type Capabilities struct {
	Resume  bool
	Fork    bool
	Rename  bool
	Delete  bool
	Inspect bool
}

type Diagnostic struct {
	Provider ID
	Message  string
}

type Adapter interface {
	ID() ID
	DisplayName() string
	Capabilities() Capabilities
	Discover() ([]session.Session, []Diagnostic)
	Resume(session.Session) error
	Fork(session.Session) error
}

type execRunner func(name string, args ...string) error

var runCommand execRunner = runInteractiveCommand

func IDs() []ID {
	return []ID{Droid, ClaudeCode, OpenCode, Kilo, All}
}

func ParseID(value string) (ID, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "droid":
		return Droid, true
	case "claude", "claude-code", "claudecode":
		return ClaudeCode, true
	case "opencode", "open-code":
		return OpenCode, true
	case "kilo":
		return Kilo, true
	case "all":
		return All, true
	default:
		return "", false
	}
}

func AllAdapters() []Adapter {
	return []Adapter{
		DroidAdapter{},
		ClaudeAdapter{},
		SQLiteAgentAdapter{id: OpenCode, name: "OpenCode", dbName: "opencode.db", cliName: "opencode", hasModel: true},
		SQLiteAgentAdapter{id: Kilo, name: "Kilo", dbName: "kilo.db", cliName: "kilo"},
	}
}

func AdapterFor(id ID) (Adapter, bool) {
	for _, adapter := range AllAdapters() {
		if adapter.ID() == id {
			return adapter, true
		}
	}
	return nil, false
}

func AdaptersFor(id ID) ([]Adapter, error) {
	if id == All {
		return AllAdapters(), nil
	}
	adapter, ok := AdapterFor(id)
	if !ok {
		return nil, fmt.Errorf("unknown agent %q", id)
	}
	return []Adapter{adapter}, nil
}

func Discover(id ID) ([]session.Session, []Diagnostic, error) {
	adapters, err := AdaptersFor(id)
	if err != nil {
		return nil, nil, err
	}

	var sessions []session.Session
	var diagnostics []Diagnostic
	for _, adapter := range adapters {
		ss, ds := adapter.Discover()
		sessions = append(sessions, ss...)
		diagnostics = append(diagnostics, ds...)
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ModTime.After(sessions[j].ModTime)
	})
	return sessions, diagnostics, nil
}

func Resume(s session.Session) error {
	adapter, ok := AdapterFor(ID(s.Provider))
	if !ok {
		return fmt.Errorf("unknown session provider %q", s.Provider)
	}
	if !adapter.Capabilities().Resume {
		return fmt.Errorf("%s resume is not supported", adapter.DisplayName())
	}
	return adapter.Resume(s)
}

func Fork(s session.Session) error {
	adapter, ok := AdapterFor(ID(s.Provider))
	if !ok {
		return fmt.Errorf("unknown session provider %q", s.Provider)
	}
	if !adapter.Capabilities().Fork {
		return fmt.Errorf("%s fork is not supported", adapter.DisplayName())
	}
	return adapter.Fork(s)
}

func ResolveByPrefix(sessions []session.Session, prefix string) (session.Session, error) {
	var matches []session.Session
	for _, s := range sessions {
		if strings.HasPrefix(s.Meta.ID, prefix) {
			matches = append(matches, s)
		}
	}

	switch len(matches) {
	case 0:
		return session.Session{}, fmt.Errorf("no session matches prefix %q", prefix)
	case 1:
		return matches[0], nil
	default:
		ids := make([]string, 0, len(matches))
		for _, s := range matches {
			ids = append(ids, fmt.Sprintf("%s:%s", s.Provider, s.Meta.ID))
		}
		return session.Session{}, fmt.Errorf("multiple sessions match prefix %q: %s", prefix, strings.Join(ids, ", "))
	}
}

func runInteractiveCommand(name string, args ...string) error {
	path, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%s not found: %w", name, err)
	}
	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func homePath(parts ...string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(append([]string{home}, parts...)...)
}

func addDiag(diags []Diagnostic, id ID, format string, args ...any) []Diagnostic {
	return append(diags, Diagnostic{Provider: id, Message: fmt.Sprintf(format, args...)})
}

func ignoreMissing(err error) bool {
	return err == nil || errors.Is(err, os.ErrNotExist)
}
