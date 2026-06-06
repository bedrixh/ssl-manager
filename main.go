package main

import (
	"flag"
	"fmt"
	"os"
)

var Config *Configuration

func main() {
	argConfFilePtr := flag.String("config", "./conf.toml", "Config file to be loaded on the start of the program (can be json or toml)")
	argGenCAPtr := flag.Bool("gen-ca", false, "Generates certification authority certificate and stores it on the disk")
	argRenewCertsPtr := flag.Bool("renew-certs", false, "Creates missing certificates and renews certificates that will expire soon")
	argForcePtr := flag.Bool("force", false, "Forces certificate generation, even when certificates already exists")
	flag.Parse()

	Config = LoadConfig(*argConfFilePtr)

	if *argGenCAPtr == true {
		err := os.MkdirAll(Config.CACertificate.Path, 0750)
		if err != nil{
			panic(err)
		}
		certExists := getCertificateExists(&Config.CACertificate)
		if (certExists == false) || (certExists == true && *argForcePtr == true) {
			GenerateCACert(&Config.CACertificate)
		}else{
			panic("CA Certificate already exists, if u want to overwrite the old one use argument force")
		}
	}

	if *argRenewCertsPtr == true {
		renewCerts(*argForcePtr)
	}

}

func renewCerts(force bool) {
	for i := 0; i < len(Config.Certificates); i++ {
		certificateConfig := &Config.Certificates[i]
		certExists := getCertificateExists(certificateConfig)
		if certExists {
			daysRemaining := GetValidDaysRemaining(GetCertFromDisk(getPublicCertPath(certificateConfig.Path)))
			fmt.Println(certificateConfig.RenewThresholdDays)
			fmt.Println(daysRemaining)
			if int64(certificateConfig.RenewThresholdDays) >= daysRemaining {
				GenerateSSLCert(certificateConfig)
			}
		} else {
			os.MkdirAll(certificateConfig.Path, 0750)
			GenerateSSLCert(certificateConfig)
		}
	}
}
