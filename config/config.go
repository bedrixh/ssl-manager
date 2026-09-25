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
	CACertificates       []CertificateConfig `yaml:"CACertificates" json:"CACertificates" toml:"CACertificates"`
	Certificates         []CertificateConfig `yaml:"Certificates" json:"Certificates" toml:"Certificate"`
	CertificatesDefaults CertificateConfig   `yaml:"CertificatesDefaults" json:"CertificatesDefaults" toml:"CertificatesDefaults"`
}

type CertificateConfig struct {
	Name               string   `yaml:"Name" json:"Name" toml:"Name"`
	Path               string   `yaml:"Path" json:"Path" toml:"Path"`
	UserOwner          string   `yaml:"UserOwner" json:"UserOwner" toml:"UserOwner"`
	GroupOwner         string   `yaml:"GroupOwner" json:"GroupOwner" toml:"GroupOwner"`
	Permissions        uint16   `yaml:"Permissions" json:"Permissions" toml:"Permissions"`
	CACertName         string   `yaml:"CACertName" json:"CACertName" toml:"CACertName"`
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
	LiveConfigReload     string                `yaml:"LiveConfigReload" json:"LiveConfigReload" toml:"LiveConfigReload"`
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

func (c *Configuration) GetFormatedJson() (string, error) {
	json, err := json.MarshalIndent(c, "", "  ")
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

func (c *CertificateConfig) GetCACertConfig() *CertificateConfig {
	for i := range len(appConfig.CACertificates) {
		if appConfig.CACertificates[i].Name == c.CACertName {
			return &appConfig.CACertificates[i]
		}
	}
	return nil
}

func GetConfig() (*Configuration, error) {
	if appConfig == nil {
		return nil, fmt.Errorf("cannot get configuration, because it is not loaded")
	}
	return appConfig, nil
}

func GetConfigNoErr() *Configuration {
	if appConfig == nil {
		panic(fmt.Errorf("cannot get configuration, because it is not loaded"))
	}
	return appConfig
}

// variable holding current config path, for live config reloads
var loadedConfigPath string

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
	loadedConfigPath = path

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

	for i := 0; i < len(config.CACertificates); i++ {
		certConfig := &config.CACertificates[i]
		err := validateCertificateConfig(certConfig)
		if err != nil {
			return fmt.Errorf("certificate: %s (%s)", certConfig.Name, err)
		}
	}

	if config.Daemon.RenewIntervalDays <= 0 {
		return fmt.Errorf("daemon renew interval must be greater than zero")
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
		return fmt.Errorf("OrganizationName cannot be empty")

	case certConfig.Email == "":
		return fmt.Errorf("Email cannot be empty")

	case 0 >= certConfig.ValidDays:
		return fmt.Errorf("Validity must be greater than 0 days")

	case certConfig.ValidDays < certConfig.RenewThresholdDays:
		return fmt.Errorf("renew threshold has to be smaller or equal to validity")

	}

	return nil
}

func validateCACertificateConfig(certConfig *CertificateConfig) error {

	switch {
	case certConfig.Name == "":
		return fmt.Errorf("Name cannot be empty")

	case certConfig.Path == "":
		return fmt.Errorf("Path cannot be empty")

	case certConfig.OrganizationName == "":
		return fmt.Errorf("OrganizationName cannot be empty")

	case certConfig.Email == "":
		return fmt.Errorf("Email cannot be empty")

	case 0 >= certConfig.ValidDays:
		return fmt.Errorf("Validity must be greater than 0 days")

	case certConfig.ValidDays < certConfig.RenewThresholdDays:
		return fmt.Errorf("renew threshold has to be smaller or equal to validity")

	}

	return nil
}

func populateDefaults(config *Configuration) {
	if config.CertificatesDefaults.Permissions == 0 {
		config.CertificatesDefaults.Permissions = 0644
	}

	// certificates defaults
	for i := range len(config.Certificates) {
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
		if certificateConfig.Permissions == 0 {
			certificateConfig.Permissions = config.CertificatesDefaults.Permissions
		}
	}

	// CA certificates defaults
	for i := range len(config.CACertificates) {
		if config.CACertificates[i].Permissions == 0 {
			config.CACertificates[i].Permissions = 0644
		}
	}
}
