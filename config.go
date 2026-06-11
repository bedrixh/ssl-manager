package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"go.yaml.in/yaml/v4"
)

type Configuration struct {
	CACertificate        CertificateConfig
	Certificates         []CertificateConfig
	CertificatesDefaults CertificateConfig
}

type CertificateConfig struct {
	Name               string
	Path               string
	OrganizationName   string
	Email              string
	IPs                []string
	DNSNames           []string
	ValidDays          int
	RenewThresholdDays int
}

func (c *CertificateConfig) GetIPAdresses() ([]net.IP, error) {
	var ipAdresses []net.IP = make([]net.IP, len(c.IPs))
	for i := 0; i < len(c.IPs); i++ {
		ipAdresses[i] = net.ParseIP(c.IPs[i])
		if ipAdresses[i] == nil {
			return nil, fmt.Errorf("ip address number %d is not valid ip address",i)
		}
	}
	return ipAdresses, nil
}

func GetConfig(path string) (*Configuration, error) {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	appConfig := &Configuration{}

	switch {
	case strings.HasSuffix(path, ".json"):
		{
			appConfig, err = loadJsonConfig(configBytes)
		}
	case strings.HasSuffix(path, ".toml"):
		{
			appConfig, err = loadTomlConfig(configBytes)
		}
	case strings.HasSuffix(path, ".yaml"):
		{
			appConfig, err = loadYamlConfig(configBytes)
		}
	default:
		{
			err = fmt.Errorf("unknown file type, use json, toml or yaml")
		}
	}
	if err != nil {
		return nil, err
	}

	populateDefaults(appConfig)

	fmt.Println(appConfig)
	err = validateConfig(appConfig)
	if err != nil {
		return nil, fmt.Errorf("Config validation error (%s)", err)
	}

	return appConfig, nil
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
			return fmt.Errorf("Certificate: %s (%s)", certConfig.Name, err)
		}
	}
	err := validateCertificateConfig(&config.CACertificate)
	if err != nil {
		return fmt.Errorf("CA certificate configuration (%s)", err)
	}

	return nil
}

func validateCertificateConfig(certConfig *CertificateConfig) error {

	if _, err := certConfig.GetIPAdresses(); err != nil {
		return err
	}

	switch {
	case "" == certConfig.Name:
		return fmt.Errorf("Name cannot be empty")

	case "" == certConfig.Path:
		return fmt.Errorf("Path cannot be empty")

	case "" == certConfig.OrganizationName:
		return fmt.Errorf("OrganiationName cannot be empty")

	case "" == certConfig.Email:
		return fmt.Errorf("Email cannot be empty")

	case 0 > certConfig.ValidDays:
		return fmt.Errorf("Validity has to be bigger than 0 days")

	case certConfig.ValidDays < certConfig.RenewThresholdDays:
		return fmt.Errorf("Renew threshold has to be smaller or equal to validity")

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
