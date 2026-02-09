package tui

import (
	"skli/internal/gitrepo"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type State int

const (
	StateInputRemote State = iota
	StateInputNewRemote
	StateSelectingRemote
	StateManageRemotes
	StateScanning
	StateSelectingSkills
	StateSelectingEditor
	StateConfigMenu
	StateDownloading
	StateDone
	StateError
)

type Editor struct {
	Name string
	Path string
}

var Editors = []Editor{
	{Name: "Windsurf", Path: ".windsurf/skills"},
	{Name: "Antigravity", Path: ".antigravity/skills"},
	{Name: "Cursor", Path: ".cursor/skills"},
	{Name: "VSCode", Path: ".vscode/skills"},
	{Name: "OpenCode", Path: ".opencode/skills"},
	{Name: "Custom", Path: ""},
}

// Skill representa una habilidad encontrada en el repositorio
type Skill struct {
	Info     gitrepo.SkillInfo
	Selected bool
}

// Model define el estado de nuestra aplicación
type Model struct {
	State           State
	TempDir         string
	RemoteURL       string
	SkillsRoot      string // Directorio base dentro del repo (ej: "skills")
	ConfigLocalPath string // Ruta local destino (del config o seleccionada)
	CommitHash      string
	TextInput       textinput.Model
	Spinner         spinner.Model
	Skills          []Skill
	Cursor          int
	ConfigCursor    int
	EditorCursor    int
	ErrorMessage    string
	WindowWidth     int
	WindowHeight    int
	Quitting        bool
	ConfigMode      bool
	Remotes         []string
	RemoteCursor    int
}

// InitialModel inicializa el estado por defecto
func InitialModel(initialURL, skillsRoot, configLocalPath string, configMode bool, remotes []string) Model {
	ti := textinput.New()
	ti.Placeholder = "https://github.com/usuario/repo.git"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	if initialURL != "" {
		ti.SetValue(initialURL)
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	state := StateInputRemote
	if configMode {
		state = StateConfigMenu
	} else if initialURL != "" {
		state = StateScanning
	} else if len(remotes) > 0 {
		state = StateSelectingRemote
	}

	// Buscar el cursor del editor actual
	edCursor := 0
	for i, ed := range Editors {
		if ed.Path == configLocalPath && ed.Path != "" {
			edCursor = i
			break
		}
	}
	// Si no coincide con ninguno conocido pero hay path, marcar como Custom (último)
	if edCursor == 0 && configLocalPath != "" && configLocalPath != Editors[0].Path {
		edCursor = len(Editors) - 1
	}

	return Model{
		State:           state,
		RemoteURL:       initialURL,
		SkillsRoot:      skillsRoot,
		ConfigLocalPath: configLocalPath,
		TextInput:       ti,
		Spinner:         s,
		ConfigMode:      configMode,
		EditorCursor:    edCursor,
		Remotes:         remotes,
	}
}

// Init es el primer método que se llama
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{textinput.Blink, m.Spinner.Tick}
	if m.State == StateScanning && m.RemoteURL != "" {
		cmds = append(cmds, scanRepoCmd(m.RemoteURL, m.SkillsRoot))
	}
	return tea.Batch(cmds...)
}
