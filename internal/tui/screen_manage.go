package tui

import (
	"fmt"
	"io"

	"os"
	"path/filepath"

	"skli/internal/db"
	"skli/internal/gitrepo"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// installedSkillItem implementa list.Item para un skill instalado
type installedSkillItem struct {
	skill db.InstalledSkill
}

func (i installedSkillItem) Title() string       { return i.skill.Name }
func (i installedSkillItem) Description() string { return i.skill.Description }
func (i installedSkillItem) FilterValue() string { return i.skill.Name }

// manageDelegate es un delegate para el listado de gestión
type manageDelegate struct {
	styles list.DefaultItemStyles
}

func newManageDelegate() manageDelegate {
	styles := list.NewDefaultItemStyles()
	styles.SelectedTitle = styles.SelectedTitle.
		Foreground(lipgloss.Color("#FF0000")).
		BorderForeground(lipgloss.Color("#FF0000"))
	styles.SelectedDesc = styles.SelectedDesc.
		Foreground(lipgloss.Color("#FF0000")).
		BorderForeground(lipgloss.Color("#FF0000"))

	return manageDelegate{styles: styles}
}

func (d manageDelegate) Height() int  { return 2 }
func (d manageDelegate) Spacing() int { return 0 }
func (d manageDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d manageDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(installedSkillItem)
	if !ok {
		return
	}

	title := i.Title()
	desc := i.Description()
	if len(desc) > 80 {
		desc = desc[:77] + "..."
	}

	if index == m.Index() {
		fmt.Fprintf(w, "%s\n%s",
			d.styles.SelectedTitle.Render("➜ "+title),
			d.styles.SelectedDesc.Render("    "+desc))
	} else {
		fmt.Fprintf(w, "%s\n%s",
			d.styles.NormalTitle.Render("  "+title),
			d.styles.NormalDesc.Render("    "+desc))
	}
}

type ManageScreenState int

const (
	ManageStateList ManageScreenState = iota
	ManageStateConfirm
	ManageStateSelectingRemote
	ManageStateInputRemote
	ManageStateUploading
)

// ManageScreen es el modelo para gestionar skills instalados
type ManageScreen struct {
	state         ManageScreenState
	list          list.Model
	skills        []db.InstalledSkill
	toDelete      *db.InstalledSkill
	msg           string // Mensaje de estado/error
	confirmCursor int    // 0 para Sí, 1 para No
	remoteList    list.Model
	remoteInput   textinput.Model
	selectedSkill *db.InstalledSkill // Skill seleccionado para subir
	configRemotes []string           // Remotes configurados
}

// scanLocalSkills busca skills en la carpeta local que no estén en el lockfile
func scanLocalSkills(existingSkills []db.InstalledSkill) []db.InstalledSkill {
	localSkillsPath := "skills" // gitrepo.DefaultSkillsPath // Hardcoded por simplicidad o usar constante
	var newSkills []db.InstalledSkill

	// Mapa para búsqueda rápida
	existingMap := make(map[string]bool)
	for _, s := range existingSkills {
		existingMap[s.Path] = true
	}

	filepath.Walk(localSkillsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == "SKILL.md" {
			// Encontrado un skill local
			dir := filepath.Dir(path)
			relPath, _ := filepath.Rel(".", dir) // Ruta relativa desde root

			if !existingMap[relPath] {
				// Es un skill nuevo/unmanaged
				// Parsear SKILL.md para obtener nombre (o usar nombre carpeta)
				skillName := filepath.Base(dir) // Fallback
				// Intentar leer contenido básico si queremos nombre real (pendiente, por ahora nombre carpeta es suficiente indicativo)

				newSkills = append(newSkills, db.InstalledSkill{
					Name:        skillName + " (Local)",
					Path:        relPath,
					Description: "Local skill (unmanaged)",
					RemoteRepo:  "", // Sin remote asociado
				})
			}
		}
		return nil
	})
	return newSkills
}

// NewManageScreen crea una nueva pantalla de gestión
func NewManageScreen(remotes []string) (ManageScreen, tea.Cmd) {
	lock, _ := db.LoadLockFile()
	skills := lock.Skills

	// Escanear locales
	localSkills := scanLocalSkills(skills)
	skills = append(skills, localSkills...)

	items := make([]list.Item, len(skills))
	for i, s := range skills {
		items[i] = installedSkillItem{skill: s}
	}

	delegate := newManageDelegate()
	l := list.New(items, delegate, 60, 20)
	l.Title = "Gestionar Skills Instalados"
	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("skill", "skills")
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "upload PR")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		}
	}
	l.Styles.Title = TitleStyle

	// Cargar remotes para la selección
	// remotes ya se pasa como argumento

	ti := textinput.New()
	ti.Placeholder = "https://github.com/usuario/repo.git"
	ti.CharLimit = 156
	ti.Width = 50

	return ManageScreen{
		state:         ManageStateList,
		list:          l,
		skills:        skills,
		confirmCursor: 1, // Por defecto en No por seguridad
		remoteInput:   ti,
		configRemotes: remotes,
	}, nil
}

