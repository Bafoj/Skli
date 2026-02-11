package managing

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
)

func View(l list.Model) string {
	originalTitle := l.Title
	l.Title = fmt.Sprintf("%s (Page %d/%d)", originalTitle, l.Paginator.Page+1, l.Paginator.TotalPages)
	view := l.View()
	l.Title = originalTitle
	return view
}
