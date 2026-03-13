package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"github.com/zhenninglang/mantis/internal/session"
)

func filterSessions(sessions []session.Session, query, projectFilter string) []int {
	// step 1: project filter
	var candidates []int
	if projectFilter == "" {
		candidates = make([]int, len(sessions))
		for i := range candidates {
			candidates[i] = i
		}
	} else {
		for i, s := range sessions {
			if s.ProjectShort() == projectFilter {
				candidates = append(candidates, i)
			}
		}
	}

	if query == "" {
		return candidates
	}

	// step 2: fuzzy search within candidates
	source := make(sessionSource, len(candidates))
	for i, idx := range candidates {
		s := sessions[idx]
		source[i] = fmt.Sprintf("%s %s %s", s.Meta.Title, s.ProjectShort(), extractFirstUserMsg(s))
	}

	matches := fuzzy.FindFrom(query, source)
	indices := make([]int, len(matches))
	for i, m := range matches {
		indices[i] = candidates[m.Index]
	}
	return indices
}

type sessionSource []string

func (s sessionSource) String(i int) string { return s[i] }
func (s sessionSource) Len() int            { return len(s) }

func extractFirstUserMsg(s session.Session) string {
	for _, msg := range s.Messages {
		if msg.Role == "user" {
			return extractText(msg.Content)
		}
	}
	return ""
}

func renderListItem(s *session.Session, width int, selected, marked, fullPath bool) string {
	proj := projectStyle.Render(fmt.Sprintf("[%s]", s.ProjectDisplay(fullPath)))
	title := s.Meta.Title
	if len(title) > 40 {
		title = title[:37] + "..."
	}

	model := modelShort(s.Settings.Model)
	ago := timeAgo(s.ModTime)

	// short session id (first 8 chars)
	sid := s.Meta.ID
	if len(sid) > 8 {
		sid = sid[:8]
	}
	sid = dimStyle.Render(sid)

	// build the line
	left := fmt.Sprintf("  %s %s %s", proj, sid, title)
	right := fmt.Sprintf("%s  %s", ago, model)

	// pad between left and right
	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	line := left + strings.Repeat(" ", gap) + right

	if marked {
		return markedStyle.Render("● " + line[2:])
	}
	if selected {
		return selectedStyle.Width(width).Render(line)
	}
	return normalStyle.Render(line)
}

func modelShort(model string) string {
	m := model
	m = strings.TrimPrefix(m, "custom:")
	m = strings.TrimPrefix(m, "anthropic/")
	// shorten common patterns
	replacements := map[string]string{
		"Claude-Opus-4.6-0":   "Opus 4.6",
		"Claude-Opus-4-0":     "Opus 4",
		"Claude-Sonnet-4-0":   "Sonnet 4",
		"claude-sonnet-4-20250514": "Sonnet 4",
		"claude-opus-4-20250514":   "Opus 4",
	}
	for k, v := range replacements {
		if strings.Contains(m, k) {
			return v
		}
	}
	if len(m) > 12 {
		return m[:12]
	}
	return m
}
