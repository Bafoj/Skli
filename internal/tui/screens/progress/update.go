package progress

import (
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

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
