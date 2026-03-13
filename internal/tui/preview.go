package tui

import (
	"fmt"
	"strings"

	"github.com/zhenninglang/mantis/internal/session"
)

func renderPreview(s *session.Session, width int) string {
	if s == nil {
		return dimStyle.Render("No session selected")
	}

	var b strings.Builder

	b.WriteString(previewTitleStyle.Render(s.Meta.Title))
	b.WriteString("\n")

	info := fmt.Sprintf("%s  |  %s  |  %s",
		previewLabelStyle.Render("Project: ")+previewValueStyle.Render(s.ProjectShort()),
		previewLabelStyle.Render("Model: ")+previewValueStyle.Render(modelShort(s.Settings.Model)),
		previewLabelStyle.Render("Time: ")+previewValueStyle.Render(timeAgo(s.ModTime)),
	)
	b.WriteString(info)
	b.WriteString("\n")

	tokens := fmt.Sprintf("%s  |  %s",
		previewLabelStyle.Render("Tokens: ")+previewValueStyle.Render(
			fmt.Sprintf("%s in / %s out",
				formatTokens(s.Settings.TokenUsage.InputTokens),
				formatTokens(s.Settings.TokenUsage.OutputTokens))),
		previewLabelStyle.Render("Active: ")+previewValueStyle.Render(formatDuration(s.ActiveDuration())),
	)
	b.WriteString(tokens)
	b.WriteString("\n")

	sep := dimStyle.Render(strings.Repeat("─", min(width-2, 60)))
	b.WriteString(sep)
	b.WriteString("\n")

	// show first few conversation turns
	count := 0
	for _, msg := range s.Messages {
		if msg.Role != "user" && msg.Role != "assistant" {
			continue
		}
		text := extractText(msg.Content)
		if text == "" {
			continue
		}
		// skip system reminders and tool noise
		if strings.HasPrefix(text, "<system-reminder>") || strings.HasPrefix(text, "<EXTREMELY") {
			continue
		}
		if len(text) > 120 {
			text = text[:117] + "..."
		}

		if msg.Role == "user" {
			b.WriteString(userMsgStyle.Render("User: ") + previewValueStyle.Render(text))
		} else {
			b.WriteString(assistantMsgStyle.Render("Asst: ") + dimStyle.Render(text))
		}
		b.WriteString("\n")
		count++
		if count >= 4 {
			break
		}
	}

	if count == 0 {
		b.WriteString(dimStyle.Render("(no messages)"))
	}

	return b.String()
}

func extractText(content any) string {
	switch v := content.(type) {
	case string:
		return singleLine(v)
	case []any:
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				if m["type"] == "text" {
					if t, ok := m["text"].(string); ok {
						return singleLine(t)
					}
				}
			}
		}
	}
	return ""
}

func singleLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	return s
}
