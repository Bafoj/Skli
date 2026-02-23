package shared

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Estilos compartidos para todas las pantallas del TUI
var (
	ColorBgMain     = lipgloss.Color("#0F1117")
	ColorSurface    = lipgloss.Color("#1B2130")
	ColorAccent     = lipgloss.Color("#5DD6C0")
	ColorAccentSoft = lipgloss.Color("#2A9D8F")
	ColorTextMain   = lipgloss.Color("#E6EAF2")
	ColorTextDim    = lipgloss.Color("#7D8596")
	ColorDanger     = lipgloss.Color("#FF6B6B")
	ColorSuccess    = lipgloss.Color("#7CFCB2")

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorTextMain).
			Background(ColorSurface).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccentSoft).
			Padding(0, 1).
			MarginBottom(1)

	ItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(ColorAccent).
				Bold(true)

	DescriptionStyle = lipgloss.NewStyle().
				Foreground(ColorTextDim).
				Italic(true).
				PaddingLeft(6)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			MarginTop(1)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Italic(true).
			MarginTop(1)

	SpinnerStyle = lipgloss.NewStyle().Foreground(ColorAccent)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true).
			MarginTop(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true).
			MarginTop(1)

	DimStyle = lipgloss.NewStyle().Foreground(ColorTextDim)

	InfoStyle = lipgloss.NewStyle().Foreground(ColorAccent)

	PopupErrorStyle = lipgloss.NewStyle().
			Foreground(ColorDanger).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorDanger).
			Background(ColorBgMain).
			Padding(0, 1)

	SelectorOnStyle  = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	SelectorOffStyle = lipgloss.NewStyle().Foreground(ColorTextDim)
)

func NewListItemStyles() list.DefaultItemStyles {
	styles := list.NewDefaultItemStyles()
	styles.NormalTitle = styles.NormalTitle.Foreground(ColorTextMain)
	styles.NormalDesc = styles.NormalDesc.Foreground(ColorTextDim)
	styles.SelectedTitle = styles.SelectedTitle.
		Foreground(ColorAccent).
		BorderForeground(ColorAccentSoft)
	styles.SelectedDesc = styles.SelectedDesc.
		Foreground(ColorAccentSoft).
		BorderForeground(ColorAccentSoft)
	return styles
}

func SelectorDot(active bool) string {
	if active {
		return SelectorOnStyle.Render("●")
	}
	return SelectorOffStyle.Render("○")
}

func ErrorPopup(msg string) string {
	return PopupErrorStyle.Render(fmt.Sprintf("⚠ %s", msg))
}
