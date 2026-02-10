package manage

import (
	"fmt"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (s ManageScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch s.State {
	case StateList:
		return s.updateList(msg)
	case StateConfirm:
		return s.updateConfirm(msg)
	case StateSelectingRemote:
		return s.updateSelectingRemote(msg)
	case StateInputRemote:
		return s.updateInputRemote(msg)
	case StateUploading:
		return s.updateUploading(msg)
	}
	return s, nil
}

func (s ManageScreen) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.List.SetSize(msg.Width, msg.Height-4)
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "d", "backspace":
			if item, ok := s.List.SelectedItem().(InstalledSkillItem); ok {
				s.ToDelete = &item.Skill
				s.State = StateConfirm
				s.ConfirmCursor = 1 // Reset a No
				return s, nil
			}
		case "esc", "q":
			return s, func() tea.Msg { return shared.QuitMsg{} }
		case "u":
			if item, ok := s.List.SelectedItem().(InstalledSkillItem); ok {
				s.SelectedSkill = &item.Skill
				// Preparar lista de remotes
				var items []list.Item
				seen := make(map[string]bool)

				// 1. Añadir remote del skill si existe (ORIGIN)
				if item.Skill.RemoteRepo != "" {
					items = append(items, remoteItem{
						url:         item.Skill.RemoteRepo,
						displayName: fmt.Sprintf("%s (Origin)", item.Skill.RemoteRepo),
					})
					seen[item.Skill.RemoteRepo] = true
				}

				// 2. Añadir remotes configurados
				for _, r := range s.ConfigRemotes {
					if !seen[r] {
						items = append(items, remoteItem{url: r})
						seen[r] = true
					}
				}

				// 3. Añadir opcion custom
				items = append(items, customURLItem{})

				delegate := newRemoteDelegate()
				l := list.New(items, delegate, 60, 14)
				l.Title = "Selecciona repositorio destino"
				l.SetShowStatusBar(false)
				l.SetFilteringEnabled(false)
				l.Styles.Title = shared.TitleStyle
				s.RemoteList = l

				s.State = StateSelectingRemote
				return s, nil
			}
		}
	}

	var cmd tea.Cmd
	s.List, cmd = s.List.Update(msg)
	return s, cmd
}

func (s ManageScreen) updateSelectingRemote(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.State = StateList
			return s, nil
		case "enter":
			selected := s.RemoteList.SelectedItem()
			if selected == nil {
				return s, nil
			}
			switch item := selected.(type) {
			case remoteItem:
				// Iniciar upload con esta URL
				s.State = StateUploading
				s.Msg = fmt.Sprintf("Iniciando subida de '%s' a %s...", s.SelectedSkill.Name, item.url)
				return s, uploadSkillCmd(*s.SelectedSkill, item.url)
			case customURLItem:
				s.State = StateInputRemote
				s.RemoteInput.Focus()
				s.RemoteInput.SetValue("")
				return s, textinput.Blink
			}
		}
	}
	var cmd tea.Cmd
	s.RemoteList, cmd = s.RemoteList.Update(msg)
	return s, cmd
}

func (s ManageScreen) updateInputRemote(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.State = StateSelectingRemote
			return s, nil
		case "enter":
			url := s.RemoteInput.Value()
			if url != "" {
				s.State = StateUploading
				s.Msg = fmt.Sprintf("Iniciando subida de '%s' a %s...", s.SelectedSkill.Name, url)
				return s, uploadSkillCmd(*s.SelectedSkill, url)
			}
		}
	}
	var cmd tea.Cmd
	s.RemoteInput, cmd = s.RemoteInput.Update(msg)
	return s, cmd
}

func (s ManageScreen) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "right", "l", "tab":
			if s.ConfirmCursor == 0 {
				s.ConfirmCursor = 1
			} else {
				s.ConfirmCursor = 0
			}
		case "y", "Y":
			s.ConfirmCursor = 0
			if s.ToDelete != nil {
				return s, shared.DeleteSkillCmd(*s.ToDelete)
			}
		case "n", "N", "esc":
			s.State = StateList
			s.ToDelete = nil
			return s, nil
		case "enter":
			if s.ConfirmCursor == 0 {
				if s.ToDelete != nil {
					return s, shared.DeleteSkillCmd(*s.ToDelete)
				}
			} else {
				s.State = StateList
				s.ToDelete = nil
				return s, nil
			}
		}
	}
	return s, nil
}

func (s ManageScreen) updateUploading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case uploadSkillMsg:
		if msg.Err != nil {
			s.Msg = fmt.Sprintf("Error: %v", msg.Err)
			return s, nil
		}
		s.Msg = fmt.Sprintf("PR Creado: %s", msg.PRURL)
		return s, nil
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "enter" {
			s.State = StateList
			s.Msg = ""
			return s, nil
		}
	}
	return s, nil
}
