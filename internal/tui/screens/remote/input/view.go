package input

import (
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/textinput"
)

func View(ti textinput.Model, hasRemotes bool) string {
	help := "\nenter: continue • esc/q: quit"
	if hasRemotes {
		help = "\nenter: continue • esc: back"
	}
	return "Enter the remote Git repository URL:\n\n" + ti.View() + "\n" +
		shared.HelpStyle.Render(help)
}
