package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Two accent colours, monospace, restrained — matches the kyc.rip aesthetic.
	colAccent  = lipgloss.Color("#00D7AF") // teal — actionable / focus
	colWarn    = lipgloss.Color("#FFAF00") // amber — secondary / pending
	colMuted   = lipgloss.Color("#7A7A7A")
	colError   = lipgloss.Color("#FF5F5F")
	colSuccess = lipgloss.Color("#5FFF87")

	styleTitle = lipgloss.NewStyle().
			Foreground(colAccent).
			Bold(true)

	styleTabActive = lipgloss.NewStyle().
			Foreground(colAccent).
			Underline(true).
			Bold(true)

	styleTabIdle = lipgloss.NewStyle().
			Foreground(colMuted)

	styleField = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colMuted).
			Padding(0, 1)

	styleFieldActive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colAccent).
				Padding(0, 1)

	styleButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(colWarn).
			Bold(true).
			Padding(0, 2)

	styleButtonIdle = lipgloss.NewStyle().
			Foreground(colMuted).
			Padding(0, 2)

	styleErr = lipgloss.NewStyle().Foreground(colError)
	styleOk  = lipgloss.NewStyle().Foreground(colSuccess)
	styleDim = lipgloss.NewStyle().Foreground(colMuted)

	styleFooter = lipgloss.NewStyle().
			Foreground(colMuted).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colMuted).
			PaddingTop(1)
)
