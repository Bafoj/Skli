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
)

func (m Model) View() string {
	if m.Quitting {
		return "¡Hasta luego!\n"
	}

	var s strings.Builder

	// Header
	s.WriteString(titleStyle.Render("skli - Skills Management") + "\n\n")

	switch m.State {
	case StateInputRemote:
		s.WriteString("Introduce la URL del repositorio Git remoto:\n\n")
		s.WriteString(m.TextInput.View() + "\n")
		s.WriteString(helpStyle.Render("\nenter: continuar • esc/q: salir"))

	case StateScanning:
		s.WriteString(fmt.Sprintf("%s Escaneando el repositorio remoto...", m.Spinner.View()))

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

	case StateDownloading:
		s.WriteString(fmt.Sprintf("%s Instalando skills seleccionadas...", m.Spinner.View()))

	case StateDone:
		s.WriteString(successStyle.Render("✔ ¡Skills instaladas correctamente en ./skills/!"))
		s.WriteString(helpStyle.Render("\nPresiona cualquier tecla para salir"))

	case StateError:
		s.WriteString(errorStyle.Render(fmt.Sprintf("✘ Error: %s", m.ErrorMessage)))
		s.WriteString(helpStyle.Render("\nPresiona 'r' para reintentar o 'q' para salir"))
	}

	return s.String()
}
