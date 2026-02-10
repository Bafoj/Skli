package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// remoteItem implementa list.DefaultItem para un remote
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

// customURLItem es el item especial "Custom URL..."
type customURLItem struct{}

func (i customURLItem) Title() string       { return "✏️  Custom URL..." }
func (i customURLItem) Description() string { return "Introduce una URL manualmente" }
func (i customURLItem) FilterValue() string { return "custom url" }

// addNewItem es el item especial "Añadir nuevo..."
type addNewItem struct{}

func (i addNewItem) Title() string       { return "➕ Añadir nuevo..." }
func (i addNewItem) Description() string { return "Introduce una URL nueva" }
func (i addNewItem) FilterValue() string { return "añadir nuevo add new" }

// remoteDelegate es un delegate personalizado para remotes
type remoteDelegate struct {
	styles list.DefaultItemStyles
}

func newRemoteDelegate() remoteDelegate {
	styles := list.NewDefaultItemStyles()
	styles.SelectedTitle = styles.SelectedTitle.
		Foreground(lipgloss.Color("#7D56F4")).
		BorderForeground(lipgloss.Color("#7D56F4"))
	styles.SelectedDesc = styles.SelectedDesc.
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

// RemoteScreenState representa el sub-estado de la pantalla de remote
type RemoteScreenState int

const (
	RemoteStateSelecting RemoteScreenState = iota
	RemoteStateInput
	RemoteStateInputNew
	RemoteStateManaging
)

// RemoteScreen es el modelo para la pantalla de selección de remote
type RemoteScreen struct {
	state           RemoteScreenState
	list            list.Model
	textInput       textinput.Model
	remotes         []string
	configLocalPath string
	fromConfig      bool
}

func buildRemoteListItems(remotes []string, extraItem list.Item) []list.Item {
	items := make([]list.Item, 0, len(remotes)+1)
	for _, r := range remotes {
		items = append(items, remoteItem{url: r})
	}
	items = append(items, extraItem)
	return items
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
			state:           RemoteStateInput,
			textInput:       ti,
			remotes:         remotes,
			configLocalPath: configLocalPath,
			fromConfig:      fromConfig,
		}
	}

	items := buildRemoteListItems(remotes, customURLItem{})
	delegate := newRemoteDelegate()
	l := list.New(items, delegate, 60, 14)
	l.Title = "Selecciona un repositorio remoto"
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("remote", "remotes")
	l.SetShowHelp(true)
	l.Styles.Title = TitleStyle

	return RemoteScreen{
		state:           RemoteStateSelecting,
		list:            l,
		textInput:       ti,
		remotes:         remotes,
		configLocalPath: configLocalPath,
		fromConfig:      fromConfig,
	}
}

// NewRemoteManageScreen crea una pantalla para gestionar remotes
func NewRemoteManageScreen(remotes []string, configLocalPath string) RemoteScreen {
	ti := textinput.New()
	ti.Placeholder = "https://github.com/usuario/repo.git"
	ti.CharLimit = 156
	ti.Width = 50

	items := buildRemoteListItems(remotes, addNewItem{})
	delegate := newRemoteDelegate()
	l := list.New(items, delegate, 60, 14)
	l.Title = "Gestionar Remotos"
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("remote", "remotes")
	l.SetShowHelp(true)
	l.Styles.Title = TitleStyle

	return RemoteScreen{
		state:           RemoteStateManaging,
		list:            l,
		textInput:       ti,
		remotes:         remotes,
		configLocalPath: configLocalPath,
		fromConfig:      true,
	}
}

func (s RemoteScreen) Init() tea.Cmd {
	if s.state == RemoteStateInput || s.state == RemoteStateInputNew {
		return textinput.Blink
	}
	return nil
}

func (s RemoteScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global window resize
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		s.list.SetSize(msg.Width, msg.Height-4)
	}

	switch s.state {
	case RemoteStateInput:
		return s.updateInput(msg)
	case RemoteStateInputNew:
		return s.updateInputNew(msg)
	case RemoteStateSelecting:
		return s.updateSelecting(msg)
	case RemoteStateManaging:
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
			url := s.textInput.Value()
			if url == "" {
				return s, nil
			}
			return s, func() tea.Msg {
				return NavigateToScanningMsg{URL: url}
			}
		case "esc":
			if len(s.remotes) > 0 {
				s.state = RemoteStateSelecting
				return s, nil
			}
			return s, func() tea.Msg { return QuitMsg{} }
		}
	}
	s.textInput, cmd = s.textInput.Update(msg)
	return s, cmd
}

func (s RemoteScreen) updateInputNew(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			url := s.textInput.Value()
			if url == "" {
				return s, nil
			}
			s.remotes = append(s.remotes, url)
			s.state = RemoteStateManaging
			s.textInput.Reset()
			items := buildRemoteListItems(s.remotes, addNewItem{})
			cmdList := s.list.SetItems(items)
			return s, tea.Batch(
				cmdList,
				SaveConfigCmd(s.configLocalPath, s.remotes),
				func() tea.Msg { return RemotesUpdatedMsg{Remotes: s.remotes} },
			)
		case "esc":
			s.state = RemoteStateManaging
			s.textInput.Reset()
			return s, nil
		}
	}
	s.textInput, cmd = s.textInput.Update(msg)
	return s, cmd
}

func (s RemoteScreen) updateSelecting(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selected := s.list.SelectedItem()
			if selected == nil {
				return s, nil
			}
			switch item := selected.(type) {
			case remoteItem:
				url := item.url
				return s, func() tea.Msg {
					return NavigateToScanningMsg{URL: url}
				}
			case customURLItem:
				s.state = RemoteStateInput
				s.textInput.Focus()
				return s, textinput.Blink
			}
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

func (s RemoteScreen) updateManaging(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s, func() tea.Msg { return NavigateToConfigMsg{} }
		case "enter":
			selected := s.list.SelectedItem()
			if selected == nil {
				return s, nil
			}
			if _, ok := selected.(addNewItem); ok {
				s.state = RemoteStateInputNew
				s.textInput.Focus()
				s.textInput.SetValue("")
				return s, textinput.Blink
			}
		case "d", "backspace":
			idx := s.list.Index()
			if idx < len(s.remotes) {
				s.remotes = append(s.remotes[:idx], s.remotes[idx+1:]...)
				items := buildRemoteListItems(s.remotes, addNewItem{})
				cmd := s.list.SetItems(items)
				return s, tea.Batch(
					cmd,
					SaveConfigCmd(s.configLocalPath, s.remotes),
					func() tea.Msg { return RemotesUpdatedMsg{Remotes: s.remotes} },
				)
			}
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

func (s RemoteScreen) View() string {
	switch s.state {
	case RemoteStateSelecting, RemoteStateManaging:
		originalTitle := s.list.Title
		s.list.Title = fmt.Sprintf("%s (Pág. %d/%d)", originalTitle, s.list.Paginator.Page+1, s.list.Paginator.TotalPages)
		view := s.list.View()
		s.list.Title = originalTitle // Restaurar para que no se acumule
		return view
	case RemoteStateInputNew:
		return "Introduce la URL del nuevo repositorio:\n\n" + s.textInput.View() + "\n" +
			HelpStyle.Render("\nenter: guardar • esc: cancelar")
	case RemoteStateInput:
		return "Introduce la URL del repositorio Git remoto:\n\n" + s.textInput.View() + "\n" +
			HelpStyle.Render("\nenter: continuar • esc/q: salir")
	}
	return ""
}
