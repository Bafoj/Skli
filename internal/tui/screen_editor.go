package tui

import (
	"fmt"
	"io"

	"skli/internal/gitrepo"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// editorItem implementa list.DefaultItem para un editor
type editorItem struct {
	editor Editor
}

func (i editorItem) Title() string { return i.editor.Name }
func (i editorItem) Description() string {
	if i.editor.Path != "" {
		return "(" + i.editor.Path + ")"
	}
	return "Ruta personalizada"
}
func (i editorItem) FilterValue() string { return i.editor.Name }

// editorDelegate es un delegate para editores
type editorDelegate struct {
	styles list.DefaultItemStyles
}

func newEditorDelegate() editorDelegate {
	styles := list.NewDefaultItemStyles()
	styles.SelectedTitle = styles.SelectedTitle.
		Foreground(lipgloss.Color("#7D56F4")).
		BorderForeground(lipgloss.Color("#7D56F4"))
	styles.SelectedDesc = styles.SelectedDesc.
		Foreground(lipgloss.Color("#7D56F4")).
		BorderForeground(lipgloss.Color("#7D56F4"))

	return editorDelegate{styles: styles}
}

func (d editorDelegate) Height() int  { return 2 }
func (d editorDelegate) Spacing() int { return 0 }
func (d editorDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d editorDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(editorItem)
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

// EditorScreen es el modelo para la pantalla de selección de editor
type EditorScreen struct {
	list       list.Model
	skills     []Skill
	tempDir    string
	remoteURL  string
	skillsRoot string
	commitHash string
	configMode bool
	remotes    []string
}

// NewEditorScreen crea una nueva pantalla de selección de editor
func NewEditorScreen(skills []Skill, tempDir, remoteURL, skillsRoot, commitHash string, configMode bool, remotes []string) EditorScreen {
	items := make([]list.Item, len(Editors))
	for i, ed := range Editors {
		items[i] = editorItem{editor: ed}
	}

	delegate := newEditorDelegate()
	l := list.New(items, delegate, 60, 15)
	l.Title = "Selecciona tu editor"
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("editor", "editores")
	l.Styles.Title = TitleStyle

	return EditorScreen{
		list:       l,
		skills:     skills,
		tempDir:    tempDir,
		remoteURL:  remoteURL,
		skillsRoot: skillsRoot,
		commitHash: commitHash,
		configMode: configMode,
		remotes:    remotes,
	}
}

// NewEditorScreenForConfig crea una pantalla de editor desde config
func NewEditorScreenForConfig(currentPath string, remotes []string) EditorScreen {
	items := make([]list.Item, len(Editors))
	cursor := 0
	for i, ed := range Editors {
		items[i] = editorItem{editor: ed}
		if ed.Path == currentPath && ed.Path != "" {
			cursor = i
		}
	}
	// Si no coincide con ninguno conocido pero hay path, marcar como Custom (último)
	if cursor == 0 && currentPath != "" && currentPath != Editors[0].Path {
		cursor = len(Editors) - 1
	}

	delegate := newEditorDelegate()
	l := list.New(items, delegate, 60, 15)
	l.Title = "Configura tu editor"
	l.Select(cursor)
	l.SetShowStatusBar(true)
	l.Styles.Title = TitleStyle

	return EditorScreen{
		list:       l,
		configMode: true,
		remotes:    remotes,
	}
}

func (s EditorScreen) Init() tea.Cmd {
	return nil
}

func (s EditorScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.list.SetSize(msg.Width, msg.Height-4)
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			return s, func() tea.Msg { return NavigateToConfigMsg{} }
		case "enter":
			selected := s.list.SelectedItem()
			if selected == nil {
				return s, nil
			}
			item := selected.(editorItem)
			destPath := item.editor.Path
			if item.editor.Name == "Custom" {
				destPath = "skills"
			}

			if s.configMode {
				return s, tea.Batch(
					SaveConfigCmd(destPath, s.remotes),
					func() tea.Msg { return NavigateToConfigMsg{} },
				)
			}

			var selectedSkills []gitrepo.SkillInfo
			for _, sk := range s.skills {
				if sk.Selected {
					selectedSkills = append(selectedSkills, sk.Info)
				}
			}

			return s, func() tea.Msg {
				return NavigateToProgressMsg{
					TempDir:         s.tempDir,
					RemoteURL:       s.remoteURL,
					SkillsRoot:      s.skillsRoot,
					ConfigLocalPath: destPath,
					CommitHash:      s.commitHash,
					Selected:        selectedSkills,
				}
			}
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

func (s EditorScreen) View() string {
	originalTitle := s.list.Title
	s.list.Title = fmt.Sprintf("%s (Pág. %d/%d)", originalTitle, s.list.Paginator.Page+1, s.list.Paginator.TotalPages)
	view := s.list.View()
	s.list.Title = originalTitle
	return view
}
