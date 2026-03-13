package session

import "time"

type TokenUsage struct {
	InputTokens        int `json:"inputTokens"`
	OutputTokens       int `json:"outputTokens"`
	CacheCreationTokens int `json:"cacheCreationTokens"`
	CacheReadTokens    int `json:"cacheReadTokens"`
	ThinkingTokens     int `json:"thinkingTokens"`
}

type Settings struct {
	AssistantActiveTimeMs int        `json:"assistantActiveTimeMs"`
	Model                string     `json:"model"`
	AutonomyMode         string     `json:"autonomyMode"`
	TokenUsage           TokenUsage `json:"tokenUsage"`
}

type SessionMeta struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	WorkingDirectory string `json:"cwd"`
}

type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type Session struct {
	Meta       SessionMeta
	Settings   Settings
	Project    string
	ModTime    time.Time
	FilePath   string
	Messages   []Message
	Selected   bool // for batch operations
}

func (s *Session) ProjectShort() string {
	if s.Project == "" {
		return "global"
	}
	return s.Project
}

func (s *Session) TotalTokens() int {
	u := s.Settings.TokenUsage
	return u.InputTokens + u.OutputTokens + u.CacheCreationTokens + u.CacheReadTokens + u.ThinkingTokens
}

func (s *Session) ActiveDuration() time.Duration {
	return time.Duration(s.Settings.AssistantActiveTimeMs) * time.Millisecond
}
