package provider

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/zhenninglang/mantis/internal/session"
)

type ClaudeAdapter struct{}

func (ClaudeAdapter) ID() ID              { return ClaudeCode }
func (ClaudeAdapter) DisplayName() string { return "Claude Code" }
func (ClaudeAdapter) Capabilities() Capabilities {
	return Capabilities{Resume: true, Fork: true, Inspect: true}
}

func (ClaudeAdapter) Discover() ([]session.Session, []Diagnostic) {
	root := homePath(".claude", "projects")
	entries, err := os.ReadDir(root)
	if err != nil {
		if ignoreMissing(err) {
			return nil, []Diagnostic{{Provider: ClaudeCode, Message: "~/.claude/projects not found"}}
		}
		return nil, []Diagnostic{{Provider: ClaudeCode, Message: err.Error()}}
	}

	var sessions []session.Session
	var diags []Diagnostic
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projectDir := filepath.Join(root, entry.Name())
		files, err := os.ReadDir(projectDir)
		if err != nil {
			diags = addDiag(diags, ClaudeCode, "read %s: %v", projectDir, err)
			continue
		}
		projectFull := encodedProjectPath(entry.Name())
		project := filepath.Base(projectFull)
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".jsonl") {
				continue
			}
			path := filepath.Join(projectDir, file.Name())
			info, err := os.Stat(path)
			if err != nil {
				continue
			}
			id := strings.TrimSuffix(file.Name(), ".jsonl")
			meta := readClaudeMeta(path)
			if meta.ID == "" {
				meta.ID = id
			}
			if meta.Title == "" {
				meta.Title = "Untitled"
			}
			if meta.WorkingDirectory == "" {
				meta.WorkingDirectory = projectFull
			}
			sessions = append(sessions, session.Session{
				Provider:     string(ClaudeCode),
				ProviderName: "Claude Code",
				ResumeRef:    meta.ID,
				ForkRef:      meta.ID,
				Meta:         meta,
				Project:      project,
				ProjectFull:  meta.WorkingDirectory,
				ModTime:      info.ModTime(),
				FilePath:     path,
			})
		}
	}
	return sessions, diags
}

func (ClaudeAdapter) Resume(s session.Session) error {
	return runCommand("claude", "--resume", s.ResumeRef)
}

func (ClaudeAdapter) Fork(s session.Session) error {
	return runCommand("claude", "--resume", s.ForkRef, "--fork-session")
}

func readClaudeMeta(path string) session.SessionMeta {
	f, err := os.Open(path)
	if err != nil {
		return session.SessionMeta{}
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var raw map[string]any
		if json.Unmarshal(line, &raw) != nil {
			continue
		}
		meta := session.SessionMeta{}
		for _, key := range []string{"sessionId", "session_id", "id"} {
			if v, ok := raw[key].(string); ok && strings.TrimSpace(v) != "" {
				meta.ID = v
				break
			}
		}
		for _, key := range []string{"title", "summary"} {
			if v, ok := raw[key].(string); ok && strings.TrimSpace(v) != "" {
				meta.Title = v
				break
			}
		}
		for _, key := range []string{"cwd", "workingDirectory"} {
			if v, ok := raw[key].(string); ok && strings.TrimSpace(v) != "" {
				meta.WorkingDirectory = v
				break
			}
		}
		return meta
	}
	return session.SessionMeta{}
}

func encodedProjectPath(name string) string {
	if name == "" {
		return ""
	}
	return "/" + strings.TrimLeft(strings.ReplaceAll(name, "-", "/"), "/")
}
