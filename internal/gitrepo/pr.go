package gitrepo

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"skli/internal/db"
)

// CheckGhInstalled verifica si la herramienta CLI 'gh' está instalada
func CheckGhInstalled() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

func checkGlabInstalled() bool {
	_, err := exec.LookPath("glab")
	return err == nil
}

// CloneForPush clona el repositorio completo para realizar cambios y push
func CloneForPush(remoteURL string) (string, error) {
	tempDir, err := os.MkdirTemp("", "skli-pr-*")
	if err != nil {
		return "", fmt.Errorf("error creando dir temporal: %w", err)
	}

	if err := runGit(tempDir, "clone", remoteURL, "."); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("error clonando repo: %w", err)
	}

	return tempDir, nil
}

// PrepareSkillBranch crea una rama para el skill y devuelve el nombre de la rama
func PrepareSkillBranch(repoDir, skillName string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	sanitizedName := strings.ReplaceAll(strings.ToLower(skillName), " ", "-")
	sanitizedName = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, sanitizedName)

	branchName := fmt.Sprintf("feat/update-%s-%s", sanitizedName, timestamp)

	if err := runGit(repoDir, "checkout", "-b", branchName); err != nil {
		return "", fmt.Errorf("error creando rama %s: %w", branchName, err)
	}

	return branchName, nil
}

