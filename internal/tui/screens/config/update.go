package config

import (
	"skli/internal/tui/shared"

	tea "github.com/charmbracelet/bubbletea"
)

func (s ConfigScreen) Init() tea.Cmd {
	return nil
}

func (s ConfigScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.Cursor > 0 {
				s.Cursor--
			}
		case "down", "j":
			if s.Cursor < 2 {
				s.Cursor++
			}
		case "enter":
			if s.Cursor == 0 {
				return s, func() tea.Msg { return shared.NavigateToManageRemotesMsg{} }
			} else if s.Cursor == 1 {
				return s, func() tea.Msg { return shared.NavigateToEditorMsg{} }
			} else if s.Cursor == 2 {
				return s, tea.Batch(
					shared.SaveConfigCmd(s.ConfigLocalPath, s.Remotes),
					func() tea.Msg { return shared.NavigateToDoneMsg{ConfigMode: true, LocalPath: s.ConfigLocalPath} },
				)
			}
		case "q", "esc":
			return s, func() tea.Msg { return shared.QuitMsg{} }
		}
	}

	return s, nil
}
