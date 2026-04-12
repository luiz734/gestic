package clip

import (
	"github.com/charmbracelet/lipgloss"
)

var defaultStyle = lipgloss.NewStyle().
	Bold(false).
	Background(lipgloss.Color("#232627")).
	Foreground(lipgloss.Color("#fcfcfc"))

var blinkStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#232627")).
	Background(lipgloss.Color("#fcfcfc"))
