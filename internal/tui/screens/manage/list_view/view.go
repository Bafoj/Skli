package list_view

import (
	"fmt"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func View(l list.Model, skillCount int, msg string) string {
	if skillCount == 0 {
		return "\n  No hay skills para mostrar.\n\n" + shared.HelpStyle.Render("  q: salir")
	}

	if msg != "" {
		return fmt.Sprintf("\n  %s\n\n%s", lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render(msg), l.View())
	}

	originalTitle := l.Title
	l.Title = fmt.Sprintf("%s (Pág. %d/%d)", originalTitle, l.Paginator.Page+1, l.Paginator.TotalPages)
	view := l.View()
	l.Title = originalTitle
	return view
}
