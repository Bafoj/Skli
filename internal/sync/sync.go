package sync

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"skli/internal/db"
	"skli/internal/gitrepo"
)

// SyncResult contiene el resultado de la sincronización
type SyncResult struct {
	SkillName string
	Updated   bool
	Error     error
}

// SyncAllSkills sincroniza todos los skills instalados desde sus repos de origen
func SyncAllSkills(database *sql.DB) ([]SyncResult, error) {
	grouped, err := db.GetSkillsByRepo(database)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo skills: %w", err)
	}

	if len(grouped) == 0 {
		return nil, nil
	}

	var allResults []SyncResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Procesar cada repo en paralelo
	for repoURL, skills := range grouped {
		wg.Add(1)
		go func(repoURL string, skills []db.InstalledSkill) {
			defer wg.Done()

			results := syncRepo(repoURL, skills)

			mu.Lock()
			allResults = append(allResults, results...)
			mu.Unlock()
		}(repoURL, skills)
	}

	wg.Wait()
	return allResults, nil
}

// syncRepo sincroniza todos los skills de un repo específico
func syncRepo(repoURL string, skills []db.InstalledSkill) []SyncResult {
	var results []SyncResult

	// Clonar el repo una sola vez
	remoteSkills, tempDir, err := gitrepo.CloneAndScan(repoURL)
	if err != nil {
		// Marcar todos los skills de este repo como error
		for _, s := range skills {
			results = append(results, SyncResult{
				SkillName: s.Name,
				Error:     fmt.Errorf("error clonando repo: %w", err),
			})
		}
		return results
	}
	defer os.RemoveAll(tempDir)

	// Crear un mapa de skills remotos para búsqueda rápida
	remoteMap := make(map[string]gitrepo.SkillInfo)
	for _, rs := range remoteSkills {
		remoteMap[rs.Path] = rs
	}

	// Actualizar cada skill instalado
	for _, installed := range skills {
		remote, exists := remoteMap[installed.RemotePath]
		if !exists {
			results = append(results, SyncResult{
				SkillName: installed.Name,
				Error:     fmt.Errorf("skill ya no existe en el repo remoto"),
			})
			continue
		}

		// Copiar la versión más reciente
		src := filepath.Join(tempDir, "skills", remote.Path)
		dest := filepath.Join("skills", installed.Path)

		// Eliminar la versión anterior
		os.RemoveAll(dest)

		// Copiar la nueva versión
		if err := copyDir(src, dest); err != nil {
			results = append(results, SyncResult{
				SkillName: installed.Name,
				Error:     fmt.Errorf("error copiando: %w", err),
			})
			continue
		}

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
