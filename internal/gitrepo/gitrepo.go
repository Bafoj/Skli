package gitrepo

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"skli/internal/skillmeta"
)

// DefaultSkillsPath es el path por defecto donde buscar skills
const DefaultSkillsPath = "skills"

// RepoInfo contiene la información parseada de un repositorio
type RepoInfo struct {
	BaseURL string // URL clonable (ej: https://github.com/user/repo)
	Branch  string // Rama/Tag (ej: main)
	SubPath string // Directorio interno (ej: skills/.experimental)
}

// ParseGitURL analiza una URL de repositorio y extrae el repo base, rama y path interno
// Soporta GitHub (/tree/), Bitbucket (/src/) y URLs estándar (HTTPS, SSH)
func ParseGitURL(urlStr string) RepoInfo {
	info := RepoInfo{
		BaseURL: urlStr,
		Branch:  "HEAD",
		SubPath: "",
	}

	urlStr = strings.TrimSuffix(urlStr, "/")

	// Caso SSH (ej: git@github.com:user/repo.git)
	if strings.Contains(urlStr, "@") && strings.Contains(urlStr, ":") {
		// En SSH no solemos tener "tree/rama/path" en la URL misma
		return info
	}

	var delimiter string
	if strings.Contains(urlStr, "github.com") {
		delimiter = "/tree/"
	} else if strings.Contains(urlStr, "bitbucket.org") {
		delimiter = "/src/"
	}

	if delimiter != "" {
		parts := strings.Split(urlStr, delimiter)
		if len(parts) == 2 {
			info.BaseURL = parts[0]

			// El resto es "rama/path"
			remaining := parts[1]
			subParts := strings.Split(remaining, "/")
			if len(subParts) > 0 {
				info.Branch = subParts[0]
				if len(subParts) > 1 {
					info.SubPath = strings.Join(subParts[1:], "/")
				}
			}
			return info
		}
	}

	// Fallback para URLs estándar
	return info
}

// SkillInfo contiene los metadatos de un skill
type SkillInfo struct {
	Name        string
	Description string
	Path        string // Ruta relativa dentro del repo (para copiar)
}

// ScanResult contiene el resultado del escaneo de un repositorio
type ScanResult struct {
	Skills     []SkillInfo
	TempDir    string
	CommitHash string
	SkillsPath string
}

// CloneAndScan usa sparse-checkout para clonar SOLO la carpeta especificada
func CloneAndScan(remoteURL, skillsPath string) (ScanResult, error) {
	repoInfo := ParseGitURL(remoteURL)

	// Si se nos pasa un path vacío pero la URL tenía uno, lo usamos
	if skillsPath == "" || skillsPath == DefaultSkillsPath {
		if repoInfo.SubPath != "" {
			skillsPath = repoInfo.SubPath
		}
	}

	if skillsPath == "" {
		skillsPath = DefaultSkillsPath
	}

	tempDir, err := os.MkdirTemp("", "skli-repo-*")
	if err != nil {
		return ScanResult{}, fmt.Errorf("error creating temp dir: %w", err)
	}

	// 1. Inicializar un repo vacío
	if err := runGit(tempDir, "init"); err != nil {
		os.RemoveAll(tempDir)
		return ScanResult{}, fmt.Errorf("error initializing repo: %w", err)
	}

	// 2. Añadir el remote (usamos la BaseURL parseada)
	if err := runGit(tempDir, "remote", "add", "origin", repoInfo.BaseURL); err != nil {
		os.RemoveAll(tempDir)
		return ScanResult{}, fmt.Errorf("error adding remote: %w", err)
	}

	// 3. Configurar sparse-checkout
	if err := runGit(tempDir, "config", "core.sparseCheckout", "true"); err != nil {
		os.RemoveAll(tempDir)
		return ScanResult{}, fmt.Errorf("error configuring sparse-checkout: %w", err)
	}

	// 4. Definir qué carpeta descargar
	sparseFile := filepath.Join(tempDir, ".git", "info", "sparse-checkout")
	sparseContent := skillsPath
	if skillsPath != "." {
		sparseContent = skillsPath + "/"
	}
	if err := os.WriteFile(sparseFile, []byte(sparseContent+"\n"), 0644); err != nil {
		os.RemoveAll(tempDir)
		return ScanResult{}, fmt.Errorf("error writing sparse-checkout: %w", err)
	}

	// 5. Fetch solo la rama especificada (o HEAD) con depth 1
	if err := runGit(tempDir, "fetch", "--depth", "1", "origin", repoInfo.Branch); err != nil {
		os.RemoveAll(tempDir)
		return ScanResult{}, fmt.Errorf("error fetching repo (branch %s): %w", repoInfo.Branch, err)
	}

	// 6. Checkout
	if err := runGit(tempDir, "checkout", "FETCH_HEAD"); err != nil {
		os.RemoveAll(tempDir)
		return ScanResult{}, fmt.Errorf("error checking out: %w", err)
	}

	// 7. Obtener el hash del commit actual
	commitHash, err := getCommitHash(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		return ScanResult{}, fmt.Errorf("error getting commit hash: %w", err)
	}

	// Buscar recursivamente archivos SKILL.md
	searchPath := tempDir
	if skillsPath != "." {
		searchPath = filepath.Join(tempDir, skillsPath)
	}
	skills, err := findSkills(searchPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return ScanResult{}, err
	}

	if len(skills) == 0 {
		os.RemoveAll(tempDir)
		return ScanResult{}, fmt.Errorf("no skills found (SKILL.md files) in '%s'", skillsPath)
	}

	return ScanResult{
		Skills:     skills,
		TempDir:    tempDir,
		CommitHash: commitHash,
		SkillsPath: skillsPath,
	}, nil
}

