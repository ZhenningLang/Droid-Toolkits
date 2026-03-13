package tui

import (
	"fmt"
	"strings"

	"github.com/zhenninglang/mantis/internal/session"
)

func renderStats(sessions []session.Session, width, height int) string {
	var b strings.Builder

	b.WriteString(previewTitleStyle.Render("Session Statistics"))
	b.WriteString("\n\n")

	b.WriteString(statLabelStyle.Render("Total sessions: "))
	b.WriteString(statValueStyle.Render(fmt.Sprintf("%d", len(sessions))))
	b.WriteString("\n\n")

	// per-project counts
	projects := map[string]int{}
	var totalInput, totalOutput int
	var totalActiveMs int

	for i := range sessions {
		s := &sessions[i]
		projects[s.ProjectShort()]++
		totalInput += s.Settings.TokenUsage.InputTokens
		totalOutput += s.Settings.TokenUsage.OutputTokens
		totalActiveMs += s.Settings.AssistantActiveTimeMs
	}

	b.WriteString(statLabelStyle.Render("By Project:"))
	b.WriteString("\n")

	type kv struct {
		k string
		v int
	}
	var sorted []kv
	for k, v := range projects {
		sorted = append(sorted, kv{k, v})
	}
	// sort by count desc
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].v > sorted[i].v {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	maxShow := min(height-10, len(sorted))
	if maxShow < 1 {
		maxShow = len(sorted)
	}
	for i, p := range sorted {
		if i >= maxShow {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  ... and %d more", len(sorted)-maxShow)))
			b.WriteString("\n")
			break
		}
		bar := strings.Repeat("█", min(p.v, 30))
		b.WriteString(fmt.Sprintf("  %-20s %s %d\n",
			projectStyle.Render(p.k),
			dimStyle.Render(bar),
			p.v))
	}

	b.WriteString("\n")
	b.WriteString(statLabelStyle.Render("Total Tokens: "))
	b.WriteString(statValueStyle.Render(fmt.Sprintf("%s in / %s out", formatTokens(totalInput), formatTokens(totalOutput))))
	b.WriteString("\n")

	b.WriteString(statLabelStyle.Render("Total Active Time: "))
	b.WriteString(statValueStyle.Render(formatDurationMs(totalActiveMs)))
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("Press Ctrl+S or Esc to return"))

	return b.String()
}

func formatDurationMs(ms int) string {
	secs := ms / 1000
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	mins := secs / 60
	if mins < 60 {
		return fmt.Sprintf("%dm %ds", mins, secs%60)
	}
	hours := mins / 60
	return fmt.Sprintf("%dh %dm", hours, mins%60)
}
