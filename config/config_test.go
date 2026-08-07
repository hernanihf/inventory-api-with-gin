package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadConfig_Ok(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{"database":{"host":"localhost","port":5432,"username":"u","password":"p","dbname":"d"}}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("could not create temporary file: %v", err)
	}

	cfg, err := ReadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Database.Host != "localhost" || cfg.Database.Port != 5432 {
		t.Errorf("bad config parsing: %+v", cfg.Database)
	}
}

func TestReadConfig_FileNotFound(t *testing.T) {
	if _, err := ReadConfig("not-exists.json"); err == nil {
		t.Errorf("Not existent file error was expected")
	}
}

func TestReadConfig_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0644); err != nil {
		t.Fatalf("could not create temporary file: %v", err)
	}

	if _, err := ReadConfig(path); err == nil {
		t.Errorf("invalid json error was expected")
	}
}
