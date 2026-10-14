package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// ExpandPath expands a leading "~" to the user's home directory. Shells don't
// always do this themselves (e.g. after a ":" in a path list), so it is applied
// to -c values and $APIX_CONFIG entries alike.
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

// defaultCfgPaths returns the config files to use when no -c flag is given:
// $APIX_CONFIG (an OS path list, so a single path works as before) or
// ~/.apix.yaml.
func defaultCfgPaths() []string {
	env := os.Getenv("APIX_CONFIG")
	if env == "" {
		return []string{filepath.Join(os.Getenv("HOME"), ".apix.yaml")}
	}

	paths := filepath.SplitList(env)
	for i, p := range paths {
		if expanded, err := ExpandPath(p); err == nil {
			paths[i] = expanded
		}
	}

	return paths
}

// CfgPaths are the config files, in merge order: later files deep-merge over
// earlier ones. Only the last one is ever written to, see WritePath.
var CfgPaths = defaultCfgPaths()

// WritePath is the config file that receives all writes (the last of CfgPaths).
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

// createEmptyConfig creates path without touching any other config's content.
func createEmptyConfig(path string) error {
	v := viper.New()
	v.SetConfigType("yaml")

	return v.SafeWriteConfigAs(path)
}

// loadMerged reads every file in CfgPaths and deep-merges them, later files
// overriding earlier ones. A missing file is skipped, except the write target
// which is created.
func loadMerged() *viper.Viper {
	v := viper.New()
	v.SetConfigType("yaml")

	for _, path := range CfgPaths {
		v.SetConfigFile(path)

		err := v.MergeInConfig()
		if err == nil {
			continue
		}

		if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
			if path == WritePath() {
				if createErr := createEmptyConfig(path); createErr != nil {
					fmt.Println("Error creating config file:", createErr)
				}
			}
			continue
		}

		fmt.Println("Error reading config file:", err)
	}

	return v
}

// loadWriteTarget reads only WritePath, so that values merged in from earlier
// files are never persisted into it.
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
	cfg := LoadConfig()

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

func LoadConfig() *config {
	var cfg config

	if err := loadMerged().Unmarshal(&cfg); err != nil {
		fmt.Println("Error unmarshalling config:", err)
	}

	return &cfg
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

// SetDomainHeader sets a single header/metadata key on an existing domain
// without changing which domain is active - unlike SetDomain, which always
// makes the saved domain the active one. Only the headers are written, and
// only to WritePath, so the domain may be defined in an earlier config file.
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
	merged := loadMerged()

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

func GetDomainNames() []string {
	cfg := LoadConfig()
	domainNames := make([]string, 0, len(cfg.Domains))
	for k := range cfg.Domains {
		domainNames = append(domainNames, k)
	}

	return domainNames
}
