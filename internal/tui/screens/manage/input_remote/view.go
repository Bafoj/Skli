package input_remote

import (
	"fmt"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/textinput"
)

func View(remoteInput textinput.Model) string {
	return fmt.Sprintf("\n  Introduce la URL del repositorio destino:\n\n  %s\n\n  %s",
		remoteInput.View(),
		shared.HelpStyle.Render("enter: confirmar • esc: volver"),
	)
}
