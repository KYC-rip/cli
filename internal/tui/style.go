package tui

import "github.com/charmbracelet/lipgloss"

// Card inner width — fixed so the layout is stable across terminal sizes.
const cardInnerWidth = 64

var (
	// Saturated palette for "colorful terminal" pop. Tuned to match the
	// sshwap.com aesthetic: bright lime green for headings/info,
	// saturated yellow for active labels and the primary button,
	// near-black background bands for input fields.
	colAccent  = lipgloss.Color("#00FF87") // bright lime
	colWarn    = lipgloss.Color("#FFD700") // gold/yellow
	colMuted   = lipgloss.Color("#8A8A8A")
	colError   = lipgloss.Color("#FF6060")
	colSuccess = lipgloss.Color("#00FF87")
	colInkBg   = lipgloss.Color("#0a0a0a")
	colFieldBg = lipgloss.Color("#1a1a1a")

	styleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colAccent).
			Padding(1, 2).
			Width(cardInnerWidth + 4)

	// Username pill on the left of the header — bright accent on dark.
	styleUser = lipgloss.NewStyle().
			Foreground(colAccent).
			Bold(true)

	// Active tab — yellow background, black text, bold (sshwap-style).
	styleTabActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(colWarn).
			Padding(0, 1).
			Bold(true)

	styleTabIdle = lipgloss.NewStyle().
			Foreground(colMuted).
			Padding(0, 1)

	// "Sending:", "Receiving:" labels — inverted yellow pill.
	styleAccent = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(colWarn).
			Bold(true).
			Padding(0, 1)

	// Standalone yellow text (digit shortcuts in the picker, etc.)
	styleWarn = lipgloss.NewStyle().
			Foreground(colWarn).
			Bold(true)

	// Idle input field background.
	styleField = lipgloss.NewStyle().
			Foreground(colMuted).
			Background(colFieldBg).
			Padding(0, 1)

	// Active input — brighter, plus a yellow underline so it's obvious
	// where the cursor is on muted terminal themes.
	styleFieldActive = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(colFieldBg).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(colWarn).
				Padding(0, 1)

	// Primary button — yellow on black, padded.
	styleButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(colWarn).
			Bold(true).
			Padding(0, 2)

	styleButtonIdle = lipgloss.NewStyle().
			Foreground(colMuted).
			Padding(0, 2)

	styleErr = lipgloss.NewStyle().Foreground(colError).Bold(true)
	styleOk  = lipgloss.NewStyle().Foreground(colSuccess).Bold(true)
	styleDim = lipgloss.NewStyle().Foreground(colMuted)
)
