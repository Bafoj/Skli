package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Italic(true).
				PaddingLeft(6)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Italic(true).
			MarginTop(1)

	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			MarginTop(1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true).
			MarginTop(1)

	dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))

	infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
)

func (m Model) View() string {
	if m.Quitting {
		return "¡Hasta luego!\n"
	}

	var s strings.Builder

	// Header
	s.WriteString(titleStyle.Render("skli - Skills Management") + "\n\n")

	switch m.State {
	case StateSelectingRemote:
		s.WriteString("Selecciona un repositorio remoto:\n\n")

		// Remotes
		for i, remote := range m.Remotes {
			cursor := "  "
			if m.RemoteCursor == i {
				cursor = "➜ "
			}

			if m.RemoteCursor == i {
				s.WriteString(selectedItemStyle.Render(cursor+remote) + "\n")
			} else {
				s.WriteString(itemStyle.Render(cursor+remote) + "\n")
			}
		}

		// Custom
		cursor := "  "
		if m.RemoteCursor == len(m.Remotes) {
			cursor = "➜ "
		}
		line := "Custom URL..."
		if m.RemoteCursor == len(m.Remotes) {
			s.WriteString(selectedItemStyle.Render(cursor+line) + "\n")
		} else {
			s.WriteString(itemStyle.Render(cursor+line) + "\n")
		}

		s.WriteString(helpStyle.Render("\n↑/↓: navegar • enter: seleccionar • q: salir"))

	case StateManageRemotes:
		s.WriteString("Gestionar Remotos:\n\n")
		if len(m.Remotes) == 0 {
			s.WriteString(dimStyle.Render("  (No hay remotos configurados)\n"))
		}

		// Remotes
		for i, remote := range m.Remotes {
			cursor := "  "
			if m.RemoteCursor == i {
				cursor = "➜ "
			}

			if m.RemoteCursor == i {
				s.WriteString(selectedItemStyle.Render(cursor+remote) + "\n")
			} else {
				s.WriteString(itemStyle.Render(cursor+remote) + "\n")
			}
		}

		// Add New
		cursor := "  "
		if m.RemoteCursor == len(m.Remotes) {
			cursor = "➜ "
		}
		line := "Añadir nuevo..."
		if m.RemoteCursor == len(m.Remotes) {
			s.WriteString(selectedItemStyle.Render(cursor+line) + "\n")
		} else {
			s.WriteString(itemStyle.Render(cursor+line) + "\n")
		}

		s.WriteString(helpStyle.Render("\n↑/↓: navegar • enter: añadir • d/bksp: borrar • esc: volver"))

	case StateInputNewRemote:
		s.WriteString("Introduce la URL del nuevo repositorio:\n\n")
		s.WriteString(m.TextInput.View() + "\n")
		s.WriteString(helpStyle.Render("\nenter: guardar • esc: cancelar"))

	case StateInputRemote:
		s.WriteString("Introduce la URL del repositorio Git remoto:\n\n")
		s.WriteString(m.TextInput.View() + "\n")
		s.WriteString(helpStyle.Render("\nenter: continuar • esc/q: salir"))

	case StateScanning:
		s.WriteString(fmt.Sprintf("%s Escaneando el repositorio remoto...", m.Spinner.View()))

	case StateConfigMenu:
		s.WriteString("Configuración Global:\n\n")

		// Opción 0: Remotes
		cursor := "  "
		if m.ConfigCursor == 0 {
			cursor = "➜ "
		}
		line := "Gestionar Remotos"
		if m.ConfigCursor == 0 {
			s.WriteString(selectedItemStyle.Render(cursor+line) + "\n")
		} else {
			s.WriteString(itemStyle.Render(cursor+line) + "\n")
		}

		// Opción 1: Local Path
		cursor = "  "
		if m.ConfigCursor == 1 {
			cursor = "➜ "
		}
		line = "Local Path: " + infoStyle.Render(m.ConfigLocalPath)
		if m.ConfigCursor == 1 {
			s.WriteString(selectedItemStyle.Render(cursor+line) + "\n")
		} else {
			s.WriteString(itemStyle.Render(cursor+line) + "\n")
		}

		// Opción 2: Confirmar
		cursor = "  "
		if m.ConfigCursor == 2 {
			cursor = "➜ "
		}
		line = "Confirmar y Guardar"
		if m.ConfigCursor == 2 {
			s.WriteString(selectedItemStyle.Render(cursor+line) + "\n")
		} else {
			s.WriteString(itemStyle.Render(cursor+line) + "\n")
		}

		s.WriteString(helpStyle.Render("\n↑/↓: navegar • enter: seleccionar • q: salir"))

	case StateSelectingSkills:
		s.WriteString(fmt.Sprintf("Encontrados %d skills. Selecciona los que quieres instalar:\n\n", len(m.Skills)))

		// Calculamos el área disponible para la lista (reservamos espacio para cabecera y pie)
		maxItems := m.WindowHeight - 12
		if maxItems < 5 {
			maxItems = 5 // Mínimo de items a mostrar
		}

		start, end := 0, len(m.Skills)
		if len(m.Skills) > maxItems {
			// Lógica de scroll: centrar el cursor si es posible
			half := maxItems / 2
			start = m.Cursor - half
			if start < 0 {
				start = 0
			}
			end = start + maxItems
			if end > len(m.Skills) {
				end = len(m.Skills)
				start = end - maxItems
			}
		}

		for i := start; i < end; i++ {
			skill := m.Skills[i]
			cursor := " "
			if m.Cursor == i {
				cursor = "➜"
			}

			checked := "[ ]"
			if skill.Selected {
				checked = "[x]"
			}

			line := fmt.Sprintf("%s %s %s", cursor, checked, skill.Info.Name)
			if m.Cursor == i {
				s.WriteString(selectedItemStyle.Render(line) + "\n")
				// Mostrar descripción truncada debajo del skill seleccionado
				if skill.Info.Description != "" {
					desc := skill.Info.Description
					if len(desc) > 80 {
						desc = desc[:77] + "..."
					}
					s.WriteString(descriptionStyle.Render(desc) + "\n")
				}
			} else {
				s.WriteString(itemStyle.Render(line) + "\n")
			}
		}

		if len(m.Skills) > maxItems {
			s.WriteString(statusBarStyle.Render(fmt.Sprintf("\n--- Mostrando %d-%d de %d ---", start+1, end, len(m.Skills))))
		}

		s.WriteString(helpStyle.Render("\n↑/↓: navegar • espacio: marcar • enter: instalar • q: salir"))

	case StateSelectingEditor:
		s.WriteString("Selecciona tu editor:\n\n")
		for i, editor := range Editors {
			cursor := "  "
			if m.EditorCursor == i {
				cursor = "➜ "
			}

			line := editor.Name
			if editor.Path != "" {
				line = fmt.Sprintf("%-12s %s", editor.Name, dimStyle.Render("("+editor.Path+")"))
			}

			if m.EditorCursor == i {
				s.WriteString(selectedItemStyle.Render(cursor+line) + "\n")
			} else {
				s.WriteString(itemStyle.Render(cursor+line) + "\n")
			}
		}
		s.WriteString(helpStyle.Render("\n↑/↓: navegar • enter: seleccionar • esc: volver • q: salir"))

	case StateDownloading:
		s.WriteString(fmt.Sprintf("%s Instalando skills seleccionadas en %s...", m.Spinner.View(), infoStyle.Render(m.ConfigLocalPath)))

	case StateDone:
		if m.ConfigMode {
			s.WriteString(successStyle.Render("✔ ¡Configuración guardada correctamente!"))
		} else {
			s.WriteString(successStyle.Render(fmt.Sprintf("✔ ¡Skills instaladas correctamente en ./%s/!", m.ConfigLocalPath)))
		}
		s.WriteString(helpStyle.Render("\nPresiona cualquier tecla para salir"))

	case StateError:
		s.WriteString(errorStyle.Render(fmt.Sprintf("✘ Error: %s", m.ErrorMessage)))
		s.WriteString(helpStyle.Render("\nPresiona 'r' para reintentar o 'q' para salir"))
	}

	return s.String()
}
