package config

import (
	"path/filepath"
	"testing"
)

func testdataPath(name string) string {
	return filepath.Join("..", "..", "testdata", "config", name)
}

func TestLoadValidConfig(t *testing.T) {
	cfg, err := Load(testdataPath("valid.yaml"))
	if err != nil {
		t.Fatalf("Load valid config: %v", err)
	}

	if got := cfg.Logging.Level; got != "debug" {
		t.Errorf("Logging.Level = %q, want %q", got, "debug")
	}

	if got := len(cfg.Scanner.CustomPatterns); got != 1 {
		t.Fatalf("len(CustomPatterns) = %d, want 1", got)
	}

	cp := cfg.Scanner.CustomPatterns[0]
	if cp.Name != "Employee ID" {
		t.Errorf("CustomPatterns[0].Name = %q, want %q", cp.Name, "Employee ID")
	}
	if cp.Type != "EMPLOYEE_ID" {
		t.Errorf("CustomPatterns[0].Type = %q, want %q", cp.Type, "EMPLOYEE_ID")
	}
	if cp.Score != 0.9 {
		t.Errorf("CustomPatterns[0].Score = %v, want 0.9", cp.Score)
	}

	if got := len(cfg.Scanner.Allowlist); got != 2 {
		t.Fatalf("len(Allowlist) = %d, want 2", got)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(testdataPath("does_not_exist.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadInvalidRegex(t *testing.T) {
	_, err := Load(testdataPath("invalid_regex.yaml"))
	if err == nil {
		t.Fatal("expected error for invalid regex, got nil")
	}
}

func TestLoadInvalidAllowlist(t *testing.T) {
	_, err := Load(testdataPath("invalid_allowlist.yaml"))
	if err == nil {
		t.Fatal("expected error for invalid allowlist regex, got nil")
	}
}

func TestLoadInvalidLogLevel(t *testing.T) {
	_, err := Load(testdataPath("invalid_level.yaml"))
	if err == nil {
		t.Fatal("expected error for invalid log level, got nil")
	}
}

func TestLoadEmptyConfigMergesDefaults(t *testing.T) {
	cfg, err := Load(testdataPath("empty.yaml"))
	if err != nil {
		t.Fatalf("Load empty config: %v", err)
	}

	def := DefaultConfig()
	if cfg.Logging.Level != def.Logging.Level {
		t.Errorf("empty config Logging.Level = %q, want default %q", cfg.Logging.Level, def.Logging.Level)
	}
}

func TestDefaultConfigIsValid(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("DefaultConfig().Validate() = %v", err)
	}
}

func TestValidateCatchesInvalidRegex(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Scanner.CustomPatterns = []CustomPattern{
		{Name: "bad", Type: "BAD", Pattern: "[invalid", Score: 0.5},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected Validate to catch invalid regex")
	}
}

func TestValidateCatchesInvalidLogLevel(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Logging.Level = "trace"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected Validate to catch invalid log level")
	}
}
