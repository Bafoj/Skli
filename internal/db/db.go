package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// InstalledSkill representa un skill instalado con su origen
type InstalledSkill struct {
	ID          int
	Name        string
	Description string
	Path        string // Ruta local relativa (ej: "cloudflare-deploy")
	RemoteRepo  string // URL del repo de origen
	RemotePath  string // Ruta dentro del repo (ej: ".curated/cloudflare-deploy")
	InstalledAt time.Time
	UpdatedAt   time.Time
}

var dbPath string

func init() {
	home, _ := os.UserHomeDir()
	dbPath = filepath.Join(home, ".skli", "skills.db")
}

// InitDB inicializa la base de datos y crea las tablas si no existen
func InitDB() (*sql.DB, error) {
	// Crear directorio si no existe
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Crear tabla de skills instalados
	schema := `
	CREATE TABLE IF NOT EXISTS installed_skills (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		path TEXT NOT NULL UNIQUE,
		remote_repo TEXT NOT NULL,
		remote_path TEXT NOT NULL,
		installed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_remote_repo ON installed_skills(remote_repo);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// SaveInstalledSkill guarda o actualiza un skill instalado
func SaveInstalledSkill(db *sql.DB, skill InstalledSkill) error {
	query := `
	INSERT INTO installed_skills (name, description, path, remote_repo, remote_path)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(path) DO UPDATE SET
		name = excluded.name,
		description = excluded.description,
		remote_repo = excluded.remote_repo,
		remote_path = excluded.remote_path,
		updated_at = CURRENT_TIMESTAMP
	`
	_, err := db.Exec(query, skill.Name, skill.Description, skill.Path, skill.RemoteRepo, skill.RemotePath)
	return err
}

// GetAllSkills obtiene todos los skills instalados
func GetAllSkills(db *sql.DB) ([]InstalledSkill, error) {
	rows, err := db.Query(`
		SELECT id, name, description, path, remote_repo, remote_path, installed_at, updated_at
		FROM installed_skills
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []InstalledSkill
	for rows.Next() {
		var s InstalledSkill
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Path, &s.RemoteRepo, &s.RemotePath, &s.InstalledAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}

// GetSkillsByRepo agrupa los skills por repo de origen
func GetSkillsByRepo(db *sql.DB) (map[string][]InstalledSkill, error) {
	skills, err := GetAllSkills(db)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]InstalledSkill)
	for _, s := range skills {
		grouped[s.RemoteRepo] = append(grouped[s.RemoteRepo], s)
	}
	return grouped, nil
}

// DeleteSkillByPath elimina un skill por su ruta local
func DeleteSkillByPath(db *sql.DB, path string) error {
	_, err := db.Exec("DELETE FROM installed_skills WHERE path = ?", path)
	return err
}
