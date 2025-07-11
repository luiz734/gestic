package selector

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var tableStyles = table.Styles{
	Selected: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#232627")).
		Background(lipgloss.Color("#fcfcfc")),
	Header: lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1),
		// Foreground(lipgloss.Color("#232627")).
		// Background(lipgloss.Color("#fcfcfc")),
	Cell: lipgloss.NewStyle().Padding(0, 1),
}
