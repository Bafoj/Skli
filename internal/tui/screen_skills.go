package tui

import (
	"fmt"
	"io"

	"skli/internal/gitrepo"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// skillItem implementa list.Item para un skill
type skillItem struct {
	skill *Skill // Referencia al skill original para mantener el estado de selección
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
	styles.SelectedDesc = styles.SelectedDesc.
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
	list            list.Model
	skills          []Skill
	tempDir         string
	remoteURL       string
	skillsRoot      string
	commitHash      string
	configLocalPath string
}

// NewSkillsScreen crea una nueva pantalla de selección de skills
func NewSkillsScreen(infos []gitrepo.SkillInfo, tempDir, remoteURL, skillsRoot, commitHash, configLocalPath string) SkillsScreen {
	skills := make([]Skill, len(infos))
	items := make([]list.Item, len(infos))
	for i, info := range infos {
		skills[i] = Skill{Info: info}
		items[i] = skillItem{skill: &skills[i]}
	}

	delegate := newSkillDelegate()
	l := list.New(items, delegate, 60, 20)
	l.Title = "Selecciona los skills a instalar"
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("skill", "skills")
	l.Styles.Title = TitleStyle

	return SkillsScreen{
		list:            l,
		skills:          skills,
		tempDir:         tempDir,
		remoteURL:       remoteURL,
		skillsRoot:      skillsRoot,
		commitHash:      commitHash,
		configLocalPath: configLocalPath,
	}
}

func (s SkillsScreen) Init() tea.Cmd {
	return nil
}

func (s SkillsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.list.SetSize(msg.Width, msg.Height-4)
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			var selected []gitrepo.SkillInfo
			for _, sk := range s.skills {
				if sk.Selected {
					selected = append(selected, sk.Info)
				}
			}
			if len(selected) > 0 {
				if s.configLocalPath == "" {
					return s, func() tea.Msg {
						return NavigateToEditorMsg{
							Skills:     s.skills,
							TempDir:    s.tempDir,
							RemoteURL:  s.remoteURL,
							SkillsRoot: s.skillsRoot,
							CommitHash: s.commitHash,
						}
					}
				}
				return s, func() tea.Msg {
					return NavigateToProgressMsg{
						TempDir:         s.tempDir,
						RemoteURL:       s.remoteURL,
						SkillsRoot:      s.skillsRoot,
						ConfigLocalPath: s.configLocalPath,
						CommitHash:      s.commitHash,
						Selected:        selected,
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

func (s SkillsScreen) View() string {
	originalTitle := s.list.Title
	s.list.Title = fmt.Sprintf("%s (Pág. %d/%d)", originalTitle, s.list.Paginator.Page+1, s.list.Paginator.TotalPages)
	view := s.list.View()
	s.list.Title = originalTitle
	return view
}
