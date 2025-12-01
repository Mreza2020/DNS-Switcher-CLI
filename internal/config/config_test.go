package config

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func setupTestConfig(t *testing.T) string {
	tmpFile := filepath.Join(t.TempDir(), "profiles.yaml")
	viper.Reset()
	viper.SetConfigFile(tmpFile)
	viper.SetConfigType("yaml")

	viper.Set("profiles", map[string]interface{}{})

	if err := viper.WriteConfigAs(tmpFile); err != nil {
		t.Fatalf("cannot write temp config: %v", err)
	}
	Path = tmpFile
	return tmpFile
}

// TestAddLoadFindProfile: verifies adding, loading, and finding DNS profiles
func TestAddLoadFindProfile(t *testing.T) {
	setupTestConfig(t)

	err := AddProfile("test", []string{"8.8.8.8"})
	if err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	profiles := LoadProfilesDns()
	if len(profiles) == 0 {
		t.Fatal("No profiles loaded")
	}

	p, ok := FindProfile(profiles, "test")
	if !ok {
		t.Fatal("Profile not found after adding")
	}

	if len(p.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(p.Servers))
	}
}

// TestDeleteProfile: verifies deleting a profile
func TestDeleteProfile(t *testing.T) {
	setupTestConfig(t)

	if err := AddProfile("deleteme", []string{"9.9.9.9"}); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	profiles := LoadProfilesDns()
	if _, ok := FindProfile(profiles, "deleteme"); !ok {
		t.Fatal("Profile not found before deletion")
	}

	if err := DeleteProfile("deleteme"); err != nil {
		t.Fatalf("DeleteProfile failed: %v", err)
	}

	profiles = LoadProfilesDns()
	if _, ok := FindProfile(profiles, "deleteme"); ok {
		t.Fatal("Profile still found after deletion")
	}
}
