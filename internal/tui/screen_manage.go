package tui

import (
	"fmt"
	"io"

	"skli/internal/db"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// installedSkillItem implementa list.Item para un skill instalado
type installedSkillItem struct {
	skill db.InstalledSkill
}

func (i installedSkillItem) Title() string       { return i.skill.Name }
func (i installedSkillItem) Description() string { return i.skill.Description }
func (i installedSkillItem) FilterValue() string { return i.skill.Name }

// manageDelegate es un delegate para el listado de gestión
type manageDelegate struct {
	styles list.DefaultItemStyles
}

func newManageDelegate() manageDelegate {
	styles := list.NewDefaultItemStyles()
	styles.SelectedTitle = styles.SelectedTitle.
		Foreground(lipgloss.Color("#FF0000")).
		BorderForeground(lipgloss.Color("#FF0000"))
	styles.SelectedDesc = styles.SelectedDesc.
		Foreground(lipgloss.Color("#FF0000")).
		BorderForeground(lipgloss.Color("#FF0000"))

	return manageDelegate{styles: styles}
}

func (d manageDelegate) Height() int  { return 2 }
func (d manageDelegate) Spacing() int { return 0 }
func (d manageDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d manageDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(installedSkillItem)
	if !ok {
		return
	}

	title := i.Title()
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

type ManageScreenState int

const (
	ManageStateList ManageScreenState = iota
	ManageStateConfirm
)

// ManageScreen es el modelo para gestionar skills instalados
type ManageScreen struct {
	state         ManageScreenState
	list          list.Model
	skills        []db.InstalledSkill
	toDelete      *db.InstalledSkill
	confirmCursor int // 0 para Sí, 1 para No
}

// NewManageScreen crea una nueva pantalla de gestión
func NewManageScreen() (ManageScreen, tea.Cmd) {
	lock, _ := db.LoadLockFile()
	skills := lock.Skills

	items := make([]list.Item, len(skills))
	for i, s := range skills {
		items[i] = installedSkillItem{skill: s}
	}

	delegate := newManageDelegate()
	l := list.New(items, delegate, 60, 20)
	l.Title = "Gestionar Skills Instalados"
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("skill", "skills")
	l.Styles.Title = TitleStyle

	return ManageScreen{
		state:         ManageStateList,
		list:          l,
		skills:        skills,
		confirmCursor: 1, // Por defecto en No por seguridad
	}, nil
}

func (s ManageScreen) Init() tea.Cmd {
	return nil
}

func (s ManageScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch s.state {
	case ManageStateList:
		return s.updateList(msg)
	case ManageStateConfirm:
		return s.updateConfirm(msg)
	}
	return s, nil
}

func (s ManageScreen) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.list.SetSize(msg.Width, msg.Height-4)
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "d", "backspace":
			if item, ok := s.list.SelectedItem().(installedSkillItem); ok {
				s.toDelete = &item.skill
				s.state = ManageStateConfirm
				s.confirmCursor = 1 // Reset a No
				return s, nil
			}
		case "esc", "q":
			return s, func() tea.Msg { return QuitMsg{} }
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

func (s ManageScreen) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "right", "l", "tab":
			if s.confirmCursor == 0 {
				s.confirmCursor = 1
			} else {
				s.confirmCursor = 0
			}
		case "y", "Y":
			s.confirmCursor = 0
			if s.toDelete != nil {
				return s, DeleteSkillCmd(*s.toDelete)
			}
		case "n", "N", "esc":
			s.state = ManageStateList
			s.toDelete = nil
			return s, nil
		case "enter":
			if s.confirmCursor == 0 {
				if s.toDelete != nil {
					return s, DeleteSkillCmd(*s.toDelete)
				}
			} else {
				s.state = ManageStateList
				s.toDelete = nil
				return s, nil
			}
		}
	}
	return s, nil
}

func (s ManageScreen) View() string {
	if s.state == ManageStateConfirm {
		var yes, no string
		if s.confirmCursor == 0 {
			yes = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true).Render("➜ [ Sí ]")
			no = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  [ No ]")
		} else {
			yes = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  [ Sí ]")
			no = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true).Render("➜ [ No ]")
		}

		return fmt.Sprintf(
			"\n  ¿Estás DE ACUERDO en eliminar el skill %s?\n\n  Ruta: %s\n\n  %s    %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true).Render(s.toDelete.Name),
			DimStyle.Render(s.toDelete.Path),
			yes,
			no,
		)
	}

	if len(s.skills) == 0 {
		return "\n  No hay skills instalados.\n\n" + HelpStyle.Render("  q: salir")
	}

	originalTitle := s.list.Title
	s.list.Title = fmt.Sprintf("%s (Pág. %d/%d)", originalTitle, s.list.Paginator.Page+1, s.list.Paginator.TotalPages)
	view := s.list.View()
	s.list.Title = originalTitle
	return view
}
