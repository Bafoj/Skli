package skills

import (
	"path/filepath"
	"testing"
)

func TestIsSafeDeletePath(t *testing.T) {
	root := t.TempDir()
	skillsRoot := filepath.Join(root, "skills")

	t.Run("allows nested path inside skills root", func(t *testing.T) {
		target := filepath.Join(skillsRoot, "my-skill")
		if err := IsSafeDeletePath(target, skillsRoot); err != nil {
			t.Fatalf("expected safe path, got error: %v", err)
		}
	})

	t.Run("blocks deleting skills root", func(t *testing.T) {
		if err := IsSafeDeletePath(skillsRoot, skillsRoot); err == nil {
			t.Fatalf("expected error deleting skills root")
		}
	})

	t.Run("blocks path outside skills root", func(t *testing.T) {
		target := filepath.Join(root, "other", "skill")
		if err := IsSafeDeletePath(target, skillsRoot); err == nil {
			t.Fatalf("expected error for outside path")
		}
	})
}