// getCommitHash obtiene el hash del commit actual del repo
func getCommitHash(repoDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetRemoteHash obtiene el hash del HEAD remoto sin clonar el repo
// Esto permite verificar si hay cambios antes de descargar nada
func GetRemoteHash(repoURL string) (string, error) {
	repoInfo := ParseGitURL(repoURL)

	cmd := exec.Command("git", "ls-remote", repoInfo.BaseURL, repoInfo.Branch)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting remote hash: %w", err)
	}

	parts := strings.Fields(string(output))
	if len(parts) == 0 {
		return "", fmt.Errorf("could not get repository hash")
	}
	return parts[0], nil
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
	var skill SkillInfo
	// Obtener la carpeta del skill (padre de SKILL.md)
	skillDir := filepath.Dir(filePath)
	relPath, _ := filepath.Rel(baseDir, skillDir)
	skill.Path = relPath

	meta, err := skillmeta.ParseFile(filePath, 20)
	if err != nil {
		return SkillInfo{}, err
	}
	skill.Name = meta.Name
	skill.Description = meta.Description

	return skill, nil
}

// InstallSkills copia las carpetas seleccionadas al PWD de forma plana
func InstallSkills(tempRepoPath, skillsPath, localPath string, selectedSkills []SkillInfo) error {
	if skillsPath == "" {
		skillsPath = DefaultSkillsPath
	}
	if localPath == "" {
		localPath = "skills"
	}

	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("error creating local folder %s: %w", localPath, err)
	}

	for _, skill := range selectedSkills {
		var src string
		if skillsPath == "." {
			src = filepath.Join(tempRepoPath, skill.Path)
		} else {
			src = filepath.Join(tempRepoPath, skillsPath, skill.Path)
		}

		// Determinar el nombre de la carpeta destino (siempre plana)
		folderName := GetSkillFolderName(skill)
		dest := filepath.Join(localPath, folderName)

		// Limpiar destino para evitar anidamiento recursivo si ya existe
		os.RemoveAll(dest)

		// Asegurar que el padre existe (aunque localPath ya debería existir)
		os.MkdirAll(filepath.Dir(dest), 0755)

		if err := copyDir(src, dest); err != nil {
			return fmt.Errorf("error copying skill %s: %w", skill.Name, err)
		}
	}

	return nil
}

// copyDir copies a directory recursively from src to dst.
func copyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

// GetSkillFolderName devuelve el nombre de la carpeta local para un skill
func GetSkillFolderName(skill SkillInfo) string {
	name := filepath.Base(skill.Path)
	if name == "." || name == "/" || name == "" {
		// Si no hay path (URL directa), usamos el nombre del skill sanitizado
		name = strings.ToLower(skill.Name)
		name = strings.ReplaceAll(name, " ", "-")
		// Eliminar caracteres no permitidos básicos
		name = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
				return r
			}
			return -1
		}, name)
	}
	return name
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
