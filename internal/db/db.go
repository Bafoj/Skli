package db

import (
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
	CommitHash  string    `toml:"commit_hash"` // Hash del commit cuando se instaló
	InstalledAt time.Time `toml:"installed_at"`
	UpdatedAt   time.Time `toml:"updated_at"`
}

// LockFile representa la estructura del archivo skli.lock
type LockFile struct {
	LastUpdated time.Time        `toml:"last_updated"`
	Skills      []InstalledSkill `toml:"skills"`
}

const lockFileName = "skli.lock"

// LoadLockFile carga el archivo skli.lock
func LoadLockFile() (*LockFile, error) {
	if _, err := os.Stat(lockFileName); os.IsNotExist(err) {
		return &LockFile{Skills: []InstalledSkill{}}, nil
	}

	var lock LockFile
	if _, err := toml.DecodeFile(lockFileName, &lock); err != nil {
		return nil, err
	}

	return &lock, nil
}

// SaveLockFile guarda el archivo skli.lock
func SaveLockFile(lock *LockFile) error {
	lock.LastUpdated = time.Now()
	f, err := os.Create(lockFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(lock)
}

// SaveInstalledSkill guarda o actualiza un skill instalado en el lock file
func SaveInstalledSkill(skill InstalledSkill) error {
	lock, err := LoadLockFile()
	if err != nil {
		return err
	}

	// Buscar si ya existe para actualizarlo
	found := false
	for i, s := range lock.Skills {
		if s.Path == skill.Path {
			skill.InstalledAt = s.InstalledAt
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

// GetSkillsByRepo agrupa los skills por repo de origen
func GetSkillsByRepo() (map[string][]InstalledSkill, error) {
	lock, err := LoadLockFile()
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]InstalledSkill)
	for _, s := range lock.Skills {
		grouped[s.RemoteRepo] = append(grouped[s.RemoteRepo], s)
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