func (s ManageScreen) Init() tea.Cmd {
	return nil
}

func (s ManageScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch s.state {
	case ManageStateList:
		return s.updateList(msg)
	case ManageStateConfirm:
		return s.updateConfirm(msg)
	case ManageStateSelectingRemote:
		return s.updateSelectingRemote(msg)
	case ManageStateInputRemote:
		return s.updateInputRemote(msg)
	case ManageStateUploading:
		return s.updateUploading(msg)
	}
	return s, nil
}

func (s ManageScreen) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.list.SetSize(msg.Width, msg.Height-4)
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "d", "backspace":
			if item, ok := s.list.SelectedItem().(installedSkillItem); ok {
				s.toDelete = &item.skill
				s.state = ManageStateConfirm
				s.confirmCursor = 1 // Reset a No
				return s, nil
			}
		case "esc", "q":
			return s, func() tea.Msg { return QuitMsg{} }
		case "u":
			if item, ok := s.list.SelectedItem().(installedSkillItem); ok {
				s.selectedSkill = &item.skill
				// Preparar lista de remotes
				var items []list.Item
				seen := make(map[string]bool)

				// 1. Añadir remote del skill si existe (ORIGIN)
				if item.skill.RemoteRepo != "" {
					items = append(items, remoteItem{
						url:         item.skill.RemoteRepo,
						displayName: fmt.Sprintf("%s (Origin)", item.skill.RemoteRepo),
					})
					seen[item.skill.RemoteRepo] = true
				}

				// 2. Añadir remotes configurados
				for _, r := range s.configRemotes {
					if !seen[r] {
						items = append(items, remoteItem{url: r})
						seen[r] = true
					}
				}

				// 3. Añadir opcion custom
				items = append(items, customURLItem{})

				delegate := newRemoteDelegate()
				l := list.New(items, delegate, 60, 14)
				l.Title = "Selecciona repositorio destino"
				l.SetShowStatusBar(false)
				l.SetFilteringEnabled(false)
				l.Styles.Title = TitleStyle
				s.remoteList = l

				s.state = ManageStateSelectingRemote
				return s, nil
			}
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

func (s ManageScreen) updateSelectingRemote(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.state = ManageStateList
			return s, nil
		case "enter":
			selected := s.remoteList.SelectedItem()
			if selected == nil {
				return s, nil
			}
			switch item := selected.(type) {
			case remoteItem:
				// Iniciar upload con esta URL
				s.state = ManageStateUploading
				s.msg = fmt.Sprintf("Iniciando subida de '%s' a %s...", s.selectedSkill.Name, item.url)
				return s, UploadSkillCmd(*s.selectedSkill, item.url)
			case customURLItem:
				s.state = ManageStateInputRemote
				s.remoteInput.Focus()
				s.remoteInput.SetValue("")
				return s, textinput.Blink
			}
		}
	}
	var cmd tea.Cmd
	s.remoteList, cmd = s.remoteList.Update(msg)
	return s, cmd
}

func (s ManageScreen) updateInputRemote(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.state = ManageStateSelectingRemote
			return s, nil
		case "enter":
			url := s.remoteInput.Value()
			if url != "" {
				s.state = ManageStateUploading
				s.msg = fmt.Sprintf("Iniciando subida de '%s' a %s...", s.selectedSkill.Name, url)
				return s, UploadSkillCmd(*s.selectedSkill, url)
			}
		}
	}
	var cmd tea.Cmd
	s.remoteInput, cmd = s.remoteInput.Update(msg)
	return s, cmd
}

func (s ManageScreen) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "right", "l", "tab":
			if s.confirmCursor == 0 {
				s.confirmCursor = 1
			} else {
				s.confirmCursor = 0
			}
		case "y", "Y":
			s.confirmCursor = 0
			if s.toDelete != nil {
				return s, DeleteSkillCmd(*s.toDelete)
			}
		case "n", "N", "esc":
			s.state = ManageStateList
			s.toDelete = nil
			return s, nil
		case "enter":
			if s.confirmCursor == 0 {
				if s.toDelete != nil {
					return s, DeleteSkillCmd(*s.toDelete)
				}
			} else {
				s.state = ManageStateList
				s.toDelete = nil
				return s, nil
			}
		}
	}
	return s, nil
}

