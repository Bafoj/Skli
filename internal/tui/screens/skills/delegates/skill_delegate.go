package delegates

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SkillDelegate struct {
	styles list.DefaultItemStyles
}

func NewSkillDelegate() SkillDelegate {
	styles := list.NewDefaultItemStyles()
	styles.SelectedTitle = styles.SelectedTitle.
		Foreground(lipgloss.Color("#7D56F4")).
		BorderForeground(lipgloss.Color("#7D56F4"))

	return SkillDelegate{styles: styles}
}

func (d SkillDelegate) Height() int  { return 2 }
func (d SkillDelegate) Spacing() int { return 0 }
func (d SkillDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == " " {
			if item, ok := m.SelectedItem().(interface{ Toggle() }); ok {
				item.Toggle()
			}
		}
	}
	return nil
}

func (d SkillDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(interface {
		Title() string
		Description() string
		IsSelected() bool
	})
	if !ok {
		return
	}

	checked := "[ ]"
	if i.IsSelected() {
		checked = "[x]"
	}

	title := fmt.Sprintf("%s %s", checked, i.Title())
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
