package app

import (
	"os"
	"path/filepath"
	"testing"

	"skli/internal/config"
	"skli/internal/db"
)

func TestListSkillsIncludesManagedAndLocal(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	managedPath := filepath.Join("skills", "managed-skill")
	localPath := filepath.Join("skills", "local-skill")
	if err := os.MkdirAll(managedPath, 0755); err != nil {
		t.Fatalf("mkdir managed: %v", err)
	}
	if err := os.MkdirAll(localPath, 0755); err != nil {
		t.Fatalf("mkdir local: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localPath, "SKILL.md"), []byte("# local\n"), 0644); err != nil {
		t.Fatalf("write local SKILL.md: %v", err)
	}

	if err := db.SaveInstalledSkill(db.InstalledSkill{
		Name:        "managed-skill",
		Description: "managed",
		Path:        managedPath,
	}); err != nil {
		t.Fatalf("save installed: %v", err)
	}

	svc := NewService(config.Config{})
	listed, err := svc.ListSkills()
	if err != nil {
		t.Fatalf("ListSkills: %v", err)
	}

	if len(listed) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(listed))
	}

	var gotManaged, gotLocal bool
	for _, item := range listed {
		if item.Skill.Name == "managed-skill" {
			gotManaged = item.Managed && !item.LocalOnly
		}
		if item.Skill.Name == "local-skill" {
			gotLocal = item.LocalOnly && !item.Managed
		}
	}

	if !gotManaged {
		t.Fatalf("managed skill not found or incorrectly labeled")
	}
	if !gotLocal {
		t.Fatalf("local skill not found or incorrectly labeled")
	}
}
