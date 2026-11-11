package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadCreatesDefaultConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if got := Load(); !reflect.DeepEqual(got, Default()) {
		t.Fatalf("Load() = %#v, want defaults", got)
	}
	data, err := os.ReadFile(configPath())
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("default config was not written")
	}
	if got := Load(); !reflect.DeepEqual(got, Default()) {
		t.Fatalf("saved config = %#v, want defaults", got)
	}
}

func TestLoadToolChoices(t *testing.T) {
	for _, tt := range []struct {
		name string
		json string
		want []string
	}{
		{"legacy", `{"theme":"nord"}`, Default().Tool},
		{"custom", `{"theme":"nord","tool":["zsh","/opt/my tool","nvim"]}`, []string{"zsh", "/opt/my tool", "nvim"}},
		{"empty", `{"tool":[]}`, []string{}},
		{"null", `{"tool":null}`, Default().Tool},
		{"invalid", `{"tool":"zsh"}`, Default().Tool},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			if err := os.MkdirAll(filepath.Dir(configPath()), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(configPath(), []byte(tt.json), 0644); err != nil {
				t.Fatal(err)
			}
			cfg := Load()
			if !reflect.DeepEqual(cfg.Tool, tt.want) {
				t.Fatalf("Tool = %#v, want %#v", cfg.Tool, tt.want)
			}
			if tt.name == "legacy" && cfg.Theme != "nord" {
				t.Fatalf("Theme = %q, want nord", cfg.Theme)
			}
		})
	}
}
