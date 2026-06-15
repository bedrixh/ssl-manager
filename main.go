package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var (
	// overridden by -ldflags
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

var Config *Configuration

func main() {
	argConfFilePtr := flag.String("config", "/etc/ssl-manager/conf.yaml", "Config file to be loaded on the start of the program (can be json, toml or yaml)")
	argGenCAPtr := flag.Bool("gen-ca", false, "Generates certification authority certificate and stores it on the disk")
	argRenewCertsPtr := flag.Bool("renew-certs", false, "Creates missing certificates and renews certificates that will expire soon")
	argForcePtr := flag.Bool("force", false, "Forces certificate generation, even when certificates already exists")
	argVersionPtr := flag.Bool("version", false, "Print version information and exit")
	argDaemonPtr := flag.Bool("daemon", false, "ssl-manager runs as daemon and renews certificates automaticaly, other flags than config are ignored")
	flag.Parse()

	if *argVersionPtr {
		fmt.Printf("ssl-manager %s (commit %s, built %s)\n", Version, Commit, BuildTime)
		os.Exit(0)
	}

	var err error
	Config, err = GetConfig(*argConfFilePtr)
	if err != nil {
		fmt.Println("Failed to load config")
		panic(err)
	}

	if *argDaemonPtr == true {
		err = runDaemon()
		if err != nil {
			panic(err)
		} else {
			os.Exit(0)
		}
	}

	if *argGenCAPtr == true {
		err := os.MkdirAll(Config.CACertificate.Path, 0750)
		if err != nil {
			panic(err)
		}
		certExists := Config.CACertificate.CertificateExists()
		if (certExists == false) || (certExists == true && *argForcePtr == true) {
			err = GenerateCACert(&Config.CACertificate)
			if err != nil {
				panic(fmt.Errorf("Error generating CA certificate (%s)", err))
			}
		} else {
			panic("CA Certificate already exists, if u want to overwrite the old one use argument force")
		}
	}

	if *argRenewCertsPtr == true {
		err = renewCerts(*argForcePtr)
		if err != nil {
			panic(fmt.Errorf("Error renewing certificates (%s)", err))
		}
	}

	if *argGenCAPtr == false && *argRenewCertsPtr == false {
		flag.PrintDefaults()
	}
}

func renewCerts(force bool) error {
	//pretty confusing junk i would like to do something about it

	for i := 0; i < len(Config.Certificates); i++ {
		certificateConfig := &Config.Certificates[i]

		certExists := certificateConfig.CertificateExists()
		if force == false && certExists == true {
			daysRemaining, err := GetValidDaysRemaining(certificateConfig)
			if err != nil {
				return fmt.Errorf("Error geting certificate %s validity (%s)", certificateConfig.Name, err)
			}

			if int64(certificateConfig.RenewThresholdDays) > daysRemaining {

				//renewing certificate if it is the time
				err = GenerateSSLCert(certificateConfig)
				if err != nil {
					return fmt.Errorf("Error renewing certificate %s (%s)", certificateConfig.Name, err)
				} else {
					fmt.Printf("%s: renewed successfully\n", certificateConfig.Name)
				}

			} else {
				fmt.Printf("\"%s\": not renewing, expires in %d days\n", certificateConfig.Name, int(daysRemaining))
			}

			continue

		} else {
			err := os.MkdirAll(certificateConfig.Path, 0755)
			if err != nil && err != os.ErrExist {
				return err
			}

			//renewing certificate even if it exist and it is not his time yet
			err = GenerateSSLCert(certificateConfig)
			if err != nil {
				return fmt.Errorf("Error renewing certificate %s (%s)", certificateConfig.Name, err)

			} else {
				fmt.Printf("\"%s\": generated successfully\n", certificateConfig.Name)
			}

			continue
		}

	}
	return nil
}

func runDaemon() error {
	ticker := time.NewTicker(time.Duration(Config.Daemon.RenewIntervalDays) * 24 * time.Hour)
	defer ticker.Stop()

	for true {
		err := renewCerts(false)
		if err != nil {
			return err
		}

		<-ticker.C
	}

	return nil
}
