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

// Item types and delegates

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
