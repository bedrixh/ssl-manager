package main

import (
	"errors"
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

// cli flags
var (
	argHelpPtr     *bool
	argConfFilePtr *string
	argForcePtr    *bool
)

// cli subcommands options
var (
	subcommandOptions map[string]subcommandOption = map[string]subcommandOption{
		"version": {
			helpMessage: "Shows ssl-manager version",
			function:    cliVersion,
			loadConfig:  false,
		},
		"daemon": {
			helpMessage: "Starts daemon mode",
			function:    cliDaemon,
			loadConfig:  true,
		},
		"check-config": {
			helpMessage: "Starts daemon mode",
			function:    cliCheckConfig,
			loadConfig:  true,
		},
		"gen-ca": {
			helpMessage: "Generates CA certificates",
			function:    cliGenCA,
			loadConfig:  true,
		},
		"renew-certs": {
			helpMessage: "Starts daemon mode",
			function:    cliRenewCerts,
			loadConfig:  true,
		},
	}
)

type subcommandOption struct {
	helpMessage string
	loadConfig  bool
	function    func() error
}

func main() {
	if len(os.Args) < 2 {
		cliUsage()
		os.Exit(1)
	}
	argHelpPtr = flag.Bool("help", false, "Shows this help message")
	argConfFilePtr = flag.String("config", "/etc/ssl-manager/ssl-manager.yaml", "Config file to be loaded on the start of the program (can be json, toml or yaml)")
	argForcePtr = flag.Bool("force", false, "Forces certificate generation, even when certificates already exist (only usable in gen-ca and renew-certs)")
	flag.CommandLine.Parse(os.Args[2:])

	subcommand, ok := subcommandOptions[os.Args[1]]
	if !ok {
		cliUsage()
		os.Exit(1)
	}
	// if *argHelpPtr || subcommand == subcommandHelp.name {
	// 	subcommandHelp.function()
	// 	os.Exit(0)
	// }

	if subcommand.loadConfig {
		var err error
		err = config.LoadAppConfig(*argConfFilePtr)
		if err != nil {
			switch {
			case errors.Is(err, os.ErrNotExist):
				log.Fatalf("Config file \"%s\" does not exist.", *argConfFilePtr)
			case errors.Is(err, os.ErrPermission):
				log.Fatalf("Config file \"%s\" incorect permissions.", *argConfFilePtr)
			default:
				log.Fatalln(err.Error())
			}

		}
		_, err = config.GetConfig()
		if err != nil {
			log.Fatalln(err)
		}
	}
	err := subcommand.function()
	if err != nil {
		log.Fatalln(err)
	}

}

func renewCerts(force bool) ([]string, error) {
	appConfig, err := config.GetConfig()
	renewedCerts := make([]string, 0)
	if err != nil {
		return renewedCerts, err
	}
	var returnErrs []error

	for i := 0; i < len(appConfig.Certificates); i++ {
		certificateConfig := &appConfig.Certificates[i]
		err := os.MkdirAll(certificateConfig.Path, 0755)
		if err != nil && !errors.Is(err, os.ErrExist) {
			errStringDetails := fmt.Errorf("cannot create folder, for certificate %s: %w", certificateConfig.Name, err)
			log.Println(errStringDetails)
			returnErrs = append(returnErrs, errStringDetails)
			continue
		}

		isOkay, err := certificateConfig.IsOkay()
		if err != nil {
			errStringDetail := fmt.Errorf("error getting certificate %s validity: %w", certificateConfig.Name, err)
			log.Println(errStringDetail)
			returnErrs = append(returnErrs, fmt.Errorf("%s", errStringDetail))
			continue
		}

		if isOkay == false {
			//renewing certificate if it is not okay
			err = certificates.GenerateSSLCert(certificateConfig, appConfig.GetCACertConfigByName(certificateConfig.CACertName))
			if err != nil {
				errStringDetail := fmt.Errorf("error renewing certificate %s: %w", certificateConfig.Name, err)
				log.Println(errStringDetail)
				returnErrs = append(returnErrs, errStringDetail)
			} else {
				log.Printf("%s: renewed successfully\n", certificateConfig.Name)
				renewedCerts = append(renewedCerts, certificateConfig.Name)
			}

		} else {
			if force {
				//renewing certificate even if it is still valid
				err = certificates.GenerateSSLCert(certificateConfig, appConfig.GetCACertConfigByName(certificateConfig.CACertName))
				if err != nil {
					errWithDetail := fmt.Errorf("error renewing certificate %s: %w", certificateConfig.Name, err)

					log.Println(errWithDetail)
					returnErrs = append(returnErrs, errWithDetail)

				} else {
					log.Printf("\"%s\": generated successfully\n", certificateConfig.Name)
					renewedCerts = append(renewedCerts, certificateConfig.Name)

				}
			} else {
				remainingDays, err := certificateConfig.GetValidDaysRemaining()
				if err != nil {
					errStringDetail := fmt.Errorf("error geting remaining days of certficate %s:%w", certificateConfig.Name, err)
					log.Println(errStringDetail)
					returnErrs = append(returnErrs, errStringDetail)
					remainingDays = -1
				}
				log.Printf("\"%s\": not renewing, expires in %d days\n", certificateConfig.Name, int(remainingDays))
			}

		}

	}
	returnErr := errors.Join(returnErrs...)
	return renewedCerts, returnErr
}

func cliDaemon() error {
	appConfig, err := config.GetConfig()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(time.Duration(appConfig.Daemon.RenewIntervalDays) * 24 * time.Hour)
	defer ticker.Stop()

	for {
		renewedCerts, certRenewErr := renewCerts(false)

		// Send notification to user if any certificate is renewed or error ocures
		if len(renewedCerts) > 0 || certRenewErr != nil {
			err := notification.SendCertRenewNotifications(appConfig.Daemon.NotificationWebhooks, renewedCerts, certRenewErr)
			if err != nil {
				log.Println(err)
			}
		}

		if certRenewErr != nil {
			log.Println(certRenewErr)
		}

		<-ticker.C
	}

}

func cliUsage() error {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [command] [flags]\n\n", os.Args[0])

	fmt.Fprintln(flag.CommandLine.Output(), "Subcommands:")
	for i, subcommand := range subcommandOptions {
		fmt.Fprintf(flag.CommandLine.Output(), "  %s		%s\n", i, subcommand.helpMessage)
	}
	fmt.Fprintln(flag.CommandLine.Output(), "\nFlags:")
	flag.PrintDefaults()
	return nil
}

func cliVersion() error {
	fmt.Printf("ssl-manager %s (commit %s, built %s)\n", Version, Commit, BuildTime)
	return nil
}

func cliCheckConfig() error {
	err := config.LoadAppConfig(*argConfFilePtr)
	if err != nil {
		return fmt.Errorf("error configuration invalid (%s)", err)

	} else {

		fmt.Printf("Configuration is valid.\n\n")
		// JSON seems like the easiest-to-read format; it is not easy to write, but for this purpose it seems best to me.
		json, err := config.GetConfigNoErr().GetFormattedJson()
		if err != nil {
			return fmt.Errorf("error rendering json: %w", err)
		}

		fmt.Println(json)
	}
	return nil
}

func cliGenCA() error {
	for i := range len(config.GetConfigNoErr().CACertificates) {
		err := os.MkdirAll(config.GetConfigNoErr().CACertificates[i].Path, os.FileMode(0775))
		if err != nil {
			return err
		}
		certExists := config.GetConfigNoErr().CACertificates[i].CertificateExists()
		if (!certExists) || (certExists && *argForcePtr) {
			err = certificates.GenerateCACert(&config.GetConfigNoErr().CACertificates[i])
			if err != nil {
				log.Printf("error generating CA certificate %s: %s", config.GetConfigNoErr().CACertificates[i].Name, err)
			} else {
				log.Printf("CA Certificate %s generated successfully.", config.GetConfigNoErr().CACertificates[i].Name)
			}
		} else {
			log.Printf("CA Certificate %s already exists, if you want to overwrite the old one use the --force argument.", config.GetConfigNoErr().CACertificates[i].Name)
		}
	}
	return nil
}

func cliRenewCerts() error {
	_, err := renewCerts(*argForcePtr)
	if err != nil {
		return fmt.Errorf("error renewing certificates: %w", err)
	}
	return nil
}
