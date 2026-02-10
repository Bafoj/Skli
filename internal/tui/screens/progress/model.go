package progress

import (
	"skli/internal/gitrepo"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type State int

const (
	StateDownloading State = iota
	StateDone
	StateError
)

// ProgressScreen es el modelo para la pantalla de progreso
type ProgressScreen struct {
	State           State
	Spinner         spinner.Model
	ConfigLocalPath string
	ConfigMode      bool
	ErrorMessage    string
}

// NewProgressScreenDownloading crea una pantalla de descarga con comando
func NewProgressScreenDownloading(tempDir, remoteURL, skillsRoot, configLocalPath, commitHash string, selected []gitrepo.SkillInfo) (ProgressScreen, tea.Cmd) {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = shared.SpinnerStyle

	screen := ProgressScreen{
		State:           StateDownloading,
		Spinner:         s,
		ConfigLocalPath: configLocalPath,
	}

	return screen, tea.Batch(
		s.Tick,
		shared.DownloadSkillsCmd(tempDir, remoteURL, skillsRoot, configLocalPath, commitHash, selected),
	)
}

// NewDoneScreen crea una pantalla de éxito
func NewDoneScreen(configMode bool, localPath string) ProgressScreen {
	return ProgressScreen{
		State:           StateDone,
		ConfigMode:      configMode,
		ConfigLocalPath: localPath,
	}
}

// NewErrorScreen crea una pantalla de error
func NewErrorScreen(err error) ProgressScreen {
	return ProgressScreen{
		State:        StateError,
		ErrorMessage: err.Error(),
	}
}
