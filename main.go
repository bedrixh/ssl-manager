package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

var Config *Configuration

func main() {
	argConfFilePtr := flag.String("config", "./conf.toml", "Config file to be loaded on the start of the program (can be json or toml)")
	argGenCAPtr := flag.Bool("gen-ca", false, "Generates certification authority certificate and stores it on the disk")
	argRenewCerts := flag.Bool("renew-certs", false, "Creates missing certificates and renews certificates that will expire soon")
	flag.Parse()

	Config = LoadConfig(*argConfFilePtr)
	bytes, _ := toml.Marshal(Config)
	fmt.Println(string(bytes))
	fmt.Println(GetValidDaysRemaining(GetCertFromDisk("./CA/cert.pem")))

	if *argGenCAPtr == true {
		GenerateCACert(&Config.CACertificate)
	}

	if *argRenewCerts == true {
		renewCerts()
	}

}

func renewCerts() {
	for i := 0; i < len(Config.Certificates); i++ {
		certificateConfig := &Config.Certificates[i]
		certExists := getCertificateExists(certificateConfig)
		if certExists {
			daysRemaining := GetValidDaysRemaining(GetCertFromDisk(getPublicCertPath(certificateConfig.Path)))

			if daysRemaining <= int64(certificateConfig.RenewThresholdDays) {
				GenerateSSLCert(certificateConfig)
			}
		} else {
			os.MkdirAll(certificateConfig.Path,0750)
			GenerateSSLCert(certificateConfig)
		}
	}
}
