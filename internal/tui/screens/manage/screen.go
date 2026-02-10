package manage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"skli/internal/db"
	"skli/internal/gitrepo"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	// Sub-views for states
	"skli/internal/tui/screens/manage/confirm"
	"skli/internal/tui/screens/manage/input_remote"
	"skli/internal/tui/screens/manage/list_view"
	"skli/internal/tui/screens/manage/selecting_remote"
	"skli/internal/tui/screens/manage/uploading"
)

type State int

const (
	StateList State = iota
	StateConfirm
	StateSelectingRemote
	StateInputRemote
	StateUploading
)

// ManageScreen es el modelo para gestionar skills instalados
type ManageScreen struct {
	State         State
	List          list.Model
	Skills        []db.InstalledSkill
	ToDelete      *db.InstalledSkill
	Msg           string // Mensaje de estado/error
	ConfirmCursor int    // 0 para Sí, 1 para No
	RemoteList    list.Model
	RemoteInput   textinput.Model
	SelectedSkill *db.InstalledSkill // Skill seleccionado para subir
	ConfigRemotes []string           // Remotes configurados
}

// NewManageScreen crea una nueva pantalla de gestión
func NewManageScreen(remotes []string) (ManageScreen, tea.Cmd) {
	lock, _ := db.LoadLockFile()
	skills := lock.Skills

	// Escanear locales
	localSkills := scanLocalSkills(skills)
	skills = append(skills, localSkills...)

	items := make([]list.Item, len(skills))
	for i, s := range skills {
		items[i] = InstalledSkillItem{Skill: s}
	}

	delegate := NewManageDelegate()
	l := list.New(items, delegate, 60, 20)
	l.Title = "Gestionar Skills Instalados"
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("skill", "skills")
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "upload PR")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		}
	}
	l.Styles.Title = shared.TitleStyle

	ti := textinput.New()
	ti.Placeholder = "https://github.com/usuario/repo.git"
	ti.CharLimit = 156
	ti.Width = 50

	return ManageScreen{
		State:         StateList,
		List:          l,
		Skills:        skills,
		ConfirmCursor: 1, // Por defecto en No por seguridad
		RemoteInput:   ti,
		ConfigRemotes: remotes,
	}, nil
}

func (s ManageScreen) Init() tea.Cmd {
	return nil
}

func (s ManageScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch s.State {
	case StateList:
		return s.updateList(msg)
	case StateConfirm:
		return s.updateConfirm(msg)
	case StateSelectingRemote:
		return s.updateSelectingRemote(msg)
	case StateInputRemote:
		return s.updateInputRemote(msg)
	case StateUploading:
		return s.updateUploading(msg)
	}
	return s, nil
}

func (s ManageScreen) View() string {
	switch s.State {
	case StateList:
		return list_view.View(s.List, s.Skills, s.Msg)
	case StateConfirm:
		return confirm.View(s.ToDelete, s.ConfirmCursor)
	case StateSelectingRemote:
		return selecting_remote.View(s.RemoteList)
	case StateInputRemote:
		return input_remote.View(s.RemoteInput)
	case StateUploading:
		return uploading.View(s.Msg)
	}
	return ""
}

// Internal updates and helpers...

