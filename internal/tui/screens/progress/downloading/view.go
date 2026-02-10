package downloading

import (
	"fmt"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/spinner"
)

func View(s spinner.Model, configLocalPath string) string {
	return fmt.Sprintf("%s Instalando skills seleccionadas en %s...", s.View(), shared.InfoStyle.Render(configLocalPath))
}
