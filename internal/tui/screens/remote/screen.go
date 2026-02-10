package remote

import (
	"fmt"
	"io"

	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	// Sub-views
	"skli/internal/tui/screens/remote/input"
	"skli/internal/tui/screens/remote/input_new"
	"skli/internal/tui/screens/remote/managing"
	"skli/internal/tui/screens/remote/selecting"
)

type State int

const (
	StateSelecting State = iota
	StateInput
	StateInputNew
	StateManaging
)

// RemoteScreen es el modelo para la pantalla de selección de remote
type RemoteScreen struct {
	State           State
	List            list.Model
	TextInput       textinput.Model
	Remotes         []string
	ConfigLocalPath string
	FromConfig      bool
}

// NewRemoteScreen crea una nueva pantalla de remote
func NewRemoteScreen(remotes []string, configLocalPath string, fromConfig bool) RemoteScreen {
	ti := textinput.New()
	ti.Placeholder = "https://github.com/usuario/repo.git"
	ti.CharLimit = 156
	ti.Width = 50

	if len(remotes) == 0 {
		ti.Focus()
		return RemoteScreen{
			State:           StateInput,
			TextInput:       ti,
			Remotes:         remotes,
			ConfigLocalPath: configLocalPath,
			FromConfig:      fromConfig,
		}
	}

	items := BuildRemoteListItems(remotes, customURLItem{})
	delegate := NewRemoteDelegate()
	l := list.New(items, delegate, 60, 14)
	l.Title = "Selecciona un repositorio remoto"
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("remote", "remotes")
	l.SetShowHelp(true)
	l.Styles.Title = shared.TitleStyle

	return RemoteScreen{
		State:           StateSelecting,
		List:            l,
		TextInput:       ti,
		Remotes:         remotes,
		ConfigLocalPath: configLocalPath,
		FromConfig:      fromConfig,
	}
}

// NewRemoteManageScreen crea una pantalla para gestionar remotes
func NewRemoteManageScreen(remotes []string, configLocalPath string) RemoteScreen {
	ti := textinput.New()
	ti.Placeholder = "https://github.com/usuario/repo.git"
	ti.CharLimit = 156
	ti.Width = 50

	items := BuildRemoteListItems(remotes, addNewItem{})
	delegate := NewRemoteDelegate()
	l := list.New(items, delegate, 60, 14)
	l.Title = "Gestionar Remotos"
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("remote", "remotes")
	l.SetShowHelp(true)
	l.Styles.Title = shared.TitleStyle

	return RemoteScreen{
		State:           StateManaging,
		List:            l,
		TextInput:       ti,
		Remotes:         remotes,
		ConfigLocalPath: configLocalPath,
		FromConfig:      true,
	}
}

func (s RemoteScreen) Init() tea.Cmd {
	if s.State == StateInput || s.State == StateInputNew {
		return textinput.Blink
	}
	return nil
}

func (s RemoteScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		s.List.SetSize(msg.Width, msg.Height-4)
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

func (s RemoteScreen) View() string {
	switch s.State {
	case StateSelecting:
		return selecting.View(s.List)
	case StateInput:
		return input.View(s.TextInput, len(s.Remotes) > 0)
	case StateInputNew:
		return input_new.View(s.TextInput)
	case StateManaging:
		return managing.View(s.List)
	}
	return ""
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

// Helpers and shared components for remote screen

type remoteItem struct {
	url         string
	displayName string
}

func (i remoteItem) Title() string {
	if i.displayName != "" {
		return i.displayName
	}
	return i.url
}
func (i remoteItem) Description() string { return "" }
func (i remoteItem) FilterValue() string { return i.url }

type customURLItem struct{}

func (i customURLItem) Title() string       { return "✏️  Custom URL..." }
func (i customURLItem) Description() string { return "Introduce una URL manualmente" }
func (i customURLItem) FilterValue() string { return "custom url" }

type addNewItem struct{}

func (i addNewItem) Title() string       { return "➕ Añadir nuevo..." }
func (i addNewItem) Description() string { return "Introduce una URL nueva" }
func (i addNewItem) FilterValue() string { return "añadir nuevo add new" }

type remoteDelegate struct {
	styles list.DefaultItemStyles
}

func NewRemoteDelegate() remoteDelegate {
	styles := list.NewDefaultItemStyles()
	styles.SelectedTitle = styles.SelectedTitle.
		Foreground(lipgloss.Color("#7D56F4")).
		BorderForeground(lipgloss.Color("#7D56F4"))

	return remoteDelegate{styles: styles}
}

func (d remoteDelegate) Height() int                             { return 1 }
func (d remoteDelegate) Spacing() int                            { return 0 }
func (d remoteDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d remoteDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(list.DefaultItem)
	if !ok {
		return
	}
	title := i.Title()
	if index == m.Index() {
		fmt.Fprint(w, d.styles.SelectedTitle.Render("➜ "+title))
	} else {
		fmt.Fprint(w, d.styles.NormalTitle.Render("  "+title))
	}
}

func BuildRemoteListItems(remotes []string, extraItem list.Item) []list.Item {
	items := make([]list.Item, 0, len(remotes)+1)
	for _, r := range remotes {
		items = append(items, remoteItem{url: r})
	}
	items = append(items, extraItem)
	return items
}
