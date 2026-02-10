package scanning

import (
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/spinner"
)

// ScanningScreen es el modelo para la pantalla de escaneo
type ScanningScreen struct {
	Spinner    spinner.Model
	URL        string
	SkillsRoot string
}

// NewScanningScreen crea una nueva pantalla de escaneo
func NewScanningScreen(url, skillsRoot string) ScanningScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = shared.SpinnerStyle

	return ScanningScreen{
		Spinner:    s,
		URL:        url,
		SkillsRoot: skillsRoot,
	}
}
