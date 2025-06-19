package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var cfgPath = filepath.Join(os.Getenv("HOME"), ".apix.yaml")

type Domain struct {
	Base string `yaml:"base"`
	Name string `yaml:"name"`
	Pass string `yaml:"pass"`
	User string `yaml:"user"`
}

type config struct {
	Active  string `mapstructure:"active"`
	Domains map[string]Domain `mapstructure:"domains"`
}

func SetActiveName(activeName string) (error) {
	cfg := LoadConfig()

	if activeName == "" {
		return fmt.Errorf("Active domain cannot be empty\n")
	}

	// TODO viper.InConfig exists...
	if _, exists := cfg.Domains[activeName]; !exists {
		return fmt.Errorf("Domain %q does not exist.\nUse 'list' command to show available domains or 'set' to add new ones.\n", activeName)
	}

	viper.Set("active", activeName)
	viper.WriteConfig()

	return nil
}

func LoadActiveDomain() (*Domain, error) {
	cfg := LoadConfig()
	domain := cfg.Domains[cfg.Active]

	return &domain, nil
}

func LoadDomain(name string) (*Domain, error) {
	cfg := LoadConfig()

	domain, ok := cfg.Domains[name]

	if !ok {
		return nil, fmt.Errorf("domain %q not found", name)
	}

	return &domain, nil
}

func LoadConfig() *config {
	viper.SetConfigFile(cfgPath)

	if err := viper.ReadInConfig(); err != nil {
		if _, statErr := os.Stat(cfgPath); os.IsNotExist(statErr) {
			viper.SafeWriteConfigAs(cfgPath)
		} else {
			println("Error reading config file:", err)
		}
	}

	var cfg config

	if err := viper.Unmarshal(&cfg); err != nil {
		println("Error unmarshalling config:", err)
	}

	return &cfg
}

func SetDomain(domain *Domain) {
	viper.SetConfigFile(cfgPath)

    if err := viper.MergeInConfig(); err != nil {
        if _, statErr := os.Stat(cfgPath); os.IsNotExist(statErr) {
            viper.SafeWriteConfigAs(cfgPath)
        } else {
			println("Error reading config file:", err)
        }
    }

	viper.Set("domains."+domain.Name, domain)
	viper.Set("active", domain.Name)
	viper.WriteConfig()
}

func RemoveDomain(nameToRemove string) error {
	viper.SetConfigFile(cfgPath)

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	// TODO Rhydian this doesn't seem to remove the key?
	viper.Set("domains."+nameToRemove, nil)

	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("error writing config: %w", err)
	}

	return nil
}
