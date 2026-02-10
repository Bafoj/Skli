package skills

import (
	"fmt"
	"io"

	"skli/internal/gitrepo"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// skillItem implementa list.Item para un skill
type skillItem struct {
	skill *shared.Skill // Referencia al skill original para mantener el estado de selección
}

func (i skillItem) Title() string       { return i.skill.Info.Name }
func (i skillItem) Description() string { return i.skill.Info.Description }
func (i skillItem) FilterValue() string { return i.skill.Info.Name }

// skillDelegate es un delegate personalizado para mostrar checkboxes en skills
type skillDelegate struct {
	styles list.DefaultItemStyles
}

func newSkillDelegate() skillDelegate {
	styles := list.NewDefaultItemStyles()
	styles.SelectedTitle = styles.SelectedTitle.
		Foreground(lipgloss.Color("#7D56F4")).
		BorderForeground(lipgloss.Color("#7D56F4"))

	return skillDelegate{styles: styles}
}

func (d skillDelegate) Height() int  { return 2 }
func (d skillDelegate) Spacing() int { return 0 }
func (d skillDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == " " {
			if item, ok := m.SelectedItem().(skillItem); ok {
				item.skill.Selected = !item.skill.Selected
			}
		}
	}
	return nil
}

func (d skillDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(skillItem)
	if !ok {
		return
	}

	checked := "[ ]"
	if i.skill.Selected {
		checked = "[x]"
	}

	title := fmt.Sprintf("%s %s", checked, i.skill.Info.Name)
	var desc string
	if i.skill.Info.Description != "" {
		desc = i.skill.Info.Description
		if len(desc) > 80 {
			desc = desc[:77] + "..."
		}
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

// SkillsScreen es el modelo para la pantalla de selección de skills
type SkillsScreen struct {
	List            list.Model
	Skills          []shared.Skill
	TempDir         string
	RemoteURL       string
	SkillsRoot      string
	CommitHash      string
	ConfigLocalPath string
}

// NewSkillsScreen crea una nueva pantalla de selección de skills
func NewSkillsScreen(infos []gitrepo.SkillInfo, tempDir, remoteURL, skillsRoot, commitHash, configLocalPath string) SkillsScreen {
	skills := make([]shared.Skill, len(infos))
	items := make([]list.Item, len(infos))
	for i, info := range infos {
		skills[i] = shared.Skill{Info: info}
		items[i] = skillItem{skill: &skills[i]}
	}

	delegate := newSkillDelegate()
	l := list.New(items, delegate, 60, 20)
	l.Title = "Selecciona los skills a instalar"
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("skill", "skills")
	l.Styles.Title = shared.TitleStyle

	return SkillsScreen{
		List:            l,
		Skills:          skills,
		TempDir:         tempDir,
		RemoteURL:       remoteURL,
		SkillsRoot:      skillsRoot,
		CommitHash:      commitHash,
		ConfigLocalPath: configLocalPath,
	}
}

func (s SkillsScreen) Init() tea.Cmd {
	return nil
}

func (s SkillsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.List.SetSize(msg.Width, msg.Height-4)
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			var selected []gitrepo.SkillInfo
			for _, sk := range s.Skills {
				if sk.Selected {
					selected = append(selected, sk.Info)
				}
			}
			if len(selected) > 0 {
				if s.ConfigLocalPath == "" {
					return s, func() tea.Msg {
						return shared.NavigateToEditorMsg{
							Skills:     s.Skills,
							TempDir:    s.TempDir,
							RemoteURL:  s.RemoteURL,
							SkillsRoot: s.SkillsRoot,
							CommitHash: s.CommitHash,
						}
					}
				}
				return s, func() tea.Msg {
					return shared.NavigateToProgressMsg{
						TempDir:         s.TempDir,
						RemoteURL:       s.RemoteURL,
						SkillsRoot:      s.SkillsRoot,
						ConfigLocalPath: s.ConfigLocalPath,
						CommitHash:      s.CommitHash,
						Selected:        selected,
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	s.List, cmd = s.List.Update(msg)
	return s, cmd
}

func (s SkillsScreen) View() string {
	originalTitle := s.List.Title
	s.List.Title = fmt.Sprintf("%s (Pág. %d/%d)", originalTitle, s.List.Paginator.Page+1, s.List.Paginator.TotalPages)
	view := s.List.View()
	s.List.Title = originalTitle
	return view
}
