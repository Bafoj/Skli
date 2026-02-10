package config

import (
	"strings"

	"skli/internal/tui/shared"

	tea "github.com/charmbracelet/bubbletea"
)

// ConfigScreen es el modelo para la pantalla de configuración
type ConfigScreen struct {
	Cursor          int
	ConfigLocalPath string
	Remotes         []string
}

// NewConfigScreen crea una nueva pantalla de configuración
func NewConfigScreen(configLocalPath string, remotes []string) ConfigScreen {
	return ConfigScreen{
		ConfigLocalPath: configLocalPath,
		Remotes:         remotes,
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
			if s.Cursor > 0 {
				s.Cursor--
			}
		case "down", "j":
			if s.Cursor < 2 {
				s.Cursor++
			}
		case "enter":
			if s.Cursor == 0 {
				return s, func() tea.Msg { return shared.NavigateToManageRemotesMsg{} }
			} else if s.Cursor == 1 {
				return s, func() tea.Msg { return shared.NavigateToEditorMsg{} }
			} else if s.Cursor == 2 {
				return s, tea.Batch(
					shared.SaveConfigCmd(s.ConfigLocalPath, s.Remotes),
					func() tea.Msg { return shared.NavigateToDoneMsg{ConfigMode: true, LocalPath: s.ConfigLocalPath} },
				)
			}
		case "q", "esc":
			return s, func() tea.Msg { return shared.QuitMsg{} }
		}
	}

	return s, nil
}

func (s ConfigScreen) View() string {
	var b strings.Builder

	b.WriteString("Configuración Global:\n\n")

	// Opción 0: Remotes
	cursor := "  "
	if s.Cursor == 0 {
		cursor = "➜ "
	}
	line := "Gestionar Remotos"
	if s.Cursor == 0 {
		b.WriteString(shared.SelectedItemStyle.Render(cursor+line) + "\n")
	} else {
		b.WriteString(shared.ItemStyle.Render(cursor+line) + "\n")
	}

	// Opción 1: Local Path
	cursor = "  "
	if s.Cursor == 1 {
		cursor = "➜ "
	}
	line = "Local Path: " + shared.InfoStyle.Render(s.ConfigLocalPath)
	if s.Cursor == 1 {
		b.WriteString(shared.SelectedItemStyle.Render(cursor+line) + "\n")
	} else {
		b.WriteString(shared.ItemStyle.Render(cursor+line) + "\n")
	}

	// Opción 2: Confirmar
	cursor = "  "
	if s.Cursor == 2 {
		cursor = "➜ "
	}
	line = "Confirmar y Guardar"
	if s.Cursor == 2 {
		b.WriteString(shared.SelectedItemStyle.Render(cursor+line) + "\n")
	} else {
		b.WriteString(shared.ItemStyle.Render(cursor+line) + "\n")
	}

	b.WriteString(shared.HelpStyle.Render("\n↑/↓: navegar • enter: seleccionar • q: salir"))

	return b.String()
}

// UpdateConfigPath actualiza el path de la configuración
func (s *ConfigScreen) UpdateConfigPath(path string) {
	s.ConfigLocalPath = path
}

// UpdateRemotes actualiza la lista de remotes
func (s *ConfigScreen) UpdateRemotes(remotes []string) {
	s.Remotes = remotes
}
