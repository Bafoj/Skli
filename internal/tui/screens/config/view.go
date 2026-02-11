package config

import (
	"strings"

	"skli/internal/tui/shared"
)

func (s ConfigScreen) View() string {
	var b strings.Builder

	b.WriteString("Global Configuration:\n\n")

	// Opción 0: Remotes
	cursor := "  "
	if s.Cursor == 0 {
		cursor = "➜ "
	}
	line := "Manage Remotes"
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
	line = "Confirm and Save"
	if s.Cursor == 2 {
		b.WriteString(shared.SelectedItemStyle.Render(cursor+line) + "\n")
	} else {
		b.WriteString(shared.ItemStyle.Render(cursor+line) + "\n")
	}

	b.WriteString(shared.HelpStyle.Render("\n↑/↓: navigate • enter: select • q: quit"))

	return b.String()
}
