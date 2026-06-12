package provider

import (
	"reflect"
	"testing"

	"github.com/zhenninglang/mantis/internal/session"
)

func TestParseIDAcceptsSupportedAgents(t *testing.T) {
	cases := map[string]ID{
		"droid":       Droid,
		"claude":      ClaudeCode,
		"claude-code": ClaudeCode,
		"opencode":    OpenCode,
		"kilo":        Kilo,
		"all":         All,
	}
	for input, want := range cases {
		got, ok := ParseID(input)
		if !ok || got != want {
			t.Fatalf("ParseID(%q) = %q, %v; want %q, true", input, got, ok, want)
		}
	}
}

func TestAdaptersForKnownProvidersDoesNotError(t *testing.T) {
	for _, id := range []ID{Droid, ClaudeCode, OpenCode, Kilo, All} {
		if _, err := AdaptersFor(id); err != nil {
			t.Fatalf("AdaptersFor(%q) error = %v", id, err)
		}
	}
}

func TestResumeAndForkUseProviderCommands(t *testing.T) {
	original := runCommand
	t.Cleanup(func() { runCommand = original })

	var calls [][]string
	runCommand = func(name string, args ...string) error {
		calls = append(calls, append([]string{name}, args...))
		return nil
	}

	sessions := []session.Session{
		{Provider: string(Droid), ResumeRef: "d1", ForkRef: "d1"},
		{Provider: string(ClaudeCode), ResumeRef: "c1", ForkRef: "c1"},
		{Provider: string(OpenCode), ResumeRef: "o1", ForkRef: "o1"},
		{Provider: string(Kilo), ResumeRef: "k1", ForkRef: "k1"},
	}
	for _, s := range sessions {
		if err := Resume(s); err != nil {
			t.Fatalf("Resume(%q) error = %v", s.Provider, err)
		}
		if err := Fork(s); err != nil {
			t.Fatalf("Fork(%q) error = %v", s.Provider, err)
		}
	}

	want := [][]string{
		{"droid", "-r", "d1"},
		{"droid", "--fork", "d1"},
		{"claude", "--resume", "c1"},
		{"claude", "--resume", "c1", "--fork-session"},
		{"opencode", "--session", "o1"},
		{"opencode", "--session", "o1", "--fork"},
		{"kilo", "--session", "k1"},
		{"kilo", "--session", "k1", "--fork"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("commands = %#v, want %#v", calls, want)
	}
}
