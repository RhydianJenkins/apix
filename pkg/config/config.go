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

func SetActiveDomain(activeDomain string) (error) {
	cfg := LoadConfig()

	if activeDomain == "" {
		return fmt.Errorf("active domain cannot be empty")
	}

	if _, exists := cfg.Domains[activeDomain]; !exists {
		return fmt.Errorf("domain %q does not exist", activeDomain)
	}

	viper.Set("active", activeDomain)
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
			viper.SafeWriteConfigAs(cfgPath) // writes only if not present
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
            viper.SafeWriteConfigAs(cfgPath) // writes only if not present
        } else {
			println("Error reading config file:", err)
        }
    }

	viper.Set("domains."+domain.Name, domain)
	viper.Set("active", domain.Name)
	viper.WriteConfig()
}
