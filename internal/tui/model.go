package tui

import (
	"skli/internal/tui/screens/config"
	"skli/internal/tui/screens/manage"
	"skli/internal/tui/screens/remote"
	"skli/internal/tui/screens/scanning"

	tea "github.com/charmbracelet/bubbletea"
)

// RootModel es el modelo principal que actua como router
type RootModel struct {
	activeScreen    tea.Model
	configLocalPath string
	remotes         []string
	skillsRoot      string
	manageMode      manage.Mode
	quitting        bool
	windowWidth     int
	windowHeight    int
}

// NewRootModel crea el modelo principal
func NewRootModel(initialURL, skillsRoot, configLocalPath string, configMode bool, manageMode manage.Mode, remotes []string) RootModel {
	var activeScreen tea.Model

	switch {
	case manageMode != manage.ModeNone:
		activeScreen, _ = manage.NewManageScreen(remotes, manageMode)
	case configMode:
		activeScreen = config.NewConfigScreen(configLocalPath, remotes)
	case initialURL != "":
		activeScreen = scanning.NewScanningScreen(initialURL, skillsRoot)
	default:
		activeScreen = remote.NewRemoteScreen(remotes, configLocalPath, false)
	}

	return RootModel{
		activeScreen:    activeScreen,
		configLocalPath: configLocalPath,
		remotes:         remotes,
		skillsRoot:      skillsRoot,
		manageMode:      manageMode,
	}
}

func (m RootModel) Init() tea.Cmd {
	return m.activeScreen.Init()
}
