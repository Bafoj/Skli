package gitrepo

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SkillInfo contiene los metadatos de un skill
type SkillInfo struct {
	Name        string
	Description string
	Path        string // Ruta relativa dentro del repo (para copiar)
}

// CloneAndScan usa sparse-checkout para clonar SOLO la carpeta skills/
func CloneAndScan(remoteURL string) ([]SkillInfo, string, error) {
	tempDir, err := os.MkdirTemp("", "skli-repo-*")
	if err != nil {
		return nil, "", fmt.Errorf("error creando dir temporal: %w", err)
	}

	// 1. Inicializar un repo vacío
	if err := runGit(tempDir, "init"); err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("error inicializando repo: %w", err)
	}

	// 2. Añadir el remote
	if err := runGit(tempDir, "remote", "add", "origin", remoteURL); err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("error añadiendo remote: %w", err)
	}

	// 3. Configurar sparse-checkout
	if err := runGit(tempDir, "config", "core.sparseCheckout", "true"); err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("error configurando sparse-checkout: %w", err)
	}

	// 4. Definir qué carpeta descargar (solo skills/)
	sparseFile := filepath.Join(tempDir, ".git", "info", "sparse-checkout")
	if err := os.WriteFile(sparseFile, []byte("skills/\n"), 0644); err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("error escribiendo sparse-checkout: %w", err)
	}

	// 5. Fetch solo la rama principal con depth 1 (sin historial)
	if err := runGit(tempDir, "fetch", "--depth", "1", "origin", "HEAD"); err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("error descargando repo: %w", err)
	}

	// 6. Checkout
	if err := runGit(tempDir, "checkout", "FETCH_HEAD"); err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("error haciendo checkout: %w", err)
	}

	// Buscar recursivamente archivos SKILL.md
	skillsPath := filepath.Join(tempDir, "skills")
	skills, err := findSkills(skillsPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, "", err
	}

	if len(skills) == 0 {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("no se encontraron skills (archivos SKILL.md) en el repositorio")
	}

	return skills, tempDir, nil
}

// findSkills busca recursivamente archivos SKILL.md y extrae sus metadatos
func findSkills(baseDir string) ([]SkillInfo, error) {
	var skills []SkillInfo

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Ignorar errores de permisos, etc.
		}

		if !info.IsDir() && info.Name() == "SKILL.md" {
			skill, parseErr := parseSkillFile(path, baseDir)
			if parseErr == nil && skill.Name != "" {
				skills = append(skills, skill)
			}
		}
		return nil
	})

	return skills, err
}

// parseSkillFile extrae el nombre y descripción del YAML frontmatter
func parseSkillFile(filePath, baseDir string) (SkillInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return SkillInfo{}, err
	}
	defer file.Close()

	var skill SkillInfo
	// Obtener la carpeta del skill (padre de SKILL.md)
	skillDir := filepath.Dir(filePath)
	relPath, _ := filepath.Rel(baseDir, skillDir)
	skill.Path = relPath

	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// Detectar inicio/fin del frontmatter
		if strings.TrimSpace(line) == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				break // Fin del frontmatter
			}
		}

		if inFrontmatter {
			if strings.HasPrefix(line, "name:") {
				skill.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			} else if strings.HasPrefix(line, "description:") {
				skill.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			}
		}

		// Solo leer las primeras líneas para evitar parsear todo el archivo
		if lineCount > 20 {
			break
		}
	}

	return skill, scanner.Err()
}

// InstallSkills copia las carpetas seleccionadas al PWD
func InstallSkills(tempRepoPath string, selectedSkills []SkillInfo) error {
	destBase := "skills"
	if err := os.MkdirAll(destBase, 0755); err != nil {
		return fmt.Errorf("error creando carpeta local skills: %w", err)
	}

	for _, skill := range selectedSkills {
		src := filepath.Join(tempRepoPath, "skills", skill.Path)
		dest := filepath.Join(destBase, filepath.Base(skill.Path))

		cmd := exec.Command("cp", "-r", src, dest)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error copiando skill %s: %w", skill.Name, err)
		}
	}

	return nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}