func (s ManageScreen) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.List.SetSize(msg.Width, msg.Height-4)
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "d", "backspace":
			if item, ok := s.List.SelectedItem().(InstalledSkillItem); ok {
				s.ToDelete = &item.Skill
				s.State = StateConfirm
				s.ConfirmCursor = 1 // Reset a No
				return s, nil
			}
		case "esc", "q":
			return s, func() tea.Msg { return shared.QuitMsg{} }
		case "u":
			if item, ok := s.List.SelectedItem().(InstalledSkillItem); ok {
				s.SelectedSkill = &item.Skill
				// Preparar lista de remotes
				var items []list.Item
				seen := make(map[string]bool)

				// 1. Añadir remote del skill si existe (ORIGIN)
				if item.Skill.RemoteRepo != "" {
					items = append(items, remoteItem{
						url:         item.Skill.RemoteRepo,
						displayName: fmt.Sprintf("%s (Origin)", item.Skill.RemoteRepo),
					})
					seen[item.Skill.RemoteRepo] = true
				}

				// 2. Añadir remotes configurados
				for _, r := range s.ConfigRemotes {
					if !seen[r] {
						items = append(items, remoteItem{url: r})
						seen[r] = true
					}
				}

				// 3. Añadir opcion custom
				items = append(items, customURLItem{})

				delegate := newRemoteDelegate()
				l := list.New(items, delegate, 60, 14)
				l.Title = "Selecciona repositorio destino"
				l.SetShowStatusBar(false)
				l.SetFilteringEnabled(false)
				l.Styles.Title = shared.TitleStyle
				s.RemoteList = l

				s.State = StateSelectingRemote
				return s, nil
			}
		}
	}

	var cmd tea.Cmd
	s.List, cmd = s.List.Update(msg)
	return s, cmd
}

func (s ManageScreen) updateSelectingRemote(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.State = StateList
			return s, nil
		case "enter":
			selected := s.RemoteList.SelectedItem()
			if selected == nil {
				return s, nil
			}
			switch item := selected.(type) {
			case remoteItem:
				// Iniciar upload con esta URL
				s.State = StateUploading
				s.Msg = fmt.Sprintf("Iniciando subida de '%s' a %s...", s.SelectedSkill.Name, item.url)
				return s, uploadSkillCmd(*s.SelectedSkill, item.url)
			case customURLItem:
				s.State = StateInputRemote
				s.RemoteInput.Focus()
				s.RemoteInput.SetValue("")
				return s, textinput.Blink
			}
		}
	}
	var cmd tea.Cmd
	s.RemoteList, cmd = s.RemoteList.Update(msg)
	return s, cmd
}

func (s ManageScreen) updateInputRemote(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.State = StateSelectingRemote
			return s, nil
		case "enter":
			url := s.RemoteInput.Value()
			if url != "" {
				s.State = StateUploading
				s.Msg = fmt.Sprintf("Iniciando subida de '%s' a %s...", s.SelectedSkill.Name, url)
				return s, uploadSkillCmd(*s.SelectedSkill, url)
			}
		}
	}
	var cmd tea.Cmd
	s.RemoteInput, cmd = s.RemoteInput.Update(msg)
	return s, cmd
}

func (s ManageScreen) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "right", "l", "tab":
			if s.ConfirmCursor == 0 {
				s.ConfirmCursor = 1
			} else {
				s.ConfirmCursor = 0
			}
		case "y", "Y":
			s.ConfirmCursor = 0
			if s.ToDelete != nil {
				return s, shared.DeleteSkillCmd(*s.ToDelete)
			}
		case "n", "N", "esc":
			s.State = StateList
			s.ToDelete = nil
			return s, nil
		case "enter":
			if s.ConfirmCursor == 0 {
				if s.ToDelete != nil {
					return s, shared.DeleteSkillCmd(*s.ToDelete)
				}
			} else {
				s.State = StateList
				s.ToDelete = nil
				return s, nil
			}
		}
	}
	return s, nil
}

func (s ManageScreen) updateUploading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case uploadSkillMsg:
		if msg.Err != nil {
			s.Msg = fmt.Sprintf("Error: %v", msg.Err)
			return s, nil
		}
		s.Msg = fmt.Sprintf("PR Creado: %s", msg.PRURL)
		return s, nil
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "enter" {
			s.State = StateList
			s.Msg = ""
			return s, nil
		}
	}
	return s, nil
}

