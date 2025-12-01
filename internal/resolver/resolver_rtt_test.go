package resolver

import (
	"testing"
	"time"

	config2 "github.com/Mreza2020/DNS-Switcher/internal/config"
)

// TestMeasureRTT: measures round-trip time (RTT) to given servers
func TestMeasureRTT(t *testing.T) {
	r := MeasureRTT("8.8.8.8", "google.com", 2*time.Second)

	if r.Error != nil {
		t.Logf("RTT test warning (no failure): %v", r.Error)
	} else {
		t.Logf("RTT OK: %v", r.RTT)
	}
}

// TestFindFastestProfile: identifies the fastest DNS profile from a list
func TestFindFastestProfile(t *testing.T) {
	profiles := []config2.Profile{
		{Name: "google", Servers: []string{"8.8.8.8"}},
		{Name: "cloudflare", Servers: []string{"1.1.1.1"}},
	}

	fastest := FindFastestProfile(profiles)

	if fastest == nil {
		t.Fatal("No fastest profile returned")
	}

	t.Logf("Fastest profile: %s", fastest.Name)
}
