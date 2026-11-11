package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/mdipanjan/hive/internal/styles"
)

type Config struct {
	Theme string   `json:"theme"`
	Tool  []string `json:"tool"`
}

func Default() Config {
	return Config{
		Theme: styles.TokyoNight.Key,
		Tool:  []string{"pi", "claude", "codex", "zsh"},
	}
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "hive", "config.json")
}

func Load() Config {
	cfg := Default()
	data, err := os.ReadFile(configPath())
	if os.IsNotExist(err) {
		_ = Save(cfg)
		return cfg
	}
	if err != nil || json.Unmarshal(data, &cfg) != nil {
		return Default()
	}
	if cfg.Tool == nil {
		cfg.Tool = Default().Tool
	}
	return cfg
}

func Save(cfg Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
