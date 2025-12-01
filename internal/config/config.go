package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

var Path string

// init: initializes the configuration path for DNS profiles
// Reads from environment variable or defaults to "profiles.yaml"
func init() {
	Path = os.Getenv("DNS_SWITCHER_PATH")

	if Path == "" {
		log.Fatal("Environment variable DNS_SWITCHER_PATH is not set. Please set it before running.")
		return
	}
}

type Profile struct {
	Name      string
	Servers   []string
	Interface string
}

// LoadProfilesDns reads profiles from profiles.yaml and returns a slice of Profile
func LoadProfilesDns() []Profile {
	var out []Profile
	if Path == "" {
		return out
	}

	viper.Reset()
	viper.SetConfigFile(Path)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return out
	}

	m := viper.Get("profiles")
	if m == nil {
		return out
	}

	profilesMap, ok := m.(map[string]interface{})
	if !ok {
		return out
	}

	for name, v := range profilesMap {
		p := Profile{Name: name, Servers: []string{}}

		vv, ok1 := v.(map[string]interface{})
		if !ok1 {
			continue
		}

		// ipv4
		if ipsRaw, exists := vv["ipv4"]; exists {
			switch ips := ipsRaw.(type) {
			case []interface{}:
				for _, ip := range ips {
					if s, ok := ip.(string); ok {
						p.Servers = append(p.Servers, s)
					}
				}
			case []string:
				p.Servers = append(p.Servers, ips...)
			}
		}

		out = append(out, p)
	}

	return out
}

// FindProfile searches for a profile by name in the slice
func FindProfile(profiles []Profile, name string) (*Profile, bool) {
	for i := range profiles {
		if profiles[i].Name == name {
			return &profiles[i], true
		}
	}
	return nil, false
}

// AddProfile adds a new profile to YAML
// AddProfile adds a new DNS profile to profiles.yaml
func AddProfile(name string, servers []string) error {
	if Path == "" {
		return fmt.Errorf("config path not set")
	}

	// Load config
	viper.Reset()
	viper.SetConfigFile(Path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		// If file doesn't exist, initialize config
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			viper.Set("profiles", map[string]interface{}{})
		}
	}

	// Ensure profiles map exists
	profiles := viper.GetStringMap("profiles")
	if profiles == nil {
		profiles = make(map[string]interface{})
	}

	// Add or overwrite profile
	profiles[name] = map[string]interface{}{
		"ipv4": servers,
	}

	viper.Set("profiles", profiles)

	// Save config properly
	if err := viper.WriteConfig(); err != nil {
		if err := viper.SafeWriteConfigAs(Path); err != nil {
			return fmt.Errorf("cannot write config: %v", err)
		}
	}

	return nil
}
