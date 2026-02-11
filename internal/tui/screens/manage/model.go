package manage

import (
	"skli/internal/db"
	"skli/internal/skills"
	"skli/internal/tui/screens/manage/delegates"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type State int

const (
	StateList State = iota
	StateConfirm
	StateSelectingRemote
	StateInputRemote
	StateUploading
)

type Mode int

const (
	ModeNone   Mode = -1
	ModeManage Mode = iota
	ModeRemove
	ModeUpload
)

// ManageScreen es el modelo para gestionar skills instalados
type ManageScreen struct {
	State         State
	Mode          Mode
	List          list.Model
	Skills        []managedSkill
	ToDelete      *db.InstalledSkill
	Msg           string // Mensaje de estado/error
	ConfirmCursor int    // 0 para Si, 1 para No
	RemoteList    list.Model
	RemoteInput   textinput.Model
	SelectedSkill *db.InstalledSkill // Skill seleccionado para subir
	ConfigRemotes []string           // Remotes configurados
	TargetRemote  string
}

// NewManageScreen crea una nueva pantalla de gestion
func NewManageScreen(remotes []string, mode Mode) (ManageScreen, tea.Cmd) {
	lock, _ := db.LoadLockFile()
	localOnly, _ := skills.ScanLocalUnmanaged(lock.Skills, skills.DefaultRoot)

	var sourceSkills []db.InstalledSkill
	switch mode {
	case ModeUpload:
		sourceSkills = localOnly
	default:
		sourceSkills = append(lock.Skills, localOnly...)
	}

	skills := make([]managedSkill, len(sourceSkills))
	items := make([]list.Item, len(sourceSkills))
	for i, sk := range sourceSkills {
		skills[i] = managedSkill{Skill: sk}
		items[i] = InstalledSkillItem{Skill: &skills[i]}
	}

	showCheckbox := mode == ModeRemove || mode == ModeUpload
	delegate := delegates.NewManageDelegate(showCheckbox)
	l := list.New(items, delegate, 60, 20)
	l.Title = listTitleForMode(mode)
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("skill", "skills")
	l.AdditionalShortHelpKeys = func() []key.Binding {
		switch mode {
		case ModeManage:
			return []key.Binding{
				key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "upload PR")),
				key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
			}
		case ModeRemove:
			return []key.Binding{
				key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "marcar")),
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "eliminar")),
			}
		case ModeUpload:
			return []key.Binding{
				key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "marcar")),
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "subir")),
			}
		}
		return nil
	}
	l.Styles.Title = shared.TitleStyle

	ti := textinput.New()
	ti.Placeholder = "https://github.com/usuario/repo.git"
	ti.CharLimit = 156
	ti.Width = 50

	initialState := StateList
	var remoteList list.Model
	if mode == ModeUpload {
		remoteList = buildUploadRemoteList(remotes)
		initialState = StateSelectingRemote
		if len(remotes) == 0 {
			initialState = StateInputRemote
			ti.Focus()
		}
	}

	return ManageScreen{
		State:         initialState,
		Mode:          mode,
		List:          l,
		Skills:        skills,
		ConfirmCursor: 1,
		RemoteList:    remoteList,
		RemoteInput:   ti,
		ConfigRemotes: remotes,
	}, nil
}

func (s ManageScreen) Init() tea.Cmd {
	if s.State == StateInputRemote {
		return textinput.Blink
	}
	return nil
}

// Item types and delegates

type managedSkill struct {
	Skill    db.InstalledSkill
	Selected bool
}

type InstalledSkillItem struct {
	Skill *managedSkill
}

func (i InstalledSkillItem) Title() string       { return i.Skill.Skill.Name }
func (i InstalledSkillItem) Description() string { return i.Skill.Skill.Description }
func (i InstalledSkillItem) FilterValue() string { return i.Skill.Skill.Name }
func (i InstalledSkillItem) IsSelected() bool    { return i.Skill.Selected }

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

func buildUploadRemoteList(remotes []string) list.Model {
	items := make([]list.Item, 0, len(remotes)+1)
	for _, remote := range remotes {
		items = append(items, remoteItem{url: remote})
	}
	items = append(items, customURLItem{})

	delegate := delegates.NewRemoteDelegate()
	l := list.New(items, delegate, 60, 14)
	l.Title = "Paso 1/2: Selecciona repositorio destino"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = shared.TitleStyle
	return l
}

func listTitleForMode(mode Mode) string {
	switch mode {
	case ModeRemove:
		return "Eliminar skills locales"
	case ModeUpload:
		return "Paso 2/2: Skills locales no sincronizados"
	default:
		return "Gestionar Skills Instalados"
	}
}

func (s ManageScreen) selectedSkills() []db.InstalledSkill {
	out := make([]db.InstalledSkill, 0)
	for _, sk := range s.Skills {
		if sk.Selected {
			out = append(out, sk.Skill)
		}
	}
	return out
}

func (s ManageScreen) toggleSelectedCurrent() ManageScreen {
	item, ok := s.List.SelectedItem().(InstalledSkillItem)
	if !ok || item.Skill == nil {
		return s
	}
	item.Skill.Selected = !item.Skill.Selected
	return s
}
