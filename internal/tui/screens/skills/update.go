package skills

import (
	"skli/internal/gitrepo"
	"skli/internal/tui/shared"

	tea "github.com/charmbracelet/bubbletea"
)

func (s SkillsScreen) Init() tea.Cmd {
	return nil
}

func (s SkillsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.List.SetSize(msg.Width, msg.Height-4)
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			var selected []gitrepo.SkillInfo
			for _, sk := range s.Skills {
				if sk.Selected {
					selected = append(selected, sk.Info)
				}
			}
			if len(selected) > 0 {
				if s.ConfigLocalPath == "" {
					return s, func() tea.Msg {
						return shared.NavigateToEditorMsg{
							Skills:     s.Skills,
							TempDir:    s.TempDir,
							RemoteURL:  s.RemoteURL,
							SkillsRoot: s.SkillsRoot,
							CommitHash: s.CommitHash,
						}
					}
				}
				return s, func() tea.Msg {
					return shared.NavigateToProgressMsg{
						TempDir:         s.TempDir,
						RemoteURL:       s.RemoteURL,
						SkillsRoot:      s.SkillsRoot,
						ConfigLocalPath: s.ConfigLocalPath,
						CommitHash:      s.CommitHash,
						Selected:        selected,
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	s.List, cmd = s.List.Update(msg)
	return s, cmd
}
