package main

import (
	"flag"
	"os"
)

var Config *Configuration

func main() {
	argConfFilePtr := flag.String("config", "./conf.toml", "Config file to be loaded on the start of the program (can be json or toml)")
	argGenCAPtr := flag.Bool("gen-ca", false, "Generates certification authority certificate and stores it on the disk")
	argRenewCerts := flag.Bool("renew-certs", false, "Creates missing certificates and renews certificates that will expire soon")
	flag.Parse()

	Config = LoadConfig(*argConfFilePtr)

	if *argGenCAPtr == true {
		os.MkdirAll(Config.CACertificate.Path,0750)
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
