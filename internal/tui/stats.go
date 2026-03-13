package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

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
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].v > sorted[j].v })

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
	b.WriteString(statValueStyle.Render(formatDuration(time.Duration(totalActiveMs) * time.Millisecond)))
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("Press Ctrl+S or Esc to return"))

	return b.String()
}
