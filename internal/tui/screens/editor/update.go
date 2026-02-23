package editor

import (
	"strings"

	"skli/internal/gitrepo"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (s EditorScreen) Init() tea.Cmd {
	if s.State == StateInputCustom {
		return textinput.Blink
	}
	return nil
}

func (s EditorScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		s.List.SetSize(msg.Width, msg.Height-4)
	}

	switch s.State {
	case StateSelecting:
		return s.updateSelecting(msg)
	case StateInputCustom:
		return s.updateInputCustom(msg)
	}
	return s, nil
}

func (s EditorScreen) updateSelecting(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			if s.ConfigMode {
				return s, func() tea.Msg { return shared.NavigateToConfigMsg{} }
			}
			// Should ideally go back to previous screen
			return s, func() tea.Msg { return shared.QuitMsg{} }
		case "enter":
			selected := s.List.SelectedItem()
			if selected == nil {
				return s, nil
			}
			item := selected.(editorItem)

			if item.editor.Name == "Custom" {
				s.State = StateInputCustom
				s.TextInput.Focus()
				return s, textinput.Blink
			}

			destPath := item.editor.Path
			return s.proceedWithSelection(destPath)
		}
	}

	var cmd tea.Cmd
	s.List, cmd = s.List.Update(msg)
	return s, cmd
}

func (s EditorScreen) updateInputCustom(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.State = StateSelecting
			return s, nil
		case "enter":
			path := strings.TrimSpace(s.TextInput.Value())
			if path == "" {
				return s, nil
			}
			return s.proceedWithSelection(path)
		}
	}

	s.TextInput, cmd = s.TextInput.Update(msg)
	return s, cmd
}

func (s EditorScreen) proceedWithSelection(destPath string) (tea.Model, tea.Cmd) {
	if s.ConfigMode {
		return s, tea.Batch(
			shared.SaveConfigCmd(destPath, s.Remotes, true),
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
