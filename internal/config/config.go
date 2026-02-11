package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	LocalPath string   `toml:"local_path"`
	Remotes   []string `toml:"remotes"`
}

func GetConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".skli")
}

func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.toml")
}

func LoadConfig() (Config, error) {
	var cfg Config
	path := GetConfigPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding config: %w", err)
	}

	return cfg, nil
}

func SaveConfig(cfg Config) error {
	dir := GetConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	path := GetConfigPath()
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating config file: %w", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("error encoding config: %w", err)
	}

	return nil
}
