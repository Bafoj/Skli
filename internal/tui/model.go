package tui

import (
	"skli/internal/tui/screens/config"
	"skli/internal/tui/screens/manage"
	"skli/internal/tui/screens/remote"
	"skli/internal/tui/screens/scanning"

	tea "github.com/charmbracelet/bubbletea"
)

// RootModel es el modelo principal que actúa como router
type RootModel struct {
	activeScreen    tea.Model
	configLocalPath string
	remotes         []string
	skillsRoot      string
	quitting        bool
	windowWidth     int
	windowHeight    int
}

// NewRootModel crea el modelo principal
func NewRootModel(initialURL, skillsRoot, configLocalPath string, configMode, manageMode bool, remotes []string) RootModel {
	var activeScreen tea.Model

	if manageMode {
		activeScreen, _ = manage.NewManageScreen(remotes)
	} else if configMode {
		activeScreen = config.NewConfigScreen(configLocalPath, remotes)
	} else if initialURL != "" {
		activeScreen = scanning.NewScanningScreen(initialURL, skillsRoot)
	} else if len(remotes) > 0 {
		activeScreen = remote.NewRemoteScreen(remotes, configLocalPath, false)
	} else {
		activeScreen = remote.NewRemoteScreen(remotes, configLocalPath, false)
	}

	return RootModel{
		activeScreen:    activeScreen,
		configLocalPath: configLocalPath,
		remotes:         remotes,
		skillsRoot:      skillsRoot,
	}
}

func (m RootModel) Init() tea.Cmd {
	return m.activeScreen.Init()
}
