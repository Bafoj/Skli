package uploading

import (
	"fmt"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/lipgloss"
)

func View(msg string) string {
	return fmt.Sprintf("\n  Subiendo Skill...\n\n  %s\n\n  %s",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Render(msg),
		shared.HelpStyle.Render("esc/enter: volver"),
	)
}
