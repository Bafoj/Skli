package delegates

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditorDelegate struct {
	styles list.DefaultItemStyles
}

func NewEditorDelegate() EditorDelegate {
	styles := list.NewDefaultItemStyles()
	styles.SelectedTitle = styles.SelectedTitle.
		Foreground(lipgloss.Color("#7D56F4")).
		BorderForeground(lipgloss.Color("#7D56F4"))

	return EditorDelegate{styles: styles}
}

func (d EditorDelegate) Height() int  { return 2 }
func (d EditorDelegate) Spacing() int { return 0 }
func (d EditorDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d EditorDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(interface {
		Title() string
		Description() string
	})
	if !ok {
		return
	}

	title := i.Title()
	desc := i.Description()

	if index == m.Index() {
		fmt.Fprintf(w, "%s\n%s",
			d.styles.SelectedTitle.Render("➜ "+title),
			d.styles.SelectedDesc.Render("    "+desc))
	} else {
		fmt.Fprintf(w, "%s\n%s",
			d.styles.NormalTitle.Render("  "+title),
			d.styles.NormalDesc.Render("    "+desc))
	}
}
