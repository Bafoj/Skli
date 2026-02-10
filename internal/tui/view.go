package tui

import (
	"strings"
)

func (m RootModel) View() string {
	if m.quitting {
		return "¡Hasta luego!\n"
	}

	var s strings.Builder
	s.WriteString(m.activeScreen.View())

	return s.String()
}
