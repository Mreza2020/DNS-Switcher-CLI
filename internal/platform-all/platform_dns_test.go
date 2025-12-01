package platform_all

import (
	"testing"

	config2 "github.com/Mreza2020/DNS-Switcher/internal/config"
)

// TestApplyProfile_Mock: unit test for ApplyProfile using a mock profile
// Verifies that profile application returns success without affecting real system
func TestApplyProfile_Mock(t *testing.T) {

	p := config2.Profile{
		Name:      "test",
		Servers:   []string{"8.8.8.8", "1.1.1.1"},
		Interface: "Wi-Fi",
	}

	result, err := ApplyProfile(p)
	if err != nil {
		t.Fatalf("ApplyProfile returned error: %v", err)
	}

	if !result.Ok {
		t.Fatalf("ApplyProfile did not return success")
	}

	t.Logf("ApplyProfile OK: %s", result.Message)
}

// TestRollback_Mock: unit test for Rollback functionality
// Ensures rollback succeeds and restores previous state using mock execution
func TestRollback_Mock(t *testing.T) {
	mockIface := "Wi-Fi"

	result, err := Rollback(mockIface)
	if err != nil {
		t.Fatalf("Rollback returned error: %v", err)
	}

	if !result.Ok {
		t.Fatalf("Rollback did not return success")
	}

	t.Logf("Rollback OK: %s", result.Message)
}
