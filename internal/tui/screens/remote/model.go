package remote

import (
	"skli/internal/tui/screens/remote/delegates"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
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
	ti.Placeholder = "https://github.com/user/repo.git"
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
	delegate := delegates.NewRemoteDelegate()
	l := list.New(items, delegate, 60, 14)
	l.Title = "Select a remote repository"
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
	ti.Placeholder = "https://github.com/user/repo.git"
	ti.CharLimit = 156
	ti.Width = 50

	items := BuildRemoteListItems(remotes, addNewItem{})
	delegate := delegates.NewRemoteDelegate()
	l := list.New(items, delegate, 60, 14)
	l.Title = "Manage Remotes"
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
func (i customURLItem) Description() string { return "Enter a URL manually" }
func (i customURLItem) FilterValue() string { return "custom url" }

type addNewItem struct{}

func (i addNewItem) Title() string       { return "➕ Add New..." }
func (i addNewItem) Description() string { return "Enter a new URL" }
func (i addNewItem) FilterValue() string { return "add new" }

func BuildRemoteListItems(remotes []string, extraItem list.Item) []list.Item {
	items := make([]list.Item, 0, len(remotes)+1)
	for _, r := range remotes {
		items = append(items, remoteItem{url: r})
	}
	items = append(items, extraItem)
	return items
}