func (s ManageScreen) updateUploading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case UploadSkillMsg:
		if msg.Err != nil {
			s.msg = fmt.Sprintf("Error: %v", msg.Err)
			// Volver al listado tras un error, o dejar que el usuario pulse esc?
			// Mejor dejar leer el error.
			return s, nil
		}
		s.msg = fmt.Sprintf("PR Creado: %s", msg.PRURL)
		return s, nil
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "enter" {
			s.state = ManageStateList
			s.msg = ""
			return s, nil
		}
	}
	return s, nil
}

func (s ManageScreen) View() string {
	if s.state == ManageStateSelectingRemote {
		return "\n" + s.remoteList.View()
	}
	if s.state == ManageStateInputRemote {
		return fmt.Sprintf("\n  Introduce la URL del repositorio destino:\n\n  %s\n\n  %s",
			s.remoteInput.View(),
			HelpStyle.Render("enter: confirmar • esc: volver"),
		)
	}
	if s.state == ManageStateUploading {
		return fmt.Sprintf("\n  Subiendo Skill...\n\n  %s\n\n  %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Render(s.msg),
			HelpStyle.Render("esc/enter: volver"),
		)
	}

	if s.state == ManageStateConfirm {
		var yes, no string
		if s.confirmCursor == 0 {
			yes = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true).Render("➜ [ Sí ]")
			no = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  [ No ]")
		} else {
			yes = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  [ Sí ]")
			no = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true).Render("➜ [ No ]")
		}

		return fmt.Sprintf(
			"\n  ¿Estás DE ACUERDO en eliminar el skill %s?\n\n  Ruta: %s\n\n  %s    %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true).Render(s.toDelete.Name),
			DimStyle.Render(s.toDelete.Path),
			yes,
			no,
		)
	}

	if len(s.skills) == 0 {
		return "\n  No hay skills instalados.\n\n" + HelpStyle.Render("  q: salir")
	}

	if s.msg != "" {
		return fmt.Sprintf("\n  %s\n\n%s", lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render(s.msg), s.list.View())
	}

	originalTitle := s.list.Title
	s.list.Title = fmt.Sprintf("%s (Pág. %d/%d)", originalTitle, s.list.Paginator.Page+1, s.list.Paginator.TotalPages)
	view := s.list.View()
	s.list.Title = originalTitle
	return view
}

// UploadSkillMsg es el mensaje con el resultado de la subida
type UploadSkillMsg struct {
	PRURL string
	Err   error
}

// UploadSkillCmd realiza la operación de subida y PR
func UploadSkillCmd(skill db.InstalledSkill, targetRemoteURL string) tea.Cmd {
	return func() tea.Msg {
		if !gitrepo.CheckGhInstalled() {
			return UploadSkillMsg{Err: fmt.Errorf("gh CLI no está instalada. Instálala para usar esta función.")}
		}

		// 1. Clonar repo temporal (usando el targetRemoteURL seleccionado)
		tempDir, err := gitrepo.CloneForPush(targetRemoteURL)
		if err != nil {
			return UploadSkillMsg{Err: fmt.Errorf("error clonando %s: %w", targetRemoteURL, err)}
		}
		defer os.RemoveAll(tempDir)

		// 2. Crear rama
		branchName, err := gitrepo.PrepareSkillBranch(tempDir, skill.Name)
		if err != nil {
			return UploadSkillMsg{Err: err}
		}

		// 3. Buscar ruta del skill en repo
		// Si estamos subiendo a un nuevo repo (o uno custom), es posible que el skill NO exista aún.
		// Si existe, intentamos actualizarlo en su sitio.
		// Si no existe, ¿dónde lo ponemos? Por defecto en skills/<nombre-skill> o en root/<nombre-skill> según estructura.
		// Vamos a asumir que lo ponemos en skills/<nombre-carpeta-local>

		repoSkillPath := skill.RemotePath
		// Si estamos cambiando de remote, el RemotePath antiguo puede no valer.
		// Intentamos buscarlo primero por si acaso es una actualización.
		foundPath, err := gitrepo.FindSkillInRepo(tempDir, skill.Name)
		if err == nil {
			repoSkillPath = foundPath
		} else {
			// No encontrado, es un skill nuevo en este repo
			// Definimos ruta por defecto
			skillFolderName := filepath.Base(skill.Path)
			repoSkillPath = filepath.Join("skills", skillFolderName)
		}

		// 4. Copiar archivos locales
		localSkillPath := skill.Path
		err = gitrepo.CopySkillFiles(tempDir, localSkillPath, repoSkillPath)
		if err != nil {
			return UploadSkillMsg{Err: err}
		}

		// 5. Push y PR
		prURL, err := gitrepo.PushAndCreatePR(tempDir, branchName, skill.Name, skill.Description)
		if err != nil {
			return UploadSkillMsg{Err: err}
		}

		return UploadSkillMsg{PRURL: prURL}
	}
}
