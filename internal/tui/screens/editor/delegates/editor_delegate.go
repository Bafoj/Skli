package delegates

import (
	"fmt"
	"io"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type EditorDelegate struct {
	styles list.DefaultItemStyles
}

func NewEditorDelegate() EditorDelegate {
	styles := shared.NewListItemStyles()
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
			d.styles.SelectedTitle.Render(shared.SelectorDot(true)+" "+title),
			d.styles.SelectedDesc.Render("    "+desc))
	} else {
		fmt.Fprintf(w, "%s\n%s",
			d.styles.NormalTitle.Render(shared.SelectorDot(false)+" "+title),
			d.styles.NormalDesc.Render("    "+desc))
	}
}
