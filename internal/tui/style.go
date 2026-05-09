package tui

import "github.com/charmbracelet/lipgloss"

// Card is a fixed inner width — the centered modal that everything renders
// inside. Picked to match sshwap.com's compact aesthetic and to keep the
// layout predictable across terminal widths.
const cardInnerWidth = 64

var (
	// Palette mirrors the sshwap-style monospace look: dim background,
	// accent green for headings/info, accent yellow for the active button
	// and highlighted labels.
	colAccent  = lipgloss.Color("#5FFF87") // green
	colWarn    = lipgloss.Color("#FFAF00") // amber — buttons, digit shortcuts
	colMuted   = lipgloss.Color("#7A7A7A")
	colError   = lipgloss.Color("#FF5F5F")
	colSuccess = lipgloss.Color("#5FFF87")

	styleCard = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colMuted).
			Padding(1, 2).
			Width(cardInnerWidth + 4)

	styleUser = lipgloss.NewStyle().
			Foreground(colMuted).
			Bold(true)

	styleTabActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(colWarn).
			Padding(0, 1).
			Bold(true)

	styleTabIdle = lipgloss.NewStyle().
			Foreground(colMuted).
			Padding(0, 1)

	styleAccent = lipgloss.NewStyle().
			Foreground(colAccent).
			Bold(true)

	styleWarn = lipgloss.NewStyle().
			Foreground(colWarn).
			Bold(true)

	styleField = lipgloss.NewStyle().
			Foreground(colMuted).
			Background(lipgloss.Color("#111111")).
			Padding(0, 1)

	styleFieldActive = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#1a1a1a")).
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
)
