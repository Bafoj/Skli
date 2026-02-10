package tui

import (
	"fmt"
	"strings"

	"skli/internal/gitrepo"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// ProgressScreenState representa el sub-estado de la pantalla de progreso
type ProgressScreenState int

const (
	ProgressStateDownloading ProgressScreenState = iota
	ProgressStateDone
	ProgressStateError
)

// ProgressScreen es el modelo para la pantalla de progreso
type ProgressScreen struct {
	state           ProgressScreenState
	spinner         spinner.Model
	configLocalPath string
	configMode      bool
	errorMessage    string
}

// NewProgressScreen crea una nueva pantalla de progreso (downloading)
func NewProgressScreen(tempDir, remoteURL, skillsRoot, configLocalPath, commitHash string, selected []gitrepo.SkillInfo) ProgressScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	return ProgressScreen{
		state:           ProgressStateDownloading,
		spinner:         s,
		configLocalPath: configLocalPath,
	}
}

// NewProgressScreenDownloading crea una pantalla de descarga con comando
func NewProgressScreenDownloading(tempDir, remoteURL, skillsRoot, configLocalPath, commitHash string, selected []gitrepo.SkillInfo) (ProgressScreen, tea.Cmd) {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	screen := ProgressScreen{
		state:           ProgressStateDownloading,
		spinner:         s,
		configLocalPath: configLocalPath,
	}

	return screen, tea.Batch(
		s.Tick,
		DownloadSkillsCmd(tempDir, remoteURL, skillsRoot, configLocalPath, commitHash, selected),
	)
}

// NewDoneScreen crea una pantalla de éxito
func NewDoneScreen(configMode bool, localPath string) ProgressScreen {
	return ProgressScreen{
		state:           ProgressStateDone,
		configMode:      configMode,
		configLocalPath: localPath,
	}
}

// NewErrorScreen crea una pantalla de error
func NewErrorScreen(err error) ProgressScreen {
	return ProgressScreen{
		state:        ProgressStateError,
		errorMessage: err.Error(),
	}
}

func (s ProgressScreen) Init() tea.Cmd {
	if s.state == ProgressStateDownloading {
		return s.spinner.Tick
	}
	return nil
}

func (s ProgressScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DownloadResultMsg:
		if msg.Err != nil {
			s.state = ProgressStateError
			s.errorMessage = msg.Err.Error()
		} else {
			s.state = ProgressStateDone
		}
		return s, nil

	case spinner.TickMsg:
		if s.state == ProgressStateDownloading {
			var cmd tea.Cmd
			s.spinner, cmd = s.spinner.Update(msg)
			return s, cmd
		}

	case tea.KeyMsg:
		if msg.String() == "r" && s.state == ProgressStateError {
			return s, func() tea.Msg { return NavigateToInputRemoteMsg{} }
		}
		return s, func() tea.Msg { return QuitMsg{} }
	}

	return s, nil
}

func (s ProgressScreen) View() string {
	var b strings.Builder

	switch s.state {
	case ProgressStateDownloading:
		b.WriteString(fmt.Sprintf("%s Instalando skills seleccionadas en %s...", s.spinner.View(), InfoStyle.Render(s.configLocalPath)))

	case ProgressStateDone:
		if s.configMode {
			b.WriteString(SuccessStyle.Render("✔ ¡Configuración guardada correctamente!"))
		} else {
			b.WriteString(SuccessStyle.Render(fmt.Sprintf("✔ ¡Skills instaladas correctamente en ./%s/!", s.configLocalPath)))
		}
		b.WriteString(HelpStyle.Render("\nPresiona cualquier tecla para salir"))

	case ProgressStateError:
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("✘ Error: %s", s.errorMessage)))
		b.WriteString(HelpStyle.Render("\nPresiona 'r' para reintentar o 'q' para salir"))
	}

	return b.String()
}
