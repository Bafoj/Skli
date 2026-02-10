package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// ScanningScreen es el modelo para la pantalla de escaneo
type ScanningScreen struct {
	spinner    spinner.Model
	url        string
	skillsRoot string
}

// NewScanningScreen crea una nueva pantalla de escaneo
func NewScanningScreen(url, skillsRoot string) ScanningScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	return ScanningScreen{
		spinner:    s,
		url:        url,
		skillsRoot: skillsRoot,
	}
}

func (s ScanningScreen) Init() tea.Cmd {
	return tea.Batch(
		s.spinner.Tick,
		ScanRepoCmd(s.url, s.skillsRoot),
	)
}

func (s ScanningScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ScanResultMsg:
		if msg.Err != nil {
			return s, func() tea.Msg {
				return NavigateToErrorMsg{Err: msg.Err}
			}
		}
		return s, func() tea.Msg {
			return NavigateToSkillsMsg{
				Skills:     msg.Result.Skills,
				TempDir:    msg.Result.TempDir,
				RemoteURL:  msg.RemoteURL,
				SkillsRoot: msg.Result.SkillsPath,
				CommitHash: msg.Result.CommitHash,
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd
	}

	return s, nil
}

func (s ScanningScreen) View() string {
	return s.spinner.View() + " Escaneando el repositorio remoto..."
}
