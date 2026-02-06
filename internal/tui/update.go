package tui

import (
	"os"
	"path/filepath"
	"skli/internal/db"
	"skli/internal/gitrepo"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Mensajes personalizados para operaciones asíncronas
type scanMsg struct {
	skills    []gitrepo.SkillInfo
	tempDir   string
	remoteURL string
	err       error
}

type downloadMsg struct {
	err error
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.Quitting = true
			return m, tea.Quit
		case "esc":
			if m.State == StateInputRemote {
				m.Quitting = true
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.WindowWidth = msg.Width
		m.WindowHeight = msg.Height
		return m, nil

	case scanMsg:
		if msg.err != nil {
			m.State = StateError
			m.ErrorMessage = msg.err.Error()
			return m, nil
		}
		m.State = StateSelectingSkills
		m.Skills = make([]Skill, len(msg.skills))
		for i, info := range msg.skills {
			m.Skills[i] = Skill{Info: info}
		}
		m.TempDir = msg.tempDir
		m.RemoteURL = msg.remoteURL
		return m, nil

	case downloadMsg:
		if msg.err != nil {
			m.State = StateError
			m.ErrorMessage = msg.err.Error()
		} else {
			m.State = StateDone
		}
		return m, nil

	case spinner.TickMsg:
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}

	// Lógica específica por estado
	switch m.State {
	case StateInputRemote:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				url := m.TextInput.Value()
				if url == "" {
					return m, nil
				}
				m.State = StateScanning
				return m, tea.Batch(scanRepoCmd(url), m.Spinner.Tick)
			}
		}
		m.TextInput, cmd = m.TextInput.Update(msg)
		return m, cmd

	case StateSelectingSkills:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.Cursor > 0 {
					m.Cursor--
				}
			case "down", "j":
				if m.Cursor < len(m.Skills)-1 {
					m.Cursor++
				}
			case " ":
				m.Skills[m.Cursor].Selected = !m.Skills[m.Cursor].Selected
			case "enter":
				var selected []gitrepo.SkillInfo
				for _, s := range m.Skills {
					if s.Selected {
						selected = append(selected, s.Info)
					}
				}
				if len(selected) > 0 {
					m.State = StateDownloading
					return m, tea.Batch(downloadSkillsCmd(m.TempDir, m.RemoteURL, selected), m.Spinner.Tick)
				}
			}
		}

	case StateDone, StateError:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "r" && m.State == StateError {
				m.State = StateInputRemote
				m.TextInput.Reset()
				return m, nil
			}
			m.Quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// Comandos de Bubble Tea
func scanRepoCmd(url string) tea.Cmd {
	return func() tea.Msg {
		skills, tempDir, err := gitrepo.CloneAndScan(url)
		return scanMsg{skills: skills, tempDir: tempDir, remoteURL: url, err: err}
	}
}

func downloadSkillsCmd(tempDir, remoteURL string, selected []gitrepo.SkillInfo) tea.Cmd {
	return func() tea.Msg {
		err := gitrepo.InstallSkills(tempDir, selected)
		if err != nil {
			os.RemoveAll(tempDir)
			return downloadMsg{err: err}
		}

		// Guardar en la base de datos
		database, dbErr := db.InitDB()
		if dbErr == nil {
			defer database.Close()
			for _, skill := range selected {
				localPath := filepath.Base(skill.Path)
				instSkill := db.InstalledSkill{
					Name:        skill.Name,
					Description: skill.Description,
					Path:        localPath,
					RemoteRepo:  remoteURL,
					RemotePath:  skill.Path,
				}
				db.SaveInstalledSkill(database, instSkill)
			}
		}

		os.RemoveAll(tempDir)
		return downloadMsg{err: nil}
	}
}
