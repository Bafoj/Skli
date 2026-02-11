package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"skli/internal/db"
	"skli/internal/gitrepo"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	getRemoteHashFn    = gitrepo.GetRemoteHash
	cloneAndScanFn     = gitrepo.CloneAndScan
	saveInstalledFn    = db.SaveInstalledSkill
	removeAllFn        = os.RemoveAll
	statFn             = os.Stat
	copyDirFn          = copyDir
)

// SyncResult contiene el resultado de la sincronización
type SyncResult struct {
	SkillName string
	Updated   bool
	Skipped   bool // Sin cambios (hash igual)
	Error     error
}

// SyncAllSkills sincroniza todos los skills instalados desde sus repos de origen
func SyncAllSkills() ([]SyncResult, error) {
	grouped, err := db.GetSkillsByRepo()
	if err != nil {
		return nil, fmt.Errorf("error getting skills from skli.lock: %w", err)
	}

	if len(grouped) == 0 {
		return nil, nil
	}

	var allResults []SyncResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	totalRepos := len(grouped)
	processedRepos := 0

	// Spinner animation
	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	done := make(chan bool)

	// Iniciar animación de progreso
	go func() {
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				fmt.Printf("\r%s Checking repos... (%d/%d)",
					infoStyle.Render(spinnerChars[i%len(spinnerChars)]),
					processedRepos, totalRepos)
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Procesar cada repo en paralelo
	for repoURL, skills := range grouped {
		wg.Add(1)
		go func(repoURL string, skills []db.InstalledSkill) {
			defer wg.Done()

			results := syncRepo(repoURL, skills)

			mu.Lock()
			allResults = append(allResults, results...)
			processedRepos++
			mu.Unlock()
		}(repoURL, skills)
	}

	wg.Wait()
	done <- true
	fmt.Print("\r\033[K") // Limpiar línea

	return allResults, nil
}

// syncRepo sincroniza todos los skills de un repo específico
func syncRepo(repoURL string, skills []db.InstalledSkill) []SyncResult {
	var results []SyncResult

	// Usamos el RemoteRoot del primer skill como referencia para el repo
	// (Asumimos que todos los skills del mismo repo se buscan en la misma ruta)
	skillsPath := ""
	if len(skills) > 0 {
		skillsPath = skills[0].RemoteRoot
	}

	// 1. Primero verificar el hash remoto SIN clonar
	remoteHash, err := getRemoteHashFn(repoURL)
	if err != nil {
		for _, s := range skills {
			results = append(results, SyncResult{
				SkillName: s.Name,
				Error:     fmt.Errorf("error checking repo: %w", err),
			})
		}
		return results
	}

	// 2. Verificar si todos los skills tienen el mismo hash (sin cambios)
	// Y verificar si los archivos locales todavía existen
	allUpToDate := true
	for _, s := range skills {
		if s.CommitHash != remoteHash {
			allUpToDate = false
			break
		}

		// Verificar si la carpeta del skill existe localmente
		if _, err := statFn(s.Path); os.IsNotExist(err) {
			allUpToDate = false
			break
		}
	}

	// 3. Si todo está actualizado, no descargar nada
	if allUpToDate {
		for _, s := range skills {
			results = append(results, SyncResult{
				SkillName: s.Name,
				Skipped:   true,
			})
		}
		return results
	}

	// 4. Solo si hay cambios, clonar el repo
	scanRes, err := cloneAndScanFn(repoURL, skillsPath)
	if err != nil {
		for _, s := range skills {
			results = append(results, SyncResult{
				SkillName: s.Name,
				Error:     fmt.Errorf("error cloning repo: %w", err),
			})
		}
		return results
	}
	defer removeAllFn(scanRes.TempDir)

	// Usar el path detectado para mayor consistencia
	skillsPath = scanRes.SkillsPath

	// Crear un mapa de skills remotos para búsqueda rápida
	remoteMap := make(map[string]gitrepo.SkillInfo)
	for _, rs := range scanRes.Skills {
		remoteMap[rs.Path] = rs
	}

	// Actualizar cada skill instalado
	for _, installed := range skills {
		remote, exists := remoteMap[installed.RemotePath]
		if !exists {
			results = append(results, SyncResult{
				SkillName: installed.Name,
				Error:     fmt.Errorf("skill no longer exists in remote repo"),
			})
			continue
		}

		// Si el hash no ha cambiado Y el archivo existe, saltar
		if installed.CommitHash == scanRes.CommitHash {
			if _, err := statFn(installed.Path); err == nil {
				results = append(results, SyncResult{
					SkillName: installed.Name,
					Skipped:   true,
				})
				continue
			}
		}

		// Copiar la versión más reciente
		// Usamos el skillsPath del repo
		var src string
		if skillsPath == "." {
			src = filepath.Join(scanRes.TempDir, remote.Path)
		} else {
			src = filepath.Join(scanRes.TempDir, skillsPath, remote.Path)
		}

		// El destino ya está guardado en installed.Path (ej: ".cursor/skills/nombre-skill")
		dest := installed.Path

		// Eliminar la versión anterior
		removeAllFn(dest)

		// Copiar la nueva versión
		if err := copyDirFn(src, dest); err != nil {
			results = append(results, SyncResult{
				SkillName: installed.Name,
				Error:     fmt.Errorf("error copying: %w", err),
			})
			continue
		}

		// Actualizar metadatos en el lock file con el nuevo hash y el mismo path base
		saveInstalledFn(db.InstalledSkill{
			Name:        remote.Name,
			Description: remote.Description,
			Path:        installed.Path,
			RemoteRepo:  repoURL,
			RemoteRoot:  skillsPath,
			RemotePath:  remote.Path,
			CommitHash:  scanRes.CommitHash,
		})

		results = append(results, SyncResult{
			SkillName: installed.Name,
			Updated:   true,
		})
	}

	return results
}

func copyDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		targetPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, info.Mode())
	})
}
