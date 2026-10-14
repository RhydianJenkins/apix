package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func defaultCfgPath() string {
	if p := os.Getenv("APIX_CONFIG"); p != "" {
		return p
	}

	return filepath.Join(os.Getenv("HOME"), ".apix.yaml")
}

var CfgPath = defaultCfgPath()

const (
	ProtocolHTTP = "http"
	ProtocolGRPC = "grpc"
)

type Domain struct {
	Base    string            `yaml:"base"`
	Name    string            `yaml:"-"`                 // inferred from the domain's key in the domains map, never persisted
	Headers map[string]string `yaml:"headers,omitempty"` // HTTP headers or gRPC metadata, depending on Protocol()

	HTTP *HTTPOptions `yaml:"http,omitempty"`
	GRPC *GRPCOptions `yaml:"grpc,omitempty"`
	Auth *AuthOptions `yaml:"auth,omitempty"`
}

type AuthOptions struct {
	Command string `yaml:"command"` // run via `apix auth`; responsible for persisting any header itself, e.g. via `apix set-header`
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

func (d *Domain) IsGRPC() bool {
	return d != nil && d.GRPC != nil
}

func (d *Domain) Protocol() string {
	if d.IsGRPC() {
		return ProtocolGRPC
	}
	return ProtocolHTTP
}

func (d *Domain) HasAuth() bool {
	return d != nil && d.Auth != nil && d.Auth.Command != ""
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
	domain.Name = cfg.Active

	return &domain
}

func LoadDomain(name string) (*Domain, error) {
	cfg := LoadConfig()

	domain, ok := cfg.Domains[name]

	if !ok {
		return nil, fmt.Errorf("domain %q not found", name)
	}

	domain.Name = name

	return &domain, nil
}

// setConfigFile points viper at CfgPath and explicitly sets the config
// format to YAML, since viper otherwise infers it from CfgPath's file
// extension - which may not be ".yaml" when CfgPath is user-supplied
// (e.g. via --config).
func setConfigFile() {
	viper.SetConfigFile(CfgPath)
	viper.SetConfigType("yaml")
}

func LoadConfig() *config {
	setConfigFile()

	if err := viper.ReadInConfig(); err != nil {
		if _, statErr := os.Stat(CfgPath); os.IsNotExist(statErr) {
			viper.SafeWriteConfigAs(CfgPath)
		} else {
			fmt.Println("Error reading config file:", err)
		}
	}

	var cfg config

	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Println("Error unmarshalling config:", err)
	}

	return &cfg
}

func SetDomain(domain *Domain) {
	setConfigFile()

	if err := viper.MergeInConfig(); err != nil {
		if _, statErr := os.Stat(CfgPath); os.IsNotExist(statErr) {
			viper.SafeWriteConfigAs(CfgPath)
		} else {
			fmt.Println("Error reading config file:", err)
		}
	}

	viper.Set("domains."+domain.Name, domain)
	viper.Set("active", domain.Name)
	viper.WriteConfig()
}

// SetDomainHeader sets a single header/metadata key on an existing domain
// without changing which domain is active - unlike SetDomain, which always
// makes the saved domain the active one.
func SetDomainHeader(name, key, value string) error {
	domain, err := LoadDomain(name)
	if err != nil {
		return err
	}

	setConfigFile()

	if err := viper.MergeInConfig(); err != nil {
		if _, statErr := os.Stat(CfgPath); os.IsNotExist(statErr) {
			viper.SafeWriteConfigAs(CfgPath)
		} else {
			fmt.Println("Error reading config file:", err)
		}
	}

	if domain.Headers == nil {
		domain.Headers = make(map[string]string)
	}
	domain.Headers[key] = value

	viper.Set("domains."+name, domain)

	return viper.WriteConfig()
}

func RemoveDomain(nameToRemove string) error {
	setConfigFile()

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
