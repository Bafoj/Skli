package db

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// InstalledSkill representa un skill instalado con su origen
type InstalledSkill struct {
	Name        string    `toml:"name"`
	Description string    `toml:"description"`
	Path        string    `toml:"path"`        // Ruta local relativa (ej: "cloudflare-deploy")
	RemoteRepo  string    `toml:"remote_repo"` // URL del repo de origen
	RemoteRoot  string    `toml:"remote_root"` // Directorio base dentro del repo (ej: "skills")
	RemotePath  string    `toml:"remote_path"` // Ruta relativa al RemoteRoot (ej: ".curated/cloudflare-deploy")
	CommitHash  string    `toml:"commit_hash"` // Hash del commit cuando se instaló (deprecated)
	TreeHash    string    `toml:"tree_hash"`   // Hash del árbol (carpeta) cuando se instaló
	InstalledAt time.Time `toml:"installed_at"`
	UpdatedAt   time.Time `toml:"updated_at"`
}

// LockFile representa la estructura del archivo skli.lock
type LockFile struct {
	LastUpdated time.Time        `toml:"last_updated"`
	Skills      []InstalledSkill `toml:"skills"`
}

const lockFileName = "skli.lock"

func getLockFilePath() string {
	return lockFileName
}

// LoadLockFile lee el archivo skli.lock
func LoadLockFile() (*LockFile, error) {
	var lock LockFile
	path := getLockFilePath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &lock, nil
	}

	if _, err := toml.DecodeFile(path, &lock); err != nil {
		return nil, fmt.Errorf("error reading lock file: %w", err)
	}

	return &lock, nil
}

// SaveLockFile guarda el archivo skli.lock
func SaveLockFile(lock *LockFile) error {
	path := getLockFilePath()
	lock.LastUpdated = time.Now()

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating lock file: %w", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(lock); err != nil {
		return fmt.Errorf("error writing lock file: %w", err)
	}

	return nil
}

// SaveInstalledSkill añade o actualiza un skill en el lock file
func SaveInstalledSkill(skill InstalledSkill) error {
	lock, err := LoadLockFile()
	if err != nil {
		return err
	}

	found := false
	for i, s := range lock.Skills {
		// Asumimos que la combinación de repo + path es única
		if s.RemoteRepo == skill.RemoteRepo && s.RemotePath == skill.RemotePath {
			skill.InstalledAt = s.InstalledAt // Mantener fecha original
			skill.UpdatedAt = time.Now()
			lock.Skills[i] = skill
			found = true
			break
		}
	}

	if !found {
		skill.InstalledAt = time.Now()
		skill.UpdatedAt = time.Now()
		lock.Skills = append(lock.Skills, skill)
	}

	return SaveLockFile(lock)
}

// RemoveInstalledSkill elimina un skill del lock file usando su ruta local
func RemoveInstalledSkill(localPath string) error {
	lock, err := LoadLockFile()
	if err != nil {
		return err
	}

	var newSkills []InstalledSkill
	for _, s := range lock.Skills {
		if s.Path != localPath {
			newSkills = append(newSkills, s)
		}
	}

	lock.Skills = newSkills
	return SaveLockFile(lock)
}

// GetSkillsByRepo agrupa los skills instalados por su repositorio de origen
func GetSkillsByRepo() (map[string][]InstalledSkill, error) {
	lock, err := LoadLockFile()
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]InstalledSkill)
	for _, skill := range lock.Skills {
		if skill.RemoteRepo != "" {
			grouped[skill.RemoteRepo] = append(grouped[skill.RemoteRepo], skill)
		}
	}

	return grouped, nil
}

// DeleteInstalledSkill elimina un skill del lock file por su nombre (o ruta relativa)
func DeleteInstalledSkill(path string) error {
	lock, err := LoadLockFile()
	if err != nil {
		return err
	}

	newSkills := make([]InstalledSkill, 0)
	found := false
	for _, s := range lock.Skills {
		if s.Path == path {
			found = true
			continue
		}
		newSkills = append(newSkills, s)
	}

	if !found {
		return nil // No estaba, nada que hacer
	}

	lock.Skills = newSkills
	return SaveLockFile(lock)
}
