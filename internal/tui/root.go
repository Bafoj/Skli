package tui

import (
	"strings"

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
		activeScreen, _ = NewManageScreen()
	} else if configMode {
		activeScreen = NewConfigScreen(configLocalPath, remotes)
	} else if initialURL != "" {
		activeScreen = NewScanningScreen(initialURL, skillsRoot)
	} else if len(remotes) > 0 {
		activeScreen = NewRemoteScreen(remotes, configLocalPath, false)
	} else {
		activeScreen = NewRemoteScreen(remotes, configLocalPath, false)
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
			case RemoteScreen, SkillsScreen, EditorScreen, ConfigScreen:
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
	case QuitMsg:
		m.quitting = true
		return m, tea.Quit

	case NavigateToInputRemoteMsg:
		m.activeScreen = NewRemoteScreen(m.remotes, m.configLocalPath, false)
		return m, m.activeScreen.Init()

	case NavigateToScanningMsg:
		m.activeScreen = NewScanningScreen(msg.URL, m.skillsRoot)
		return m, m.activeScreen.Init()

	case NavigateToSkillsMsg:
		m.activeScreen = NewSkillsScreen(msg.Skills, msg.TempDir, msg.RemoteURL, msg.SkillsRoot, msg.CommitHash, m.configLocalPath)
		return m, m.activeScreen.Init()

	case NavigateToEditorMsg:
		if len(msg.Skills) > 0 {
			// Desde skills selection
			m.activeScreen = NewEditorScreen(msg.Skills, msg.TempDir, msg.RemoteURL, msg.SkillsRoot, msg.CommitHash, false, m.remotes)
		} else {
			// Desde config
			m.activeScreen = NewEditorScreenForConfig(m.configLocalPath, m.remotes)
		}
		return m, m.activeScreen.Init()

	case NavigateToConfigMsg:
		m.activeScreen = NewConfigScreen(m.configLocalPath, m.remotes)
		return m, m.activeScreen.Init()

	case NavigateToManageRemotesMsg:
		m.activeScreen = NewRemoteManageScreen(m.remotes, m.configLocalPath)
		return m, m.activeScreen.Init()

	case NavigateToProgressMsg:
		screen, cmd := NewProgressScreenDownloading(msg.TempDir, msg.RemoteURL, msg.SkillsRoot, msg.ConfigLocalPath, msg.CommitHash, msg.Selected)
		m.activeScreen = screen
		m.configLocalPath = msg.ConfigLocalPath
		return m, cmd

	case NavigateToDoneMsg:
		m.activeScreen = NewDoneScreen(msg.ConfigMode, msg.LocalPath)
		return m, m.activeScreen.Init()

	case NavigateToErrorMsg:
		m.activeScreen = NewErrorScreen(msg.Err)
		return m, m.activeScreen.Init()

	case NavigateToManageMsg:
		screen, cmd := NewManageScreen()
		m.activeScreen = screen
		return m, cmd

	case RemotesUpdatedMsg:
		m.remotes = msg.Remotes
		// Continuar con update normal

	case ConfigSavedMsg:
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
	// s.WriteString(TitleStyle.Render("skli - Skills Management") + "\n\n")
	s.WriteString(m.activeScreen.View())

	return s.String()
}
