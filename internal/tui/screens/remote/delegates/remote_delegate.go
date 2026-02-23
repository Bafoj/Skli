package delegates

import (
	"fmt"
	"io"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type RemoteDelegate struct {
	styles list.DefaultItemStyles
}

func NewRemoteDelegate() RemoteDelegate {
	styles := shared.NewListItemStyles()
	return RemoteDelegate{styles: styles}
}

func (d RemoteDelegate) Height() int  { return 1 }
func (d RemoteDelegate) Spacing() int { return 0 }
func (d RemoteDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d RemoteDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(interface{ Title() string })
	if !ok {
		return
	}
	title := i.Title()
	if index == m.Index() {
		fmt.Fprint(w, d.styles.SelectedTitle.Render(shared.SelectorDot(true)+" "+title))
	} else {
		fmt.Fprint(w, d.styles.NormalTitle.Render(shared.SelectorDot(false)+" "+title))
	}
}
