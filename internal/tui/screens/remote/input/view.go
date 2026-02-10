package input

import (
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/textinput"
)

func View(ti textinput.Model, hasRemotes bool) string {
	help := "\nenter: continuar • esc/q: salir"
	if hasRemotes {
		help = "\nenter: continuar • esc: volver"
	}
	return "Introduce la URL del repositorio Git remoto:\n\n" + ti.View() + "\n" +
		shared.HelpStyle.Render(help)
}
