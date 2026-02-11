package commands

import (
	"fmt"
	"os"

	"skli/internal/db"
	"skli/internal/gitrepo"

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

func UploadSkillsCmd(skills []db.InstalledSkill, targetRemoteURL string) tea.Cmd {
	return func() tea.Msg {
		results := make([]UploadResult, 0, len(skills))
		for _, sk := range skills {
			prURL, err := gitrepo.UploadSkill(sk, targetRemoteURL)
			results = append(results, UploadResult{SkillName: sk.Name, PRURL: prURL, Err: err})
		}
		return UploadSkillsMsg{Results: results}
	}
}

func DeleteSkillsCmd(skills []db.InstalledSkill) tea.Cmd {
	return func() tea.Msg {
		if len(skills) == 0 {
			return DeleteSkillsMsg{Err: fmt.Errorf("no hay skills seleccionados")}
		}
		deleted := make([]string, 0, len(skills))
		for _, sk := range skills {
			if sk.Path == "" || sk.Path == "." || sk.Path == "/" {
				return DeleteSkillsMsg{Err: fmt.Errorf("ruta insegura para eliminar: %s", sk.Path)}
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
