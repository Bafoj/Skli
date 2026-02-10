package scanning

import (
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (s ScanningScreen) Init() tea.Cmd {
	return tea.Batch(
		s.Spinner.Tick,
		shared.ScanRepoCmd(s.URL, s.SkillsRoot),
	)
}

func (s ScanningScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case shared.ScanResultMsg:
		if msg.Err != nil {
			return s, func() tea.Msg {
				return shared.NavigateToErrorMsg{Err: msg.Err}
			}
		}
		return s, func() tea.Msg {
			return shared.NavigateToSkillsMsg{
				Skills:     msg.Result.Skills,
				TempDir:    msg.Result.TempDir,
				RemoteURL:  msg.RemoteURL,
				SkillsRoot: msg.Result.SkillsPath,
				CommitHash: msg.Result.CommitHash,
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		s.Spinner, cmd = s.Spinner.Update(msg)
		return s, cmd
	}

	return s, nil
}
