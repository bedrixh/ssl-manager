package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
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
	CommonName         string
	Email              string
	IPs                []string
	DNSNames           []string
	ValidDays          int
	RenewThresholdDays int
}

func (c *CertificateConfig) GetIPAdresses() []net.IP {
	var ipAdresses []net.IP = make([]net.IP, len(c.IPs))
	for i := 0; i < len(c.IPs); i++ {
		ipAdresses[i] = net.ParseIP(c.IPs[i])
	}
	return ipAdresses
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
	default:
		{
			return nil, fmt.Errorf("unknown file type, use json or toml")
		}
	}
	if err != nil {
		return nil, err
	}

	populateDefault(appConfig)

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

func populateDefault(config *Configuration) {
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
