package gitrepo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGitURLGitHubTree(t *testing.T) {
	info := ParseGitURL("https://github.com/org/repo/tree/main/skills/experimental")
	if info.BaseURL != "https://github.com/org/repo" {
		t.Fatalf("unexpected BaseURL: %s", info.BaseURL)
	}
	if info.Branch != "main" {
		t.Fatalf("unexpected Branch: %s", info.Branch)
	}
	if info.SubPath != "skills/experimental" {
		t.Fatalf("unexpected SubPath: %s", info.SubPath)
	}
}

func TestParseGitURLBitbucketSrc(t *testing.T) {
	info := ParseGitURL("https://bitbucket.org/org/repo/src/master/skills")
	if info.BaseURL != "https://bitbucket.org/org/repo" {
		t.Fatalf("unexpected BaseURL: %s", info.BaseURL)
	}
	if info.Branch != "master" {
		t.Fatalf("unexpected Branch: %s", info.Branch)
	}
	if info.SubPath != "skills" {
		t.Fatalf("unexpected SubPath: %s", info.SubPath)
	}
}

func TestParseGitURLSSHUntouched(t *testing.T) {
	url := "git@github.com:org/repo.git"
	info := ParseGitURL(url)
	if info.BaseURL != url || info.Branch != "HEAD" || info.SubPath != "" {
		t.Fatalf("unexpected info: %+v", info)
	}
}

func TestGetSkillFolderName(t *testing.T) {
	if got := GetSkillFolderName(SkillInfo{Path: "nested/my-skill"}); got != "my-skill" {
		t.Fatalf("expected basename folder, got %q", got)
	}

	got := GetSkillFolderName(SkillInfo{Name: "My Skill! 2", Path: ""})
	if got != "my-skill-2" {
		t.Fatalf("expected sanitized name, got %q", got)
	}
}

func TestParseSkillFileAndFindSkills(t *testing.T) {
	tmp := t.TempDir()
	base := filepath.Join(tmp, "skills")
	skillDir := filepath.Join(base, "alpha")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	skillFile := filepath.Join(skillDir, "SKILL.md")
	content := "---\nname: Alpha\ndescription: Alpha description\n---\n# body\n"
	if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parsed, err := parseSkillFile(skillFile, base)
	if err != nil {
		t.Fatalf("parseSkillFile error: %v", err)
	}
	if parsed.Name != "Alpha" || parsed.Description != "Alpha description" || parsed.Path != "alpha" {
		t.Fatalf("unexpected parsed skill: %+v", parsed)
	}

	found, err := findSkills(base)
	if err != nil {
		t.Fatalf("findSkills error: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(found))
	}
	if found[0].Name != "Alpha" || found[0].Path != "alpha" {
		t.Fatalf("unexpected found skill: %+v", found[0])
	}
}

func TestInstallSkills(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	local := filepath.Join(tmp, "local")
	skillSrc := filepath.Join(repo, "skills", "sample")
	if err := os.MkdirAll(skillSrc, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillSrc, "SKILL.md"), []byte("---\nname: Sample\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}

	selected := []SkillInfo{{Name: "Sample", Path: "sample"}}
	if err := InstallSkills(repo, "skills", local, selected); err != nil {
		t.Fatalf("InstallSkills error: %v", err)
	}

	installed := filepath.Join(local, "sample", "SKILL.md")
	if _, err := os.Stat(installed); err != nil {
		t.Fatalf("expected installed skill file: %v", err)
	}
}
