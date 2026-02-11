package delegates

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ManageDelegate struct {
	styles       list.DefaultItemStyles
	showCheckbox bool
}

func NewManageDelegate(showCheckbox bool) ManageDelegate {
	styles := list.NewDefaultItemStyles()
	styles.SelectedTitle = styles.SelectedTitle.
		Foreground(lipgloss.Color("#FF0000")).
		BorderForeground(lipgloss.Color("#FF0000"))
	styles.SelectedDesc = styles.SelectedDesc.
		Foreground(lipgloss.Color("#FF0000")).
		BorderForeground(lipgloss.Color("#FF0000"))

	return ManageDelegate{styles: styles, showCheckbox: showCheckbox}
}

func (d ManageDelegate) Height() int  { return 2 }
func (d ManageDelegate) Spacing() int { return 0 }
func (d ManageDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d ManageDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(interface {
		Title() string
		Description() string
		IsSelected() bool
	})
	if !ok {
		return
	}

	title := i.Title()
	if d.showCheckbox {
		checked := "[ ]"
		if i.IsSelected() {
			checked = "[x]"
		}
		title = fmt.Sprintf("%s %s", checked, title)
	}
	desc := i.Description()
	if len(desc) > 80 {
		desc = desc[:77] + "..."
	}

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

type RemoteDelegate struct {
	styles list.DefaultItemStyles
}

func NewRemoteDelegate() RemoteDelegate {
	styles := list.NewDefaultItemStyles()
	styles.SelectedTitle = styles.SelectedTitle.
		Foreground(lipgloss.Color("#7D56F4")).
		BorderForeground(lipgloss.Color("#7D56F4"))

	return RemoteDelegate{styles: styles}
}

func (d RemoteDelegate) Height() int  { return 1 }
func (d RemoteDelegate) Spacing() int { return 0 }
func (d RemoteDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d RemoteDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	titled, ok := item.(interface{ Title() string })
	if !ok {
		return
	}
	title := titled.Title()
	if index == m.Index() {
		fmt.Fprint(w, d.styles.SelectedTitle.Render("➜ "+title))
	} else {
		fmt.Fprint(w, d.styles.NormalTitle.Render("  "+title))
	}
}
