package editor

import (
	"skli/internal/tui/screens/editor/delegates"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/list"
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
	return "Ruta personalizada"
}
func (i editorItem) FilterValue() string { return i.editor.Name }

// EditorScreen es el modelo para la pantalla de selección de editor
type EditorScreen struct {
	List       list.Model
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
	l.Title = "Selecciona tu editor"
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("editor", "editores")
	l.Styles.Title = shared.TitleStyle

	return EditorScreen{
		List:       l,
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
	if cursor == 0 && currentPath != "" && currentPath != shared.Editors[0].Path {
		cursor = len(shared.Editors) - 1
	}

	delegate := delegates.NewEditorDelegate()
	l := list.New(items, delegate, 60, 15)
	l.Title = "Configura tu editor"
	l.Select(cursor)
	l.SetShowStatusBar(true)
	l.Styles.Title = shared.TitleStyle

	return EditorScreen{
		List:       l,
		ConfigMode: true,
		Remotes:    remotes,
	}
}
