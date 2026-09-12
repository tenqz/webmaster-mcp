package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMountedSecretAndConflict(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(path, []byte("sentinel-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_SECRET", "")
	t.Setenv("TEST_SECRET_FILE", path)
	got, err := Secret("TEST_SECRET")
	if err != nil || got != "sentinel-secret" {
		t.Fatal("file secret not loaded")
	}
	t.Setenv("TEST_SECRET", "sentinel-secret")
	_, err = Secret("TEST_SECRET")
	if err == nil || strings.Contains(err.Error(), "sentinel-secret") {
		t.Fatal("unsafe conflict handling")
	}
}
