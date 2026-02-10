package tui

import "github.com/charmbracelet/lipgloss"

// Estilos compartidos para todas las pantallas del TUI
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	ItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	DescriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Italic(true).
				PaddingLeft(6)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Italic(true).
			MarginTop(1)

	SpinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			MarginTop(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true).
			MarginTop(1)

	DimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))

	InfoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
)
