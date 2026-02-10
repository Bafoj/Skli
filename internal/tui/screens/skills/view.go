package skills

import (
	"fmt"
)

func (s SkillsScreen) View() string {
	originalTitle := s.List.Title
	s.List.Title = fmt.Sprintf("%s (Pág. %d/%d)", originalTitle, s.List.Paginator.Page+1, s.List.Paginator.TotalPages)
	view := s.List.View()
	s.List.Title = originalTitle
	return view
}
