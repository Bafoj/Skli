package config

import (
	"os"
	"path/filepath"
	"testing"
)

func withTempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	old := os.Getenv("HOME")
	if err := os.Setenv("HOME", home); err != nil {
		t.Fatalf("set HOME: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Setenv("HOME", old)
	})
	return home
}

func TestLoadConfigWhenMissing(t *testing.T) {
	withTempHome(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if cfg.LocalPath != "" || len(cfg.Remotes) != 0 {
		t.Fatalf("expected zero config, got %+v", cfg)
	}
}

func TestSaveAndLoadConfigRoundtrip(t *testing.T) {
	home := withTempHome(t)

	want := Config{
		LocalPath: ".cursor/skills",
		Remotes:   []string{"https://github.com/acme/skills", "https://gitlab.com/acme/skills"},
	}
	if err := SaveConfig(want); err != nil {
		t.Fatalf("SaveConfig error: %v", err)
	}

	configPath := filepath.Join(home, ".skli", "config.toml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file at %s: %v", configPath, err)
	}

	got, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig after save: %v", err)
	}
	if got.LocalPath != want.LocalPath {
		t.Fatalf("LocalPath mismatch: got %q want %q", got.LocalPath, want.LocalPath)
	}
	if len(got.Remotes) != len(want.Remotes) {
		t.Fatalf("Remotes length mismatch: got %d want %d", len(got.Remotes), len(want.Remotes))
	}
	for i := range want.Remotes {
		if got.Remotes[i] != want.Remotes[i] {
			t.Fatalf("remote[%d] mismatch: got %q want %q", i, got.Remotes[i], want.Remotes[i])
		}
	}
}

func TestLoadConfigInvalidToml(t *testing.T) {
	home := withTempHome(t)
	configDir := filepath.Join(home, ".skli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte("not: [valid"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadConfig(); err == nil {
		t.Fatalf("expected parse error for invalid toml")
	}
}
