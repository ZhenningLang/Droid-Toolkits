package main

import (
	"strings"
	"testing"

	"github.com/zhenninglang/mantis/internal/provider"
	"github.com/zhenninglang/mantis/internal/session"
)

func TestRunForkRequiresSinglePrefix(t *testing.T) {
	err := runFork(nil)
	if err == nil || !strings.Contains(err.Error(), "usage: mantis fork <session-id-prefix>") {
		t.Fatalf("runFork() error = %v", err)
	}
}

func TestRunForkForksResolvedPrefix(t *testing.T) {
	originalResolve := resolveForkSession
	originalFork := forkResolvedSession
	originalProvider := currentProvider
	t.Cleanup(func() {
		resolveForkSession = originalResolve
		forkResolvedSession = originalFork
		currentProvider = originalProvider
	})

	currentProvider = provider.Droid
	resolveForkSession = func(agent provider.ID, prefix string) (session.Session, error) {
		if agent != provider.Droid {
			t.Fatalf("resolveForkSession() agent = %q, want droid", agent)
		}
		if prefix != "0290318a" {
			t.Fatalf("resolveForkSessionID() got %q, want %q", prefix, "0290318a")
		}
		return session.Session{
			Provider:     string(provider.Droid),
			ProviderName: "Droid",
			ForkRef:      "0290318a-b368-4621-806d-8e2cf36bbf09",
			Meta: session.SessionMeta{
				ID: "0290318a-b368-4621-806d-8e2cf36bbf09",
			},
		}, nil
	}
	var gotID string
	forkResolvedSession = func(s session.Session) error {
		gotID = s.ForkRef
		return nil
	}

	if err := runFork([]string{"0290318a"}); err != nil {
		t.Fatalf("runFork() error = %v", err)
	}
	if gotID != "0290318a-b368-4621-806d-8e2cf36bbf09" {
		t.Fatalf("forkResolvedSession() got %q, want full session id", gotID)
	}
}
