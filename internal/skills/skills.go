package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skli/internal/db"
	"skli/internal/skillmeta"
)

const DefaultRoot = "skills"

// IsSafeDeletePath valida que el path a eliminar sea un subdirectorio del root de skills.
func IsSafeDeletePath(pathToDelete, skillsRoot string) error {
	pathToDelete = filepath.Clean(strings.TrimSpace(pathToDelete))
	if pathToDelete == "" || pathToDelete == "." || pathToDelete == string(os.PathSeparator) {
		return fmt.Errorf("unsafe path to delete: %s", pathToDelete)
	}

	if strings.TrimSpace(skillsRoot) == "" {
		skillsRoot = DefaultRoot
	}
	skillsRoot = filepath.Clean(skillsRoot)

	absRoot, err := filepath.Abs(skillsRoot)
	if err != nil {
		return fmt.Errorf("could not resolve skills root: %w", err)
	}
	absTarget, err := filepath.Abs(pathToDelete)
	if err != nil {
		return fmt.Errorf("could not resolve path to delete: %w", err)
	}

	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil {
		return fmt.Errorf("could not validate path to delete: %w", err)
	}
	if rel == "." || rel == "" || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("path outside skills root: %s", pathToDelete)
	}
	return nil
}

// CollectAll combina skills del lockfile y skills locales no gestionados.
func CollectAll(skillsRoot string) ([]db.InstalledSkill, error) {
	lock, err := db.LoadLockFile()
	if err != nil {
		return nil, err
	}

	all := make([]db.InstalledSkill, 0, len(lock.Skills))
	all = append(all, lock.Skills...)

	localOnly, err := ScanLocalUnmanaged(lock.Skills, skillsRoot)
	if err != nil {
		return nil, err
	}
	all = append(all, localOnly...)

	return all, nil
}

// ScanLocalUnmanaged devuelve skills locales con SKILL.md que no están en el lockfile.
func ScanLocalUnmanaged(existing []db.InstalledSkill, skillsRoot string) ([]db.InstalledSkill, error) {
	if strings.TrimSpace(skillsRoot) == "" {
		skillsRoot = DefaultRoot
	}

	newSkills := make([]db.InstalledSkill, 0)
	existingMap := make(map[string]bool, len(existing))
	for _, sk := range existing {
		existingMap[sk.Path] = true
	}

	_ = filepath.Walk(skillsRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == "SKILL.md" {
			dir := filepath.Dir(path)
			rootClean := filepath.Clean(skillsRoot)
			if filepath.Clean(dir) == rootClean {
				return nil
			}
			relPath, _ := filepath.Rel(".", dir)
			if !existingMap[relPath] {
				newSkills = append(newSkills, db.InstalledSkill{
					Name:        filepath.Base(dir),
					Description: "Local skill (unmanaged)",
					Path:        relPath,
				})
			}
		}
		return nil
	})

	return newSkills, nil
}

// FindByName busca un skill por nombre exacto (case-insensitive) o por basename de path.
func FindByName(name, skillsRoot string) (db.InstalledSkill, error) {
	all, err := CollectAll(skillsRoot)
	if err != nil {
		return db.InstalledSkill{}, err
	}

	needle := strings.ToLower(strings.TrimSpace(name))
	if needle == "" {
		return db.InstalledSkill{}, fmt.Errorf("empty skill name")
	}

	matches := make([]db.InstalledSkill, 0)
	for _, sk := range all {
		if strings.ToLower(sk.Name) == needle || strings.ToLower(filepath.Base(sk.Path)) == needle {
			matches = append(matches, sk)
		}
	}

	if len(matches) == 0 {
		return db.InstalledSkill{}, fmt.Errorf("skill not found: %s", name)
	}
	if len(matches) > 1 {
		paths := make([]string, 0, len(matches))
		for _, m := range matches {
			paths = append(paths, m.Path)
		}
		return db.InstalledSkill{}, fmt.Errorf("ambiguous name '%s'. Matches: %s", name, strings.Join(paths, ", "))
	}

	return matches[0], nil
}

// Delete elimina el directorio del skill y su entrada en el lockfile.
func Delete(skill db.InstalledSkill) error {
	if err := IsSafeDeletePath(skill.Path, DefaultRoot); err != nil {
		return err
	}

	if err := os.RemoveAll(skill.Path); err != nil {
		return fmt.Errorf("error deleting '%s': %w", skill.Name, err)
	}
	if err := db.DeleteInstalledSkill(skill.Path); err != nil {
		return fmt.Errorf("error updating skli.lock: %w", err)
	}
	return nil
}

// DeleteByName resuelve un skill por nombre y lo elimina.
func DeleteByName(name, skillsRoot string) (db.InstalledSkill, error) {
	skill, err := FindByName(name, skillsRoot)
	if err != nil {
		return db.InstalledSkill{}, err
	}
	if err := Delete(skill); err != nil {
		return db.InstalledSkill{}, err
	}
	return skill, nil
}

// PrepareLocalForUpload valida la ruta y extrae metadata para subir un skill local.
func PrepareLocalForUpload(localSkillPath string) (db.InstalledSkill, error) {
	localSkillPath = strings.TrimSpace(localSkillPath)
	if localSkillPath == "" {
		return db.InstalledSkill{}, fmt.Errorf("local-skill-path is required")
	}

	absPath, err := filepath.Abs(localSkillPath)
	if err == nil {
		localSkillPath = absPath
	}

	info, err := os.Stat(localSkillPath)
	if err != nil || !info.IsDir() {
		return db.InstalledSkill{}, fmt.Errorf("invalid path: %s", localSkillPath)
	}

	meta, err := skillmeta.ParseDir(localSkillPath, 40)
	if err != nil {
		return db.InstalledSkill{}, fmt.Errorf("could not read SKILL.md: %w", err)
	}

	return db.InstalledSkill{
		Name:        meta.Name,
		Description: meta.Description,
		Path:        localSkillPath,
	}, nil
}
