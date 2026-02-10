package done

import (
	"fmt"
	"skli/internal/tui/shared"
)

func View(configMode bool, configLocalPath string) string {
	var msg string
	if configMode {
		msg = shared.SuccessStyle.Render("✔ ¡Configuración guardada correctamente!")
	} else {
		msg = shared.SuccessStyle.Render(fmt.Sprintf("✔ ¡Skills instaladas correctamente en ./%s/!", configLocalPath))
	}
	return msg + shared.HelpStyle.Render("\nPresiona cualquier tecla para salir")
}
