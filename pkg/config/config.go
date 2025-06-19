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

type Config struct {
	Active  string `mapstructure:"active"`
	Domains map[string]Domain `mapstructure:"domains"`
}

func loadConfig(name string) (*Domain, error) {
	viper.SetConfigFile(cfgPath)
	viper.SetConfigType("yaml")
	return &Domain{
		Base: viper.GetString(name + ".base"),
		Name: viper.GetString(name + ".name"),
		Pass: viper.GetString(name + ".pass"),
		User: viper.GetString(name + ".base"),
	}, nil
}

func SetDomain(domain *Domain) {
	viper.Set("domains."+domain.Name, domain)
	viper.Set("active", domain.Name)
	viper.SetConfigFile(cfgPath)
	viper.WriteConfig()
}
