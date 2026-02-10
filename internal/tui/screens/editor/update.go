package editor

import (
	"skli/internal/gitrepo"
	"skli/internal/tui/shared"

	tea "github.com/charmbracelet/bubbletea"
)

func (s EditorScreen) Init() tea.Cmd {
	return nil
}

func (s EditorScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.List.SetSize(msg.Width, msg.Height-4)
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			return s, func() tea.Msg { return shared.NavigateToConfigMsg{} }
		case "enter":
			selected := s.List.SelectedItem()
			if selected == nil {
				return s, nil
			}
			item := selected.(editorItem)
			destPath := item.editor.Path
			if item.editor.Name == "Custom" {
				destPath = "skills"
			}

			if s.ConfigMode {
				return s, tea.Batch(
					shared.SaveConfigCmd(destPath, s.Remotes),
					func() tea.Msg { return shared.NavigateToConfigMsg{} },
				)
			}

			var selectedSkills []gitrepo.SkillInfo
			for _, sk := range s.Skills {
				if sk.Selected {
					selectedSkills = append(selectedSkills, sk.Info)
				}
			}

			return s, func() tea.Msg {
				return shared.NavigateToProgressMsg{
					TempDir:         s.TempDir,
					RemoteURL:       s.RemoteURL,
					SkillsRoot:      s.SkillsRoot,
					ConfigLocalPath: destPath,
					CommitHash:      s.CommitHash,
					Selected:        selectedSkills,
				}
			}
		}
	}

	var cmd tea.Cmd
	s.List, cmd = s.List.Update(msg)
	return s, cmd
}
