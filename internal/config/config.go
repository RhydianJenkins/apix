package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func ExpandPath(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to resolve home directory: %w", err)
	}

	if path == "~" {
		return home, nil
	}

	return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
}

func defaultCfgPaths() []string {
	var paths []string

	for _, p := range filepath.SplitList(os.Getenv("APIX_CONFIG")) {
		if p == "" {
			continue
		}

		if expanded, err := ExpandPath(p); err == nil {
			p = expanded
		}

		paths = append(paths, p)
	}

	if len(paths) == 0 {
		return []string{filepath.Join(os.Getenv("HOME"), ".apix.yaml")}
	}

	return paths
}

var CfgPaths = defaultCfgPaths()

func WritePath() string {
	return CfgPaths[len(CfgPaths)-1]
}

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

func loadMerged() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	for _, path := range CfgPaths {
		v.SetConfigFile(path)

		err := v.MergeInConfig()
		if err == nil {
			continue
		}

		if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
			return nil, fmt.Errorf("no config file '%s' found. Run `apix new` to create one and add your first domain", path)
		}

		return nil, fmt.Errorf("error reading config file %s: %w", path, err)
	}

	return v, nil
}

func createEmptyConfig(path string) error {
	v := viper.New()
	v.SetConfigType("yaml")

	return v.SafeWriteConfigAs(path)
}

func loadWriteTarget() (*viper.Viper, error) {
	path := WritePath()

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			return nil, fmt.Errorf("error reading config: %w", err)
		}

		if createErr := createEmptyConfig(path); createErr != nil {
			return nil, fmt.Errorf("error creating config: %w", createErr)
		}
	}

	return v, nil
}

func SetActiveName(activeName string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	if activeName == "" {
		return fmt.Errorf("Active domain cannot be empty\n")
	}

	if _, exists := cfg.Domains[activeName]; !exists {
		return fmt.Errorf("Domain %q does not exist.\nUse `apix list` to show available domains or `apix new` to add new ones.\n", activeName)
	}

	v, err := loadWriteTarget()
	if err != nil {
		return err
	}

	v.Set("active", activeName)

	return v.WriteConfig()
}

func GetActiveDomain() (*Domain, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	domain := cfg.Domains[cfg.Active]
	domain.Name = cfg.Active

	return &domain, nil
}

func RequireActiveDomain() (*Domain, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	if len(cfg.Domains) == 0 {
		return nil, fmt.Errorf("no domains configured. Run `apix new <name> <base>` to create one or see `apix --help` for configuration options")
	}
	if cfg.Active == "" {
		return nil, fmt.Errorf("no active domain set. Run `apix switch <name>` to set one or `apix list` to see available domains")
	}
	if _, ok := cfg.Domains[cfg.Active]; !ok {
		return nil, fmt.Errorf("active domain %q does not exist. Run `apix switch <name>` to set a valid active domain or `apix list` to see available domains", cfg.Active)
	}
	domain := cfg.Domains[cfg.Active]
	domain.Name = cfg.Active
	return &domain, nil
}

func LoadDomain(name string) (*Domain, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	domain, ok := cfg.Domains[name]

	if !ok {
		return nil, fmt.Errorf("domain %q not found", name)
	}

	domain.Name = name

	return &domain, nil
}

func LoadConfig() (*config, error) {
	v, err := loadMerged()
	if err != nil {
		return nil, err
	}

	var cfg config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	return &cfg, nil
}

func SetDomain(domain *Domain) error {
	v, err := loadWriteTarget()
	if err != nil {
		return err
	}

	v.Set("domains."+domain.Name, domain)
	v.Set("active", domain.Name)

	return v.WriteConfig()
}

func SetDomainHeader(name, key, value string) error {
	if _, err := LoadDomain(name); err != nil {
		return err
	}

	v, err := loadWriteTarget()
	if err != nil {
		return err
	}

	headersKey := "domains." + name + ".headers"

	headers := v.GetStringMapString(headersKey)
	if headers == nil {
		headers = make(map[string]string)
	}
	headers[key] = value

	v.Set(headersKey, headers)

	return v.WriteConfig()
}

func RemoveDomain(nameToRemove string) error {
	merged, err := loadMerged()
	if err != nil {
		return err
	}

	if !merged.IsSet("domains." + nameToRemove) {
		return fmt.Errorf("Domain %q does not exist in your list of domains. See domains with `apix list`\n", nameToRemove)
	}

	if merged.GetString("active") == nameToRemove {
		return fmt.Errorf("Unable to remove %q as it is currently your active domain.\n", nameToRemove)
	}

	v, err := loadWriteTarget()
	if err != nil {
		return err
	}

	if !v.IsSet("domains." + nameToRemove) {
		return fmt.Errorf("Domain %q is defined in another config file, remove it from there.\n", nameToRemove)
	}

	domainsMap := v.GetStringMap("domains")
	delete(domainsMap, nameToRemove)
	v.Set("domains", domainsMap)

	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("Error writing config: %w\n", err)
	}

	return nil
}

// GetDomainNames is used for shell completion and best-effort iteration
// (e.g. `apix auth --all`), neither of which has anywhere useful to surface
// an error, so a missing/unreadable config is treated the same as no domains.
func GetDomainNames() []string {
	cfg, err := LoadConfig()
	if err != nil {
		return nil
	}

	domainNames := make([]string, 0, len(cfg.Domains))
	for k := range cfg.Domains {
		domainNames = append(domainNames, k)
	}

	return domainNames
}
