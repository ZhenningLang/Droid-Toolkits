package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 1)

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("229")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	projectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("110"))

	previewTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("215"))

	previewLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("110"))

	previewValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	userMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("114"))

	assistantMsgStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	confirmStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	markedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	statLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("110")).
			Bold(true)

	statValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229"))
)
