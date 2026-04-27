package config

import "testing"

func TestLoadConfig(t *testing.T) {
	t.Run("invalid file name", func(t *testing.T) {
		if _, err := LoadConfig("invalid.yaml"); err == nil {
			t.Error("expected error: file has invalid name")
		}
	})
	t.Run("type mismatch in config file", func(t *testing.T) {
		if _, err := LoadConfig("testdata/unmarshal_error_config.yaml"); err == nil {
			t.Error("expected error: type mismatch")
		}
	})
	t.Run("invalid parametrs", func(t *testing.T) {
		if _, err := LoadConfig("testdata/invalidparam_error_config.yaml"); err == nil {
			t.Error("expected error: gorutines > 1000 and IssueInOneRequest < 50")
		}
	})
	t.Run("valid parametrs", func(t *testing.T) {
		if _, err := LoadConfig("testdata/validconfig.yaml"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
