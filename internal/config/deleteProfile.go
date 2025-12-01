package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// DeleteProfile removes a profile from profiles.yaml by name
func DeleteProfile(name string) error {
	viper.SetConfigFile(Path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("cannot read config: %v", err)
	}

	profiles := viper.GetStringMap("profiles")
	if profiles == nil {
		return fmt.Errorf("no profiles found in config")
	}

	if _, exists := profiles[name]; !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	delete(profiles, name)

	viper.Set("profiles", profiles)
	if err := viper.WriteConfig(); err != nil {
		if err := viper.SafeWriteConfigAs(Path); err != nil {
			return err
		}
	}

	return nil
}
