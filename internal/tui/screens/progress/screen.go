package progress

import (
	"skli/internal/gitrepo"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	// Sub-views
	"skli/internal/tui/screens/progress/done"
	"skli/internal/tui/screens/progress/downloading"
	"skli/internal/tui/screens/progress/error_view"
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

func (s ProgressScreen) Init() tea.Cmd {
	if s.State == StateDownloading {
		return s.Spinner.Tick
	}
	return nil
}

func (s ProgressScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case shared.DownloadResultMsg:
		if msg.Err != nil {
			s.State = StateError
			s.ErrorMessage = msg.Err.Error()
		} else {
			s.State = StateDone
		}
		return s, nil

	case spinner.TickMsg:
		if s.State == StateDownloading {
			var cmd tea.Cmd
			s.Spinner, cmd = s.Spinner.Update(msg)
			return s, cmd
		}

	case tea.KeyMsg:
		if msg.String() == "r" && s.State == StateError {
			return s, func() tea.Msg { return shared.NavigateToInputRemoteMsg{} }
		}
		return s, func() tea.Msg { return shared.QuitMsg{} }
	}

	return s, nil
}

func (s ProgressScreen) View() string {
	switch s.State {
	case StateDownloading:
		return downloading.View(s.Spinner, s.ConfigLocalPath)
	case StateDone:
		return done.View(s.ConfigMode, s.ConfigLocalPath)
	case StateError:
		return error_view.View(s.ErrorMessage)
	}
	return ""
}
