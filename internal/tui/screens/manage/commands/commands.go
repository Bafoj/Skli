package commands

import (
	"fmt"
	"os"

	"skli/internal/db"
	"skli/internal/gitrepo"
	skillsvc "skli/internal/skills"

	tea "github.com/charmbracelet/bubbletea"
)

type UploadResult struct {
	SkillName string
	PRURL     string
	Err       error
}

type UploadSkillsMsg struct {
	Results []UploadResult
}

type DeleteSkillsMsg struct {
	Deleted []string
	Err     error
}

func UploadSkillsCmd(selectedSkills []db.InstalledSkill, targetRemoteURL string) tea.Cmd {
	return func() tea.Msg {
		results := make([]UploadResult, 0, len(selectedSkills))
		for _, sk := range selectedSkills {
			prURL, err := gitrepo.UploadSkill(sk, targetRemoteURL)
			results = append(results, UploadResult{SkillName: sk.Name, PRURL: prURL, Err: err})
		}
		return UploadSkillsMsg{Results: results}
	}
}

func DeleteSkillsCmd(selectedSkills []db.InstalledSkill) tea.Cmd {
	return func() tea.Msg {
		if len(selectedSkills) == 0 {
			return DeleteSkillsMsg{Err: fmt.Errorf("no hay skills seleccionados")}
		}
		deleted := make([]string, 0, len(selectedSkills))
		for _, sk := range selectedSkills {
			if err := skillsvc.IsSafeDeletePath(sk.Path, skillsvc.DefaultRoot); err != nil {
				return DeleteSkillsMsg{Err: err}
			}
			if err := os.RemoveAll(sk.Path); err != nil {
				return DeleteSkillsMsg{Err: err}
			}
			if err := db.DeleteInstalledSkill(sk.Path); err != nil {
				return DeleteSkillsMsg{Err: err}
			}
			deleted = append(deleted, sk.Name)
		}
		return DeleteSkillsMsg{Deleted: deleted}
	}
}
