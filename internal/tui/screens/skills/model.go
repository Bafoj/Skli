package skills

import (
	"skli/internal/gitrepo"
	"skli/internal/tui/screens/skills/delegates"
	"skli/internal/tui/shared"

	"github.com/charmbracelet/bubbles/list"
)

// skillItem implementa list.Item para un skill
type skillItem struct {
	skill *shared.Skill // Referencia al skill original para mantener el estado de selección
}

func (i skillItem) Title() string       { return i.skill.Info.Name }
func (i skillItem) Description() string { return i.skill.Info.Description }
func (i skillItem) FilterValue() string { return i.skill.Info.Name }
func (i skillItem) Toggle() {
	i.skill.Selected = !i.skill.Selected
}
func (i skillItem) IsSelected() bool { return i.skill.Selected }

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

	delegate := delegates.NewSkillDelegate()
	l := list.New(items, delegate, 60, 20)
	l.Title = "Select skills to install"
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
