package action

import (
	"fmt"
	"os"
	"strings"

	"github.com/zhenninglang/mantis/internal/session"
)

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
