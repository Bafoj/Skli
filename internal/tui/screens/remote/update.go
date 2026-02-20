package remote

import (
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (s RemoteScreen) Init() tea.Cmd {
	if s.State == StateInput || s.State == StateInputNew {
		return textinput.Blink
	}
	return nil
}

func (s RemoteScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		// List is only initialized when there are remotes to display.
		// In StateInput there is no list, so skip SetSize to avoid nil panic.
		if s.State != StateInput {
			s.List.SetSize(msg.Width, msg.Height-4)
		}
	}

	switch s.State {
	case StateInput:
		return s.updateInput(msg)
	case StateInputNew:
		return s.updateInputNew(msg)
	case StateSelecting:
		return s.updateSelecting(msg)
	case StateManaging:
		return s.updateManaging(msg)
	}
	return s, nil
}

func (s RemoteScreen) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			url := s.TextInput.Value()
			if url == "" {
				return s, nil
			}
			return s, func() tea.Msg {
				return shared.NavigateToScanningMsg{URL: url}
			}
		case "esc":
			if len(s.Remotes) > 0 {
				s.State = StateSelecting
				return s, nil
			}
			return s, func() tea.Msg { return shared.QuitMsg{} }
		}
	}
	s.TextInput, cmd = s.TextInput.Update(msg)
	return s, cmd
}

func (s RemoteScreen) updateInputNew(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			url := s.TextInput.Value()
			if url == "" {
				return s, nil
			}
			s.Remotes = append(s.Remotes, url)
			s.State = StateManaging
			s.TextInput.Reset()
			items := BuildRemoteListItems(s.Remotes, addNewItem{})
			cmdList := s.List.SetItems(items)
			return s, tea.Batch(
				cmdList,
				shared.SaveConfigCmd(s.ConfigLocalPath, s.Remotes),
				func() tea.Msg { return shared.RemotesUpdatedMsg{Remotes: s.Remotes} },
			)
		case "esc":
			s.State = StateManaging
			s.TextInput.Reset()
			return s, nil
		}
	}
	s.TextInput, cmd = s.TextInput.Update(msg)
	return s, cmd
}

func (s RemoteScreen) updateSelecting(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selected := s.List.SelectedItem()
			if selected == nil {
				return s, nil
			}
			switch item := selected.(type) {
			case remoteItem:
				url := item.url
				return s, func() tea.Msg {
					return shared.NavigateToScanningMsg{URL: url}
				}
			case customURLItem:
				s.State = StateInput
				s.TextInput.Focus()
				return s, textinput.Blink
			}
		}
	}

	var cmd tea.Cmd
	s.List, cmd = s.List.Update(msg)
	return s, cmd
}

func (s RemoteScreen) updateManaging(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s, func() tea.Msg { return shared.NavigateToConfigMsg{} }
		case "enter":
			selected := s.List.SelectedItem()
			if selected == nil {
				return s, nil
			}
			if _, ok := selected.(addNewItem); ok {
				s.State = StateInputNew
				s.TextInput.Focus()
				s.TextInput.SetValue("")
				return s, textinput.Blink
			}
		case "d", "backspace":
			idx := s.List.Index()
			if idx < len(s.Remotes) {
				s.Remotes = append(s.Remotes[:idx], s.Remotes[idx+1:]...)
				items := BuildRemoteListItems(s.Remotes, addNewItem{})
				cmd := s.List.SetItems(items)
				return s, tea.Batch(
					cmd,
					shared.SaveConfigCmd(s.ConfigLocalPath, s.Remotes),
					func() tea.Msg { return shared.RemotesUpdatedMsg{Remotes: s.Remotes} },
				)
			}
		}
	}

	var cmd tea.Cmd
	s.List, cmd = s.List.Update(msg)
	return s, cmd
}
