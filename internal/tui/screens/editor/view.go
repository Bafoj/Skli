package editor

import (
	"fmt"
)

func (s EditorScreen) View() string {
	originalTitle := s.List.Title
	s.List.Title = fmt.Sprintf("%s (Page %d/%d)", originalTitle, s.List.Paginator.Page+1, s.List.Paginator.TotalPages)
	view := s.List.View()
	s.List.Title = originalTitle
	return view
}