// scanLocalSkills busca skills en la carpeta local que no estén en el lockfile
func scanLocalSkills(existingSkills []db.InstalledSkill) []db.InstalledSkill {
	localSkillsPath := "skills"
	var newSkills []db.InstalledSkill

	existingMap := make(map[string]bool)
	for _, s := range existingSkills {
		existingMap[s.Path] = true
	}

	filepath.Walk(localSkillsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == "SKILL.md" {
			dir := filepath.Dir(path)
			relPath, _ := filepath.Rel(".", dir)

			if !existingMap[relPath] {
				skillName := filepath.Base(dir)
				newSkills = append(newSkills, db.InstalledSkill{
					Name:        skillName + " (Local)",
					Path:        relPath,
					Description: "Local skill (unmanaged)",
					RemoteRepo:  "",
				})
			}
		}
		return nil
	})
	return newSkills
}

// Item types and delegates (keeping them here for now or in internal types if used by views)

type InstalledSkillItem struct {
	Skill db.InstalledSkill
}

func (i InstalledSkillItem) Title() string       { return i.Skill.Name }
func (i InstalledSkillItem) Description() string { return i.Skill.Description }
func (i InstalledSkillItem) FilterValue() string { return i.Skill.Name }

type manageDelegate struct {
	styles list.DefaultItemStyles
}

func NewManageDelegate() manageDelegate {
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
	i, ok := item.(InstalledSkillItem)
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

// Remote selection types (internal to manage)

type remoteItem struct {
	url         string
	displayName string
}

func (i remoteItem) Title() string {
	if i.displayName != "" {
		return i.displayName
	}
	return i.url
}
func (i remoteItem) Description() string { return "" }
func (i remoteItem) FilterValue() string { return i.url }

type customURLItem struct{}

func (i customURLItem) Title() string       { return "✏️  Custom URL..." }
func (i customURLItem) Description() string { return "Introduce una URL manualmente" }
func (i customURLItem) FilterValue() string { return "custom url" }

type remoteDelegate struct {
	styles list.DefaultItemStyles
}

func newRemoteDelegate() remoteDelegate {
	styles := list.NewDefaultItemStyles()
	styles.SelectedTitle = styles.SelectedTitle.
		Foreground(lipgloss.Color("#7D56F4")).
		BorderForeground(lipgloss.Color("#7D56F4"))

	return remoteDelegate{styles: styles}
}

func (d remoteDelegate) Height() int                             { return 1 }
func (d remoteDelegate) Spacing() int                            { return 0 }
func (d remoteDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d remoteDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(list.DefaultItem)
	if !ok {
		return
	}
	title := i.Title()
	if index == m.Index() {
		fmt.Fprint(w, d.styles.SelectedTitle.Render("➜ "+title))
	} else {
		fmt.Fprint(w, d.styles.NormalTitle.Render("  "+title))
	}
}

// Upload commands (internal to manage)

type uploadSkillMsg struct {
	PRURL string
	Err   error
}

func uploadSkillCmd(skill db.InstalledSkill, targetRemoteURL string) tea.Cmd {
	return func() tea.Msg {
		if !gitrepo.CheckGhInstalled() {
			return uploadSkillMsg{Err: fmt.Errorf("gh CLI no está instalada")}
		}
		tempDir, err := gitrepo.CloneForPush(targetRemoteURL)
		if err != nil {
			return uploadSkillMsg{Err: err}
		}
		defer os.RemoveAll(tempDir)
		branchName, err := gitrepo.PrepareSkillBranch(tempDir, skill.Name)
		if err != nil {
			return uploadSkillMsg{Err: err}
		}
		repoSkillPath := skill.RemotePath
		foundPath, err := gitrepo.FindSkillInRepo(tempDir, skill.Name)
		if err == nil {
			repoSkillPath = foundPath
		} else {
			skillFolderName := filepath.Base(skill.Path)
			repoSkillPath = filepath.Join("skills", skillFolderName)
		}
		err = gitrepo.CopySkillFiles(tempDir, skill.Path, repoSkillPath)
		if err != nil {
			return uploadSkillMsg{Err: err}
		}
		prURL, err := gitrepo.PushAndCreatePR(tempDir, branchName, skill.Name, skill.Description)
		if err != nil {
			return uploadSkillMsg{Err: err}
		}
		return uploadSkillMsg{PRURL: prURL}
	}
}
