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
	StateScanning
	StateSelectingSkills
	StateDownloading
	StateDone
	StateError
)

// Skill representa una habilidad encontrada en el repositorio
type Skill struct {
	Info     gitrepo.SkillInfo
	Selected bool
}

// Model define el estado de nuestra aplicación
type Model struct {
	State        State
	TempDir      string
	RemoteURL    string
	TextInput    textinput.Model
	Spinner      spinner.Model
	Skills       []Skill
	Cursor       int
	ErrorMessage string
	WindowWidth  int
	WindowHeight int
	Quitting     bool
}

// InitialModel inicializa el estado por defecto
func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "https://github.com/usuario/repo.git"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return Model{
		State:     StateInputRemote,
		TextInput: ti,
		Spinner:   s,
	}
}

// Init es el primer método que se llama
func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.Spinner.Tick)
}
