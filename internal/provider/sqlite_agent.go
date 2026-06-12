package provider

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/zhenninglang/mantis/internal/session"
)

type SQLiteAgentAdapter struct {
	id       ID
	name     string
	dbName   string
	cliName  string
	hasModel bool
}

func (a SQLiteAgentAdapter) ID() ID              { return a.id }
func (a SQLiteAgentAdapter) DisplayName() string { return a.name }
func (a SQLiteAgentAdapter) Capabilities() Capabilities {
	return Capabilities{Resume: true, Fork: true, Inspect: true}
}

func (a SQLiteAgentAdapter) Discover() ([]session.Session, []Diagnostic) {
	dbPath := homePath(".local", "share", string(a.id), a.dbName)
	if _, err := os.Stat(dbPath); err != nil {
		if ignoreMissing(err) {
			return nil, []Diagnostic{{Provider: a.id, Message: fmt.Sprintf("%s not found", dbPath)}}
		}
		return nil, []Diagnostic{{Provider: a.id, Message: err.Error()}}
	}

	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return nil, []Diagnostic{{Provider: a.id, Message: err.Error()}}
	}
	defer db.Close()

	query := `SELECT id, coalesce(title, ''), coalesce(directory, ''), coalesce(path, ''), '' AS model, coalesce(time_updated, time_created, 0) FROM session ORDER BY coalesce(time_updated, time_created, 0) DESC`
	if a.hasModel {
		query = `SELECT id, coalesce(title, ''), coalesce(directory, ''), coalesce(path, ''), coalesce(model, ''), coalesce(time_updated, time_created, 0) FROM session ORDER BY coalesce(time_updated, time_created, 0) DESC`
	}
	rows, err := db.Query(query)
	if err != nil {
		return nil, []Diagnostic{{Provider: a.id, Message: err.Error()}}
	}
	defer rows.Close()

	var sessions []session.Session
	for rows.Next() {
		var id, title, directory, pathValue, model string
		var updated any
		if err := rows.Scan(&id, &title, &directory, &pathValue, &model, &updated); err != nil {
			continue
		}
		cwd := firstNonEmpty(directory, pathValue)
		project := filepath.Base(cwd)
		if project == "." || project == string(filepath.Separator) || project == "" {
			project = "global"
		}
		if strings.TrimSpace(title) == "" {
			title = "Untitled"
		}
		sessions = append(sessions, session.Session{
			Provider:     string(a.id),
			ProviderName: a.name,
			ResumeRef:    id,
			ForkRef:      id,
			Meta: session.SessionMeta{
				ID:               id,
				Title:            title,
				WorkingDirectory: cwd,
			},
			Settings:    session.Settings{Model: model},
			Project:     project,
			ProjectFull: cwd,
			ModTime:     sqliteTime(updated),
			FilePath:    dbPath + "#" + id,
		})
	}
	return sessions, nil
}

func (a SQLiteAgentAdapter) Resume(s session.Session) error {
	return runCommand(a.cliName, "--session", s.ResumeRef)
}

func (a SQLiteAgentAdapter) Fork(s session.Session) error {
	return runCommand(a.cliName, "--session", s.ForkRef, "--fork")
}

func sqliteTime(value any) time.Time {
	switch v := value.(type) {
	case int64:
		return epochTime(v)
	case float64:
		return epochTime(int64(v))
	case []byte:
		return parseTimeString(string(v))
	case string:
		return parseTimeString(v)
	default:
		return time.Time{}
	}
}

func epochTime(v int64) time.Time {
	if v <= 0 {
		return time.Time{}
	}
	if v > 1_000_000_000_000 {
		return time.UnixMilli(v)
	}
	return time.Unix(v, 0)
}

func parseTimeString(v string) time.Time {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
		return t
	}
	return time.Time{}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
