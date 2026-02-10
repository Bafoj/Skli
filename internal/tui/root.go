package tui

import (
	"strings"

	"skli/internal/tui/screens/config"
	"skli/internal/tui/screens/editor"
	"skli/internal/tui/screens/manage"
	"skli/internal/tui/screens/progress"
	"skli/internal/tui/screens/remote"
	"skli/internal/tui/screens/scanning"
	"skli/internal/tui/screens/skills"
	"skli/internal/tui/shared"

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

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Eventos globales
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		// Solo q para salir en ciertas pantallas, las pantallas lo manejan
		if msg.String() == "q" {
			// Verificar si es una pantalla que acepta q para salir
			switch m.activeScreen.(type) {
			case remote.RemoteScreen, skills.SkillsScreen, editor.EditorScreen, config.ConfigScreen:
				m.quitting = true
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		// También pasar al screen activo
		var cmd tea.Cmd
		m.activeScreen, cmd = m.activeScreen.Update(msg)
		return m, cmd

	// Mensajes de navegación
	case shared.QuitMsg:
		m.quitting = true
		return m, tea.Quit

	case shared.NavigateToInputRemoteMsg:
		m.activeScreen = remote.NewRemoteScreen(m.remotes, m.configLocalPath, false)
		return m, m.activeScreen.Init()

	case shared.NavigateToScanningMsg:
		m.activeScreen = scanning.NewScanningScreen(msg.URL, m.skillsRoot)
		return m, m.activeScreen.Init()

	case shared.NavigateToSkillsMsg:
		m.activeScreen = skills.NewSkillsScreen(msg.Skills, msg.TempDir, msg.RemoteURL, msg.SkillsRoot, msg.CommitHash, m.configLocalPath)
		return m, m.activeScreen.Init()

	case shared.NavigateToEditorMsg:
		if len(msg.Skills) > 0 {
			// Desde skills selection
			m.activeScreen = editor.NewEditorScreen(msg.Skills, msg.TempDir, msg.RemoteURL, msg.SkillsRoot, msg.CommitHash, false, m.remotes)
		} else {
			// Desde config
			m.activeScreen = editor.NewEditorScreenForConfig(m.configLocalPath, m.remotes)
		}
		return m, m.activeScreen.Init()

	case shared.NavigateToConfigMsg:
		m.activeScreen = config.NewConfigScreen(m.configLocalPath, m.remotes)
		return m, m.activeScreen.Init()

	case shared.NavigateToManageRemotesMsg:
		m.activeScreen = remote.NewRemoteManageScreen(m.remotes, m.configLocalPath)
		return m, m.activeScreen.Init()

	case shared.NavigateToProgressMsg:
		screen, cmd := progress.NewProgressScreenDownloading(msg.TempDir, msg.RemoteURL, msg.SkillsRoot, msg.ConfigLocalPath, msg.CommitHash, msg.Selected)
		m.activeScreen = screen
		m.configLocalPath = msg.ConfigLocalPath
		return m, cmd

	case shared.NavigateToDoneMsg:
		m.activeScreen = progress.NewDoneScreen(msg.ConfigMode, msg.LocalPath)
		return m, m.activeScreen.Init()

	case shared.NavigateToErrorMsg:
		// TODO: Crear ErrorScreen en package dedicado si es necesario, por ahora NewErrorScreen está en progress
		m.activeScreen = progress.NewErrorScreen(msg.Err)
		return m, m.activeScreen.Init()

	case shared.NavigateToManageMsg:
		screen, cmd := manage.NewManageScreen(m.remotes)
		m.activeScreen = screen
		return m, cmd

	case shared.RemotesUpdatedMsg:
		m.remotes = msg.Remotes
		// Continuar con update normal

	case shared.ConfigSavedMsg:
		// Config guardado, solo continuar
	}

	// Delegar al screen activo
	var cmd tea.Cmd
	m.activeScreen, cmd = m.activeScreen.Update(msg)
	return m, cmd
}

func (m RootModel) View() string {
	if m.quitting {
		return "¡Hasta luego!\n"
	}

	var s strings.Builder
	s.WriteString(m.activeScreen.View())

	return s.String()
}
