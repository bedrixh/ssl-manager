package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

var appConfig *Configuration = nil

type Configuration struct {
	Daemon               DaemonConfig        `yaml:"Daemon" json:"Daemon" toml:"Daemon"`
	CACertificate        CertificateConfig   `yaml:"CACertificate" json:"CACertificate" toml:"CACertificate"`
	Certificates         []CertificateConfig `yaml:"Certificates" json:"Certificates" toml:"Certificate"`
	CertificatesDefaults CertificateConfig   `yaml:"CertificatesDefaults" json:"CertificatesDefaults" toml:"CertificatesDefaults"`
}

type CertificateConfig struct {
	Name               string   `yaml:"Name" json:"Name" toml:"Name"`
	Path               string   `yaml:"Path" json:"Path" toml:"Path"`
	OrganizationName   string   `yaml:"OrganizationName" json:"OrganizationName" toml:"OrganizationName"`
	Email              string   `yaml:"Email" json:"Email" toml:"Email"`
	IPs                []string `yaml:"IPs" json:"IPs" toml:"IPs"`
	DNSNames           []string `yaml:"DNSNames" json:"DNSNames" toml:"DNSNames"`
	ValidDays          int      `yaml:"ValidDays" json:"ValidDays" toml:"ValidDays"`
	RenewThresholdDays int      `yaml:"RenewThresholdDays" json:"RenewThresholdDays" toml:"RenewThresholdDays"`
}

type DaemonConfig struct {
	RenewIntervalDays    int                   `yaml:"RenewIntervalDays" json:"RenewIntervalDays" toml:"RenewIntervalDays"`
	NotificationWebhooks []NotificationWebhook `yaml:"NotificationWebhooks" json:"NotificationWebhooks" toml:"NotificationWebhook"`
}

type NotificationWebhook struct {
	Name          string            `yaml:"Name" json:"Name" toml:"Name"`
	Url           string            `yaml:"Url" json:"Url" toml:"Url"`
	PostData      map[string]string `yaml:"PostData" json:"PostData" toml:"PostData"`
	NotifyFail    bool              `yaml:"NotifyFail" json:"NotifyFail" toml:"NotifyFail"`
	NotifySuccess bool              `yaml:"NotifySuccess" json:"NotifySuccess" toml:"NotifySuccess"`
}

func (c *CertificateConfig) GetIPAdresses() ([]net.IP, error) {
	var ipAdresses = make([]net.IP, len(c.IPs))
	for i := 0; i < len(c.IPs); i++ {
		ipAdresses[i] = net.ParseIP(c.IPs[i])
		if ipAdresses[i] == nil {
			return nil, fmt.Errorf("ip address %d is not valid ip address", i)
		}
	}
	return ipAdresses, nil
}

func (c *Configuration) GetYaml() (string, error) {
	yaml, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("error encoding config into yaml %s", err)
	}
	return string(yaml), nil
}

func (c *Configuration) GetToml() (string, error) {
	toml, err := toml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("error encoding config into toml %s", err)
	}
	return string(toml), nil
}

func (c *Configuration) GetJson() (string, error) {
	json, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("error encoding config into json %s", err)
	}
	return string(json), nil
}

func (c *CertificateConfig) GetCertPath() string {
	return filepath.Join(c.Path, "cert.pem")
}
func (c *CertificateConfig) GetKeyPath() string {
	return filepath.Join(c.Path, "key.pem")
}

func (c *CertificateConfig) CertificateExists() bool {
	fileInfo, err := os.Stat(c.GetCertPath())
	if err != nil {
		return false
	}
	return !fileInfo.IsDir()
}

func GetConfig() (*Configuration, error) {
	if appConfig == nil {
		return nil, fmt.Errorf("config is empty")
	}
	return appConfig, nil
}

func LoadAppConfig(path string) error {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	config := &Configuration{}

	switch {
	case strings.HasSuffix(path, ".json"):
		{
			config, err = loadJsonConfig(configBytes)
		}
	case strings.HasSuffix(path, ".toml"):
		{
			config, err = loadTomlConfig(configBytes)
		}
	case strings.HasSuffix(path, ".yaml"):
		{
			config, err = loadYamlConfig(configBytes)
		}
	default:
		{
			err = fmt.Errorf("unknown file type, use json, toml or yaml")
		}
	}
	if err != nil {
		return err
	}

	populateDefaults(config)

	err = validateConfig(config)
	if err != nil {
		return fmt.Errorf("config validation error: %s", err)
	}

	appConfig = config

	return nil
}

func loadJsonConfig(jsonBytes []byte) (*Configuration, error) {
	var appConfig = &Configuration{}
	err := json.Unmarshal(jsonBytes, appConfig)
	if err != nil {
		return nil, err
	}
	return appConfig, nil
}

func loadTomlConfig(tomlBytes []byte) (*Configuration, error) {
	var appConfig = &Configuration{}
	err := toml.Unmarshal(tomlBytes, appConfig)
	if err != nil {
		return nil, err
	}
	return appConfig, nil
}

func loadYamlConfig(yamlBytes []byte) (*Configuration, error) {
	var appConfig = &Configuration{}
	err := yaml.Unmarshal(yamlBytes, appConfig)
	if err != nil {
		return nil, err
	}
	return appConfig, nil

}

func validateConfig(config *Configuration) error {
	for i := 0; i < len(config.Certificates); i++ {
		certConfig := &config.Certificates[i]
		err := validateCertificateConfig(certConfig)
		if err != nil {
			return fmt.Errorf("certificate: %s (%s)", certConfig.Name, err)
		}
	}
	err := validateCertificateConfig(&config.CACertificate)
	if err != nil {
		return fmt.Errorf("CA certificate configuration: %s", err)
	}

	if config.Daemon.RenewIntervalDays == 0 {
		return fmt.Errorf("daemon renew interval cannot be empty")
	}

	return nil
}

func validateCertificateConfig(certConfig *CertificateConfig) error {

	if _, err := certConfig.GetIPAdresses(); err != nil {
		return err
	}

	switch {
	case certConfig.Name == "":
		return fmt.Errorf("Name cannot be empty")

	case certConfig.Path == "":
		return fmt.Errorf("Path cannot be empty")

	case certConfig.OrganizationName == "":
		return fmt.Errorf("OrganiationName cannot be empty")

	case certConfig.Email == "":
		return fmt.Errorf("Email cannot be empty")

	case 0 > certConfig.ValidDays:
		return fmt.Errorf("Validity has to be bigger than 0 days")

	case certConfig.ValidDays < certConfig.RenewThresholdDays:
		return fmt.Errorf("renew threshold has to be smaller or equal to validity")

	}

	return nil
}

func populateDefaults(config *Configuration) {
	for i := 0; i < len(config.Certificates); i++ {
		certificateConfig := &config.Certificates[i]
		if certificateConfig.ValidDays == 0 {
			certificateConfig.ValidDays = config.CertificatesDefaults.ValidDays
		}

		if certificateConfig.OrganizationName == "" {
			certificateConfig.OrganizationName = config.CertificatesDefaults.OrganizationName
		}

		if certificateConfig.Email == "" {
			certificateConfig.Email = config.CertificatesDefaults.Email
		}

		if certificateConfig.RenewThresholdDays == 0 {
			certificateConfig.RenewThresholdDays = config.CertificatesDefaults.RenewThresholdDays
		}

	}
}
