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
		t.Fatalf("no se pudo crear el archivo temporal: %v", err)
	}

	cfg, err := ReadConfig(path)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if cfg.Database.Host != "localhost" || cfg.Database.Port != 5432 {
		t.Errorf("config mal parseada: %+v", cfg.Database)
	}
}

func TestReadConfig_FileNotFound(t *testing.T) {
	if _, err := ReadConfig("no-existe.json"); err == nil {
		t.Errorf("esperaba error por archivo inexistente")
	}
}

func TestReadConfig_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0644); err != nil {
		t.Fatalf("no se pudo crear el archivo temporal: %v", err)
	}

	if _, err := ReadConfig(path); err == nil {
		t.Errorf("esperaba error por JSON inválido")
	}
}
