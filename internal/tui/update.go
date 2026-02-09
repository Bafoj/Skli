package tui

import (
	"os"
	"path/filepath"
	"skli/internal/config"
	"skli/internal/db"
	"skli/internal/gitrepo"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Mensajes personalizados para operaciones asíncronas
type scanMsg struct {
	result    gitrepo.ScanResult
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
		res := msg.result
		m.State = StateSelectingSkills
		m.Skills = make([]Skill, len(res.Skills))
		for i, info := range res.Skills {
			m.Skills[i] = Skill{Info: info}
		}
		m.TempDir = res.TempDir
		m.RemoteURL = msg.remoteURL
		m.SkillsRoot = res.SkillsPath
		m.CommitHash = res.CommitHash
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
				return m, tea.Batch(scanRepoCmd(url, m.SkillsRoot), m.Spinner.Tick)
			}
		}
		m.TextInput, cmd = m.TextInput.Update(msg)
		return m, cmd

	case StateInputNewRemote:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				url := m.TextInput.Value()
				if url == "" {
					return m, nil
				}
				// Añadir remote y volver a gestionar
				m.Remotes = append(m.Remotes, url)
				_ = config.SaveConfig(config.Config{LocalPath: m.ConfigLocalPath, Remotes: m.Remotes})
				m.State = StateManageRemotes
				m.TextInput.Reset()
				return m, nil
			} else if msg.String() == "esc" {
				m.State = StateManageRemotes
				m.TextInput.Reset()
				return m, nil
			}
		}
		m.TextInput, cmd = m.TextInput.Update(msg)
		return m, cmd

	case StateSelectingRemote:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.RemoteCursor > 0 {
					m.RemoteCursor--
				}
			case "down", "j":
				// Remotes + 1 (Custom)
				if m.RemoteCursor < len(m.Remotes) { // len(m.Remotes) is the index of "Custom"
					m.RemoteCursor++
				}
			case "enter":
				// Custom seleccionado (última posición)
				if m.RemoteCursor == len(m.Remotes) {
					m.State = StateInputRemote
					m.TextInput.Focus()
					return m, nil
				}
				// Remote existente seleccionado
				if m.RemoteCursor < len(m.Remotes) {
					m.RemoteURL = m.Remotes[m.RemoteCursor]
					m.State = StateScanning
					return m, tea.Batch(scanRepoCmd(m.RemoteURL, m.SkillsRoot), m.Spinner.Tick)
				}
			}
		}

	case StateManageRemotes:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.State = StateConfigMenu
				return m, nil
			case "up", "k":
				if m.RemoteCursor > 0 {
					m.RemoteCursor--
				}
			case "down", "j":
				// Remotes + 1 (Add New)
				if m.RemoteCursor < len(m.Remotes) {
					m.RemoteCursor++
				}
			case "enter":
				// Add New seleccionado (última posición)
				if m.RemoteCursor == len(m.Remotes) {
					m.State = StateInputNewRemote
					m.TextInput.Focus()
					m.TextInput.SetValue("")
					return m, nil
				}
			case "backspace", "d": // Eliminar remote
				if m.RemoteCursor < len(m.Remotes) {
					// Eliminar el elemento en m.RemoteCursor
					m.Remotes = append(m.Remotes[:m.RemoteCursor], m.Remotes[m.RemoteCursor+1:]...)
					_ = config.SaveConfig(config.Config{LocalPath: m.ConfigLocalPath, Remotes: m.Remotes})
					// Ajustar cursor si es necesario
					if m.RemoteCursor >= len(m.Remotes) && m.RemoteCursor > 0 {
						m.RemoteCursor--
					}
				}
			}
		}

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
					if m.ConfigLocalPath == "" {
						m.State = StateSelectingEditor
						return m, nil
					}
					m.State = StateDownloading
					return m, tea.Batch(downloadSkillsCmd(m.TempDir, m.RemoteURL, m.SkillsRoot, m.ConfigLocalPath, m.CommitHash, selected), m.Spinner.Tick)
				}
			}
		}

	case StateConfigMenu:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.ConfigCursor > 0 {
					m.ConfigCursor--
				}
			case "down", "j":
				if m.ConfigCursor < 2 { // 0: Remotes, 1: Local Path, 2: Confirm
					m.ConfigCursor++
				}
			case "enter":
				if m.ConfigCursor == 0 {
					m.State = StateManageRemotes
					m.RemoteCursor = 0
				} else if m.ConfigCursor == 1 {
					m.State = StateSelectingEditor
				} else if m.ConfigCursor == 2 {
					// Guardar definitivamente al confirmar
					_ = config.SaveConfig(config.Config{LocalPath: m.ConfigLocalPath, Remotes: m.Remotes})
					m.State = StateDone
				}
				return m, nil
			case "q", "esc":
				m.Quitting = true
				return m, tea.Quit
			}
		}

	case StateSelectingEditor:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "backspace":
				m.State = StateConfigMenu
				return m, nil
			case "up", "k":
				if m.EditorCursor > 0 {
					m.EditorCursor--
				}
			case "down", "j":
				if m.EditorCursor < len(Editors)-1 {
					m.EditorCursor++
				}
			case "enter":
				editor := Editors[m.EditorCursor]
				destPath := editor.Path
				if editor.Name == "Custom" {
					destPath = "skills"
				}

				m.ConfigLocalPath = destPath

				// En modo config, no guardamos todavía, volvemos al menú para confirmar
				if m.ConfigMode {
					m.State = StateConfigMenu
					return m, nil
				}

				// Si estamos en medio de una instalación, sí guardamos y procedemos
				m.State = StateDownloading

				// Guardar como configuración predeterminada para el futuro
				_ = config.SaveConfig(config.Config{LocalPath: destPath, Remotes: m.Remotes})

				var selected []gitrepo.SkillInfo
				for _, s := range m.Skills {
					if s.Selected {
						selected = append(selected, s.Info)
					}
				}

				return m, tea.Batch(
					downloadSkillsCmd(m.TempDir, m.RemoteURL, m.SkillsRoot, m.ConfigLocalPath, m.CommitHash, selected),
					m.Spinner.Tick,
				)
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
func scanRepoCmd(url, defaultSkillsPath string) tea.Cmd {
	return func() tea.Msg {
		res, err := gitrepo.CloneAndScan(url, defaultSkillsPath)
		return scanMsg{result: res, remoteURL: url, err: err}
	}
}

func downloadSkillsCmd(tempDir, remoteURL, skillsPath, localPath, commitHash string, selected []gitrepo.SkillInfo) tea.Cmd {
	return func() tea.Msg {
		err := gitrepo.InstallSkills(tempDir, skillsPath, localPath, selected)
		if err != nil {
			os.RemoveAll(tempDir)
			return downloadMsg{err: err}
		}

		// Guardar en el lock file (TOML) con el hash del commit
		for _, skill := range selected {
			// Usamos el nombre consistente (plano) para la ruta local
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
		return downloadMsg{err: nil}
	}
}
