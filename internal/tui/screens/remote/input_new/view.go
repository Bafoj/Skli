package input_new

import (
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/textinput"
)

func View(ti textinput.Model) string {
	return "Enter the new repository URL:\n\n" + ti.View() + "\n" +
		shared.HelpStyle.Render("\nenter: save • esc: cancel")
}
