package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func withTempWorkdir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prev)
	})
	return dir
}

func TestLoadLockFileWhenMissing(t *testing.T) {
	withTempWorkdir(t)

	lock, err := LoadLockFile()
	if err != nil {
		t.Fatalf("LoadLockFile error: %v", err)
	}
	if len(lock.Skills) != 0 {
		t.Fatalf("expected empty skills, got %d", len(lock.Skills))
	}
}

func TestSaveInstalledSkillInsertAndUpdate(t *testing.T) {
	withTempWorkdir(t)

	first := InstalledSkill{
		Name:       "skill-a",
		Path:       filepath.Join("skills", "skill-a"),
		RemoteRepo: "https://github.com/acme/repo",
	}
	if err := SaveInstalledSkill(first); err != nil {
		t.Fatalf("SaveInstalledSkill insert: %v", err)
	}

	lock, err := LoadLockFile()
	if err != nil {
		t.Fatalf("LoadLockFile after insert: %v", err)
	}
	if len(lock.Skills) != 1 {
		t.Fatalf("expected 1 skill after insert, got %d", len(lock.Skills))
	}
	inserted := lock.Skills[0]
	if inserted.InstalledAt.IsZero() || inserted.UpdatedAt.IsZero() {
		t.Fatalf("timestamps should be set on insert")
	}

	time.Sleep(10 * time.Millisecond)
	updated := first
	updated.Description = "new-description"
	if err := SaveInstalledSkill(updated); err != nil {
		t.Fatalf("SaveInstalledSkill update: %v", err)
	}

	lock, err = LoadLockFile()
	if err != nil {
		t.Fatalf("LoadLockFile after update: %v", err)
	}
	if len(lock.Skills) != 1 {
		t.Fatalf("expected 1 skill after update, got %d", len(lock.Skills))
	}
	got := lock.Skills[0]
	if got.Description != "new-description" {
		t.Fatalf("expected description updated, got %q", got.Description)
	}
	if !got.InstalledAt.Equal(inserted.InstalledAt) {
		t.Fatalf("InstalledAt must stay stable on update")
	}
	if !got.UpdatedAt.After(inserted.UpdatedAt) {
		t.Fatalf("UpdatedAt must advance on update")
	}
}

func TestDeleteInstalledSkillAndGroupByRepo(t *testing.T) {
	withTempWorkdir(t)

	s1 := InstalledSkill{Name: "a", Path: "skills/a", RemoteRepo: "repo-1"}
	s2 := InstalledSkill{Name: "b", Path: "skills/b", RemoteRepo: "repo-1"}
	s3 := InstalledSkill{Name: "c", Path: "skills/c", RemoteRepo: "repo-2"}
	if err := SaveInstalledSkill(s1); err != nil {
		t.Fatal(err)
	}
	if err := SaveInstalledSkill(s2); err != nil {
		t.Fatal(err)
	}
	if err := SaveInstalledSkill(s3); err != nil {
		t.Fatal(err)
	}

	grouped, err := GetSkillsByRepo()
	if err != nil {
		t.Fatalf("GetSkillsByRepo error: %v", err)
	}
	if len(grouped["repo-1"]) != 2 || len(grouped["repo-2"]) != 1 {
		t.Fatalf("unexpected grouping: %#v", grouped)
	}

	if err := DeleteInstalledSkill("skills/b"); err != nil {
		t.Fatalf("DeleteInstalledSkill error: %v", err)
	}

	lock, err := LoadLockFile()
	if err != nil {
		t.Fatal(err)
	}
	if len(lock.Skills) != 2 {
		t.Fatalf("expected 2 skills after delete, got %d", len(lock.Skills))
	}
	for _, sk := range lock.Skills {
		if sk.Path == "skills/b" {
			t.Fatalf("skill skills/b should be deleted")
		}
	}
}
