package editor

import (
	"skli/internal/tui/screens/editor/delegates"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
)

type State int

const (
	StateSelecting State = iota
	StateInputCustom
)

// editorItem implementa list.DefaultItem para un editor
type editorItem struct {
	editor shared.Editor
}

func (i editorItem) Title() string { return i.editor.Name }
func (i editorItem) Description() string {
	if i.editor.Path != "" {
		return "(" + i.editor.Path + ")"
	}
	return "Custom path"
}
func (i editorItem) FilterValue() string { return i.editor.Name }

// EditorScreen es el modelo para la pantalla de selección de editor
type EditorScreen struct {
	State      State
	List       list.Model
	TextInput  textinput.Model
	Skills     []shared.Skill
	TempDir    string
	RemoteURL  string
	SkillsRoot string
	CommitHash string
	ConfigMode bool
	Remotes    []string
}

// NewEditorScreen crea una nueva pantalla de selección de editor
func NewEditorScreen(skills []shared.Skill, tempDir, remoteURL, skillsRoot, commitHash string, configMode bool, remotes []string) EditorScreen {
	items := make([]list.Item, len(shared.Editors))
	for i, ed := range shared.Editors {
		items[i] = editorItem{editor: ed}
	}

	delegate := delegates.NewEditorDelegate()
	l := list.New(items, delegate, 60, 15)
	l.Title = "Select your editor"
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("editor", "editors")
	l.SetFilteringEnabled(true)
	l.SetShowFilter(true)
	l.FilterInput.Prompt = "Search: "
	l.Styles.Title = shared.TitleStyle

	ti := textinput.New()
	ti.Placeholder = "/path/to/custom/folder"
	ti.CharLimit = 256
	ti.Width = 50

	return EditorScreen{
		State:      StateSelecting,
		List:       l,
		TextInput:  ti,
		Skills:     skills,
		TempDir:    tempDir,
		RemoteURL:  remoteURL,
		SkillsRoot: skillsRoot,
		CommitHash: commitHash,
		ConfigMode: configMode,
		Remotes:    remotes,
	}
}

// NewEditorScreenForConfig crea una pantalla de editor desde config
func NewEditorScreenForConfig(currentPath string, remotes []string) EditorScreen {
	items := make([]list.Item, len(shared.Editors))
	cursor := 0
	for i, ed := range shared.Editors {
		items[i] = editorItem{editor: ed}
		if ed.Path == currentPath && ed.Path != "" {
			cursor = i
		}
	}
	if cursor == 0 && currentPath != "" {
		// Check if it matches any known editor
		found := false
		for _, ed := range shared.Editors {
			if ed.Path == currentPath {
				found = true
				break
			}
		}
		if !found {
			cursor = len(shared.Editors) - 1 // Custom
		}
	}

	delegate := delegates.NewEditorDelegate()
	l := list.New(items, delegate, 60, 15)
	l.Title = "Configure your editor"
	l.Select(cursor)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(true)
	l.FilterInput.Prompt = "Search: "
	l.Styles.Title = shared.TitleStyle

	ti := textinput.New()
	ti.Placeholder = "/path/to/custom/folder"
	if cursor == len(shared.Editors)-1 && currentPath != "" {
		ti.SetValue(currentPath)
	}
	ti.CharLimit = 256
	ti.Width = 50

	return EditorScreen{
		State:      StateSelecting,
		List:       l,
		TextInput:  ti,
		ConfigMode: true,
		Remotes:    remotes,
	}
}
