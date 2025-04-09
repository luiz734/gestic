package compare

import "github.com/charmbracelet/lipgloss"

// 	"gestic/ui"
// 	"github.com/charmbracelet/lipgloss"
// )

var (
// styleHelp  = models.StyleHelp
// styleLabel = lipgloss.NewStyle().
//
//	Bold(true).
//	Foreground(colors.Pink)
//
// styleValue = lipgloss.NewStyle().
//
//	Foreground(colors.Surface2)
//
// styleSeparator = lipgloss.NewStyle().
//
//	Foreground(colors.Surface0)
//
// styleWrapper = lipgloss.NewStyle().
//
//			Border(lipgloss.DoubleBorder()).
//			BorderForeground(colors.Surface0).
//	           Padding(1, 3)
)

func ASCIIBorder() lipgloss.Border {
	return lipgloss.Border{
		Top:          "-",
		Bottom:       "-",
		Left:         "|",
		Right:        "|",
		TopLeft:      "+",
		TopRight:     "+",
		BottomLeft:   "+",
		BottomRight:  "+",
		MiddleLeft:   "+",
		MiddleRight:  "+",
		Middle:       "+",
		MiddleTop:    "+",
		MiddleBottom: "+",
	}
}

var (
	focusStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#12")).Bold(true)
	defaultStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#12")).Bold(false)
)
