package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var CfgPath = filepath.Join(os.Getenv("HOME"), ".apix.yaml")

const (
	ProtocolHTTP = "http"
	ProtocolGRPC = "grpc"
)

type Domain struct {
	Base     string            `yaml:"base"`
	Name     string            `yaml:"name"`
	Protocol string            `yaml:"protocol,omitempty"` // "http" (default, empty) | "grpc"
	Headers  map[string]string `yaml:"headers,omitempty"`  // HTTP headers or gRPC metadata, depending on Protocol

	HTTP *HTTPOptions `yaml:"http,omitempty"`
	GRPC *GRPCOptions `yaml:"grpc,omitempty"`
}

type HTTPOptions struct {
	User            string `yaml:"user,omitempty"`
	Pass            string `yaml:"pass,omitempty"`
	OpenAPISpecPath string `yaml:"openapispecpath,omitempty"`
	Port            int    `yaml:"port,omitempty"` // if set, base must be host-only and is combined as "scheme://host:port"
}

type GRPCOptions struct {
	Insecure bool `yaml:"insecure,omitempty"`
	Port     int  `yaml:"port,omitempty"` // if set, combined with Base (host-only) as "host:port"
}

// An empty Protocol defaults to HTTP for backwards compatibility.
func (d *Domain) IsGRPC() bool {
	return d != nil && d.Protocol == ProtocolGRPC
}

type config struct {
	Active  string            `mapstructure:"active"`
	Domains map[string]Domain `mapstructure:"domains"`
}

func SetActiveName(activeName string) error {
	cfg := LoadConfig()

	if activeName == "" {
		return fmt.Errorf("Active domain cannot be empty\n")
	}

	if _, exists := cfg.Domains[activeName]; !exists {
		return fmt.Errorf("Domain %q does not exist.\nUse `apix list` to show available domains or `apix new` to add new ones.\n", activeName)
	}

	viper.Set("active", activeName)
	viper.WriteConfig()

	return nil
}

func GetActiveDomain() *Domain {
	cfg := LoadConfig()
	domain := cfg.Domains[cfg.Active]

	return &domain
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
	viper.SetConfigFile(CfgPath)

	if err := viper.ReadInConfig(); err != nil {
		if _, statErr := os.Stat(CfgPath); os.IsNotExist(statErr) {
			viper.SafeWriteConfigAs(CfgPath)
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
	viper.SetConfigFile(CfgPath)

	if err := viper.MergeInConfig(); err != nil {
		if _, statErr := os.Stat(CfgPath); os.IsNotExist(statErr) {
			viper.SafeWriteConfigAs(CfgPath)
		} else {
			println("Error reading config file:", err)
		}
	}

	viper.Set("domains."+domain.Name, domain)
	viper.Set("active", domain.Name)
	viper.WriteConfig()
}

func RemoveDomain(nameToRemove string) error {
	viper.SetConfigFile(CfgPath)

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	if !viper.IsSet("domains." + nameToRemove) {
		return fmt.Errorf("Domain %q does not exist in your list of domains. See domains with `apix list`\n", nameToRemove)
	}

	if viper.GetString("active") == nameToRemove {
		return fmt.Errorf("Unable to remove %q as it is currently your active domain.\n", nameToRemove)
	}

	domainsMap := viper.GetStringMap("domains")
	delete(domainsMap, nameToRemove)
	viper.Set("domains", domainsMap)

	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("Error writing config: %w\n", err)
	}

	return nil
}

func GetDomainNames() []string {
	cfg := LoadConfig()
	domainNames := make([]string, 0, len(cfg.Domains))
	for k := range cfg.Domains {
		domainNames = append(domainNames, k)
	}

	return domainNames
}