// FindSkillInRepo busca la ruta relativa del skill dentro del repositorio clonado.
func FindSkillInRepo(repoDir, skillName string) (string, error) {
	var foundPath string
	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == "SKILL.md" {
			skill, err := parseSkillFile(path, repoDir)
			if err == nil && skill.Name == skillName {
				foundPath = skill.Path
				return filepath.SkipAll
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	if foundPath == "" {
		return "", fmt.Errorf("skill '%s' no encontrado en el repositorio remoto", skillName)
	}
	return foundPath, nil
}

// CopySkillFiles copia los contenidos del skill local al directorio del repo clonado.
func CopySkillFiles(repoDir, localSkillPath, repoSkillPath string) error {
	destDir := filepath.Join(repoDir, repoSkillPath)

	os.RemoveAll(destDir)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio destino: %w", err)
	}

	cmd := exec.Command("cp", "-r", localSkillPath+"/", destDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error copiando archivos: %w : %s", err, string(output))
	}

	return nil
}

type gitProvider int

const (
	providerUnknown gitProvider = iota
	providerGitHub
	providerGitLab
	providerBitbucket
)

func detectProvider(repoURL string) gitProvider {
	u := strings.ToLower(repoURL)
	switch {
	case strings.Contains(u, "github.com"):
		return providerGitHub
	case strings.Contains(u, "gitlab.com"):
		return providerGitLab
	case strings.Contains(u, "bitbucket.org"):
		return providerBitbucket
	default:
		return providerUnknown
	}
}

func normalizeRepoWebURL(remoteURL string) string {
	base := ParseGitURL(remoteURL).BaseURL
	base = strings.TrimSpace(base)

	if strings.HasPrefix(base, "git@") {
		parts := strings.SplitN(strings.TrimPrefix(base, "git@"), ":", 2)
		if len(parts) == 2 {
			return "https://" + parts[0] + "/" + strings.TrimSuffix(strings.Trim(parts[1], "/"), ".git")
		}
	}

	if strings.HasPrefix(base, "ssh://") {
		if u, err := url.Parse(base); err == nil {
			return "https://" + u.Host + strings.TrimSuffix(u.Path, ".git")
		}
	}

	if u, err := url.Parse(base); err == nil {
		scheme := u.Scheme
		if scheme == "" {
			scheme = "https"
		}
		path := strings.TrimSuffix(strings.Trim(u.Path, "/"), ".git")
		if path != "" {
			path = "/" + path
		}
		return scheme + "://" + u.Host + path
	}

	return strings.TrimSuffix(strings.Trim(base, "/"), ".git")
}

func getDefaultBranch(repoDir string) string {
	cmd := exec.Command("git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	cmd.Dir = repoDir
	output, err := cmd.Output()
	if err != nil {
		return "main"
	}
	ref := strings.TrimSpace(string(output))
	if strings.HasPrefix(ref, "origin/") {
		return strings.TrimPrefix(ref, "origin/")
	}
	if ref != "" {
		return ref
	}
	return "main"
}

func buildPRURL(provider gitProvider, repoURL, branchName, targetBranch, title string) string {
	repoWeb := normalizeRepoWebURL(repoURL)

	switch provider {
	case providerGitHub:
		return fmt.Sprintf("%s/compare/%s...%s?expand=1", repoWeb, targetBranch, branchName)
	case providerGitLab:
		q := url.Values{}
		q.Set("merge_request[source_branch]", branchName)
		q.Set("merge_request[target_branch]", targetBranch)
		q.Set("merge_request[title]", title)
		return fmt.Sprintf("%s/-/merge_requests/new?%s", repoWeb, q.Encode())
	case providerBitbucket:
		q := url.Values{}
		q.Set("source", branchName)
		q.Set("dest", targetBranch)
		return fmt.Sprintf("%s/pull-requests/new?%s", repoWeb, q.Encode())
	default:
		return repoWeb
	}
}

// PushAndCreatePR hace commit, push y crea PR/MR cuando es posible.
func PushAndCreatePR(repoDir, remoteURL, branchName, skillName, description string) (string, error) {
	if err := runGit(repoDir, "add", "."); err != nil {
		return "", fmt.Errorf("git add fallo: %w", err)
	}

	if err := runGit(repoDir, "diff", "--staged", "--quiet"); err == nil {
		return "", fmt.Errorf("no hay cambios para subir (el contenido local es identico al remoto)")
	}

	msg := fmt.Sprintf("feat(%s): update skill content", skillName)
	if err := runGit(repoDir, "commit", "-m", msg); err != nil {
		return "", fmt.Errorf("git commit fallo: %w", err)
	}

	if err := runGit(repoDir, "push", "origin", branchName); err != nil {
		return "", fmt.Errorf("git push fallo: %w", err)
	}

	title := fmt.Sprintf("Update skill: %s", skillName)
	body := fmt.Sprintf("This PR updates the skill '%s'.\n\nAutomatically generated by skli.\n\n%s", skillName, description)
	targetBranch := getDefaultBranch(repoDir)
	provider := detectProvider(remoteURL)
	fallbackURL := buildPRURL(provider, remoteURL, branchName, targetBranch, title)

	switch provider {
	case providerGitHub:
		if CheckGhInstalled() {
			cmd := exec.Command("gh", "pr", "create", "--title", title, "--body", body, "--head", branchName, "--base", targetBranch)
			cmd.Dir = repoDir
			output, err := cmd.CombinedOutput()
			if err == nil {
				return strings.TrimSpace(string(output)), nil
			}
		}
	case providerGitLab:
		if checkGlabInstalled() {
			cmd := exec.Command("glab", "mr", "create", "--title", title, "--description", body, "--source-branch", branchName, "--target-branch", targetBranch, "--yes")
			cmd.Dir = repoDir
			output, err := cmd.CombinedOutput()
			if err == nil {
				return strings.TrimSpace(string(output)), nil
			}
		}
	}

	return fallbackURL, nil
}

// UploadSkill sube un skill local a un repositorio remoto y crea una PR/MR (o devuelve URL fallback).
func UploadSkill(skill db.InstalledSkill, targetRemoteURL string) (string, error) {
	tempDir, err := CloneForPush(targetRemoteURL)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	branchName, err := PrepareSkillBranch(tempDir, skill.Name)
	if err != nil {
		return "", err
	}

	repoSkillPath := skill.RemotePath
	foundPath, err := FindSkillInRepo(tempDir, skill.Name)
	if err == nil {
		repoSkillPath = foundPath
	} else {
		skillFolderName := filepath.Base(skill.Path)
		repoSkillPath = filepath.Join("skills", skillFolderName)
	}

	if err := CopySkillFiles(tempDir, skill.Path, repoSkillPath); err != nil {
		return "", err
	}

	return PushAndCreatePR(tempDir, targetRemoteURL, branchName, skill.Name, skill.Description)
}
