package confirm

import (
	"fmt"
	"skli/internal/db"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/lipgloss"
)

func View(toDelete *db.InstalledSkill, confirmCursor int) string {
	if toDelete == nil {
		return ""
	}

	var yes, no string
	if confirmCursor == 0 {
		yes = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true).Render("➜ [ Yes ]")
		no = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  [ No ]")
	} else {
		yes = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  [ Yes ]")
		no = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true).Render("➜ [ No ]")
	}

	return fmt.Sprintf(
		"\n  Are you sure you want to delete skill %s?\n\n  Path: %s\n\n  %s    %s",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true).Render(toDelete.Name),
		shared.DimStyle.Render(toDelete.Path),
		yes,
		no,
	)
}
