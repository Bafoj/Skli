package editor

import (
	"fmt"
	"skli/internal/tui/shared"
)

func (s EditorScreen) View() string {
	if s.State == StateInputCustom {
		return "Enter the custom path for skills:\n\n" + s.TextInput.View() + "\n" +
			shared.HelpStyle.Render("\nenter: confirm • esc: back")
	}

	originalTitle := s.List.Title
	s.List.Title = fmt.Sprintf("%s (Page %d/%d)", originalTitle, s.List.Paginator.Page+1, s.List.Paginator.TotalPages)
	view := s.List.View()
	s.List.Title = originalTitle
	return view
}
