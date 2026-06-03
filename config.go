package main

import (
	"encoding/json"
	"github.com/BurntSushi/toml"
	"io"
	"os"
	"strings"
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

func LoadConfig(path string) *Configuration {
	file, err := os.Open(path)

	if os.IsNotExist(err) == true {
		panic("Cannot read config file, does not exist")
	}

	if os.IsPermission(err) == true {
		panic("Cannot read config file, permission denied")
	}

	if err != nil {
		panic(err)
	}

	configBytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	appConfig := &Configuration{}

	// checking config file type
	if strings.HasSuffix(path, ".json") {
		appConfig = loadJsonConfig(configBytes)
	} else {
		if strings.HasSuffix(path, ".toml") {
			appConfig = loadTomlConfig(configBytes)
		} else {
			panic("unknown file type, use json or toml")
		}
	}

	populateDefault(appConfig)

	return appConfig
}

func loadJsonConfig(jsonBytes []byte) *Configuration {
	var appConfig = &Configuration{}
	err := json.Unmarshal(jsonBytes, appConfig)
	if err != nil {
		panic(err)
	}
	return appConfig
}

func loadTomlConfig(tomlBytes []byte) *Configuration {
	var appConfig = &Configuration{}
	err := toml.Unmarshal(tomlBytes, appConfig)
	if err != nil {
		panic(err)
	}
	return appConfig
}

func populateDefault(config *Configuration) {
	for i := 0; i < len(config.Certificates); i++ {
		certificateConfig := &config.Certificates[i]
		if certificateConfig.ValidDays == 0 {
			certificateConfig.ValidDays = config.CertificatesDefaults.ValidDays
		}
	}
}
