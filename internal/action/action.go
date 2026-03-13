package action

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/zhenninglang/mantis/internal/session"
)

func Resume(s *session.Session) error {
	cmd := exec.Command("droid", "-r", s.Meta.ID)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// change to the session's working directory if available
	if s.Meta.WorkingDirectory != "" {
		cmd.Dir = s.Meta.WorkingDirectory
	}
	return cmd.Run()
}

func Delete(s *session.Session) error {
	jsonl := s.FilePath
	settings := strings.TrimSuffix(jsonl, ".jsonl") + ".settings.json"

	if err := os.Remove(jsonl); err != nil {
		return fmt.Errorf("remove %s: %w", jsonl, err)
	}
	os.Remove(settings) // best effort
	return nil
}

func Rename(s *session.Session, newTitle string) error {
	data, err := os.ReadFile(s.FilePath)
	if err != nil {
		return err
	}

	lines := strings.SplitN(string(data), "\n", 2)
	if len(lines) == 0 {
		return fmt.Errorf("empty session file")
	}

	// replace title in first line JSON
	firstLine := lines[0]
	// find "title":"..." and replace the value
	// simple approach: unmarshal, modify, marshal the first line
	// but to preserve other fields, do string replacement
	old := fmt.Sprintf(`"title":"%s"`, s.Meta.Title)
	newStr := fmt.Sprintf(`"title":"%s"`, escapeJSON(newTitle))
	if strings.Contains(firstLine, old) {
		firstLine = strings.Replace(firstLine, old, newStr, 1)
	} else {
		// try with spaces around colon
		old = fmt.Sprintf(`"title": "%s"`, s.Meta.Title)
		newStr = fmt.Sprintf(`"title": "%s"`, escapeJSON(newTitle))
		firstLine = strings.Replace(firstLine, old, newStr, 1)
	}

	var result string
	if len(lines) > 1 {
		result = firstLine + "\n" + lines[1]
	} else {
		result = firstLine
	}

	return os.WriteFile(s.FilePath, []byte(result), 0644)
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
