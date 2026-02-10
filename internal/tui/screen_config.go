package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ConfigScreen es el modelo para la pantalla de configuración
type ConfigScreen struct {
	cursor          int
	configLocalPath string
	remotes         []string
}

// NewConfigScreen crea una nueva pantalla de configuración
func NewConfigScreen(configLocalPath string, remotes []string) ConfigScreen {
	return ConfigScreen{
		configLocalPath: configLocalPath,
		remotes:         remotes,
	}
}

func (s ConfigScreen) Init() tea.Cmd {
	return nil
}

func (s ConfigScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < 2 {
				s.cursor++
			}
		case "enter":
			if s.cursor == 0 {
				return s, func() tea.Msg { return NavigateToManageRemotesMsg{} }
			} else if s.cursor == 1 {
				return s, func() tea.Msg { return NavigateToEditorMsg{} }
			} else if s.cursor == 2 {
				return s, tea.Batch(
					SaveConfigCmd(s.configLocalPath, s.remotes),
					func() tea.Msg { return NavigateToDoneMsg{ConfigMode: true, LocalPath: s.configLocalPath} },
				)
			}
		case "q", "esc":
			return s, func() tea.Msg { return QuitMsg{} }
		}
	}

	return s, nil
}

func (s ConfigScreen) View() string {
	var b strings.Builder

	b.WriteString("Configuración Global:\n\n")

	// Opción 0: Remotes
	cursor := "  "
	if s.cursor == 0 {
		cursor = "➜ "
	}
	line := "Gestionar Remotos"
	if s.cursor == 0 {
		b.WriteString(SelectedItemStyle.Render(cursor+line) + "\n")
	} else {
		b.WriteString(ItemStyle.Render(cursor+line) + "\n")
	}

	// Opción 1: Local Path
	cursor = "  "
	if s.cursor == 1 {
		cursor = "➜ "
	}
	line = "Local Path: " + InfoStyle.Render(s.configLocalPath)
	if s.cursor == 1 {
		b.WriteString(SelectedItemStyle.Render(cursor+line) + "\n")
	} else {
		b.WriteString(ItemStyle.Render(cursor+line) + "\n")
	}

	// Opción 2: Confirmar
	cursor = "  "
	if s.cursor == 2 {
		cursor = "➜ "
	}
	line = "Confirmar y Guardar"
	if s.cursor == 2 {
		b.WriteString(SelectedItemStyle.Render(cursor+line) + "\n")
	} else {
		b.WriteString(ItemStyle.Render(cursor+line) + "\n")
	}

	b.WriteString(HelpStyle.Render("\n↑/↓: navegar • enter: seleccionar • q: salir"))

	return b.String()
}

// UpdateConfigPath actualiza el path de la configuración
func (s *ConfigScreen) UpdateConfigPath(path string) {
	s.configLocalPath = path
}

// UpdateRemotes actualiza la lista de remotes
func (s *ConfigScreen) UpdateRemotes(remotes []string) {
	s.remotes = remotes
}
