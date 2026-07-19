package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"ssl-manager/certificates"
	"ssl-manager/config"
	notification "ssl-manager/notifications"
)

var (
	// overridden by -ldflags
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	argConfFilePtr := flag.String("config", "/etc/ssl-manager/conf.yaml", "Config file to be loaded on the start of the program (can be json, toml or yaml)")
	argGenCAPtr := flag.Bool("gen-ca", false, "Generates certification authority certificate and stores it on the disk")
	argRenewCertsPtr := flag.Bool("renew-certs", false, "Creates missing certificates and renews certificates that will expire soon")
	argForcePtr := flag.Bool("force", false, "Forces certificate generation, even when certificates already exists")
	argVersionPtr := flag.Bool("version", false, "Print version information and exit")
	argCheckConfigPtr := flag.Bool("check-config", false, "Checks config file passed in config argument")
	argDaemonPtr := flag.Bool("daemon", false, "ssl-manager runs as daemon and renews certificates automaticaly, other flags than config are ignored")
	flag.Parse()

	if *argVersionPtr {
		fmt.Printf("ssl-manager %s (commit %s, built %s)\n", Version, Commit, BuildTime)
		os.Exit(0)
	}

	var err error
	err = config.LoadAppConfig(*argConfFilePtr)
	if err != nil {
		fmt.Println("Failed to load config")
		panic(err)
	}
	appConfig, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	if *argCheckConfigPtr {
		err := config.LoadAppConfig(*argConfFilePtr)
		if err != nil {
			panic(fmt.Errorf("error configuration invalid %s", err))
		}

		fmt.Printf("Configuration is valid\n\n")
		yaml, err := appConfig.GetYaml()
		if err != nil {
			panic(fmt.Errorf("error printing yaml: %s", err))
		}

		fmt.Println(yaml)

		os.Exit(0)
	}

	if *argDaemonPtr {
		err = runDaemon()
		if err != nil {
			panic(err)
		} else {
			os.Exit(0)
		}
	}

	if *argGenCAPtr {
		err := os.MkdirAll(appConfig.CACertificate.Path, 0750)
		if err != nil {
			panic(err)
		}
		certExists := appConfig.CACertificate.CertificateExists()
		if (!certExists) || (certExists && *argForcePtr) {
			err = certificates.GenerateCACert(&appConfig.CACertificate)
			if err != nil {
				panic(fmt.Errorf("error generating CA certificate: %s", err))
			}
		} else {
			panic("CA Certificate already exists, if u want to overwrite the old one use argument force")
		}
	}

	if *argRenewCertsPtr {
		_, err = renewCerts(*argForcePtr)
		if err != nil {
			panic(fmt.Errorf("error renewing certificates: %s", err))
		}
	}

	if !*argGenCAPtr && !*argRenewCertsPtr {
		flag.PrintDefaults()
	}
}

func renewCerts(force bool) ([]string, error) {
	appConfig, err := config.GetConfig()
	renewedCerts := make([]string, 0)
	if err != nil {
		return renewedCerts, err
	}

	for i := 0; i < len(appConfig.Certificates); i++ {
		certificateConfig := &appConfig.Certificates[i]

		certExists := certificateConfig.CertificateExists()
		if !force && certExists {
			daysRemaining, err := certificates.GetValidDaysRemaining(certificateConfig)
			if err != nil {
				return renewedCerts, fmt.Errorf("error geting certificate %s validity: %s", certificateConfig.Name, err)
			}

			if int64(certificateConfig.RenewThresholdDays) > daysRemaining {

				//renewing certificate if it is the time
				err = certificates.GenerateSSLCert(certificateConfig, &appConfig.CACertificate)
				if err != nil {
					return renewedCerts, fmt.Errorf("error renewing certificate %s: %s", certificateConfig.Name, err)
				} else {
					fmt.Printf("%s: renewed successfully\n", certificateConfig.Name)
					renewedCerts = append(renewedCerts, certificateConfig.Name)
				}

			} else {
				fmt.Printf("\"%s\": not renewing, expires in %d days\n", certificateConfig.Name, int(daysRemaining))
			}

			continue

		} else {
			err := os.MkdirAll(certificateConfig.Path, 0755)
			if err != nil && err != os.ErrExist {
				return renewedCerts, err
			}

			//renewing certificate even if it exist and it is not its time yet
			err = certificates.GenerateSSLCert(certificateConfig, &appConfig.CACertificate)
			if err != nil {
				return renewedCerts, fmt.Errorf("error renewing certificate %s: %s", certificateConfig.Name, err)

			} else {
				fmt.Printf("\"%s\": generated successfully\n", certificateConfig.Name)
				renewedCerts = append(renewedCerts, certificateConfig.Name)

			}

			continue
		}

	}
	return renewedCerts, nil
}

func runDaemon() error {
	appConfig, err := config.GetConfig()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(time.Duration(appConfig.Daemon.RenewIntervalDays) * 24 * time.Hour)
	defer ticker.Stop()

	for {
		renewedCerts, certRenewErr := renewCerts(false)
		if len(renewedCerts) > 0 || certRenewErr != nil {
			err := notification.SendCertRenewNotifications(appConfig.Daemon.NotificationWebhooks, renewedCerts, certRenewErr)
			if certRenewErr != nil {
				log.Fatalln(err)
			}
			if err != nil {
				log.Fatalln(err)
			}
		}

		<-ticker.C
	}

}
