package input_new

import (
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/textinput"
)

func View(ti textinput.Model) string {
	return "Introduce la URL del nuevo repositorio:\n\n" + ti.View() + "\n" +
		shared.HelpStyle.Render("\nenter: guardar • esc: cancelar")
}
