package config

import (
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

func LoadDomain(name string) (*Domain, error) {
	viper.SetConfigFile(cfgPath)
	viper.SetConfigType("yaml")

	return &Domain{
		Base: viper.GetString(name + ".base"),
		Name: viper.GetString(name + ".name"),
		Pass: viper.GetString(name + ".pass"),
		User: viper.GetString(name + ".base"),
	}, nil
}

func LoadConfig() *config {
	viper.SetConfigFile(cfgPath)
	viper.SetConfigType("yaml")

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
