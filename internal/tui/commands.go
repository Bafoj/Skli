package tui

import (
	"os"
	"path/filepath"

	"skli/internal/config"
	"skli/internal/db"
	"skli/internal/gitrepo"

	tea "github.com/charmbracelet/bubbletea"
)

// ScanRepoCmd escanea un repositorio remoto
func ScanRepoCmd(url, defaultSkillsPath string) tea.Cmd {
	return func() tea.Msg {
		res, err := gitrepo.CloneAndScan(url, defaultSkillsPath)
		return ScanResultMsg{Result: res, RemoteURL: url, Err: err}
	}
}

// DownloadSkillsCmd descarga e instala skills seleccionadas
func DownloadSkillsCmd(tempDir, remoteURL, skillsPath, localPath, commitHash string, selected []gitrepo.SkillInfo) tea.Cmd {
	return func() tea.Msg {
		err := gitrepo.InstallSkills(tempDir, skillsPath, localPath, selected)
		if err != nil {
			os.RemoveAll(tempDir)
			return DownloadResultMsg{Err: err}
		}

		for _, skill := range selected {
			folderName := gitrepo.GetSkillFolderName(skill)
			localSkillPath := filepath.Join(localPath, folderName)
			instSkill := db.InstalledSkill{
				Name:        skill.Name,
				Description: skill.Description,
				Path:        localSkillPath,
				RemoteRepo:  remoteURL,
				RemoteRoot:  skillsPath,
				RemotePath:  skill.Path,
				CommitHash:  commitHash,
			}
			db.SaveInstalledSkill(instSkill)
		}

		os.RemoveAll(tempDir)
		return DownloadResultMsg{Err: nil}
	}
}

// SaveConfigCmd guarda la configuración
func SaveConfigCmd(localPath string, remotes []string) tea.Cmd {
	return func() tea.Msg {
		_ = config.SaveConfig(config.Config{LocalPath: localPath, Remotes: remotes})
		return ConfigSavedMsg{}
	}
}

// DeleteSkillCmd elimina un skill del lock file y del sistema de archivos
func DeleteSkillCmd(skill db.InstalledSkill) tea.Cmd {
	return func() tea.Msg {
		// Eliminar del sistema de archivos
		err := os.RemoveAll(skill.Path)
		if err != nil {
			return NavigateToErrorMsg{Err: err}
		}

		// Eliminar del lock file
		err = db.DeleteInstalledSkill(skill.Path)
		if err != nil {
			return NavigateToErrorMsg{Err: err}
		}

		return NavigateToManageMsg{}
	}
}
