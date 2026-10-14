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

	"github.com/fatih/color"
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

// ENV variables name specification
const (
	envConfigPrefix   = "SSL_MANAGER_"
	envConfigFileName = envConfigPrefix + "CONFIG_FILE"
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
			helpMessage: "Validates configuration file",
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
	argHelpPtr = flag.Bool("help", false, "Shows this help message")
	argConfFilePtr = flag.String("config", "", "Config file to be loaded on the start of the program (can be json, toml or yaml)")
	argForcePtr = flag.Bool("force", false, "Forces certificate generation, even when certificates already exist (only usable in gen-ca and renew-certs)")

	if os.Args[1] == "completions" {
		if os.Args[2] == "fish" {
			cliCompletionsFish()
		}
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		cliUsage()
		os.Exit(1)
	}

	flag.CommandLine.Parse(os.Args[2:])

	if envConfigFile, envConfigFileExists := os.LookupEnv(envConfigFileName); *argConfFilePtr == "" && envConfigFileExists {
		*argConfFilePtr = envConfigFile
	}

	subcommand, ok := subcommandOptions[os.Args[1]]
	if !ok {
		cliUsage()
		os.Exit(1)
	}

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

		if isOkay == false || force {
			//renewing certificate if it is not okay
			err = certificates.GenerateSSLCert(certificateConfig, appConfig.GetCACertConfigByName(certificateConfig.CACertName))
			if err != nil {
				errStringDetail := fmt.Errorf("error generating certificate %s: %w", certificateConfig.Name, err)
				log.Println(errStringDetail)
				returnErrs = append(returnErrs, errStringDetail)
			} else {
				log.Printf("%s: generated successfully\n", certificateConfig.Name)
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
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [subcommand] [flags]\n\n", color.HiBlueString(os.Args[0]))

	fmt.Fprintln(flag.CommandLine.Output(), "Subcommands:")
	for i, subcommand := range subcommandOptions {
		fmt.Fprintf(flag.CommandLine.Output(), "  %s\n\t%s\n", color.HiGreenString(i), color.RedString(subcommand.helpMessage))
	}
	fmt.Fprintln(flag.CommandLine.Output(), "\nFlags:")
	flag.PrintDefaults()
	return nil
}

func cliVersion() error {
	fmt.Printf("%s %s (Commit: %s, Build Date: %s)\n", color.HiBlueString("ssl-manager"), color.HiGreenString(Version), Commit, BuildTime)
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

func cliCompletionsFish() error {
	for subCommand, option := range subcommandOptions {
		fmt.Printf("complete -c ssl-manager -f -n '__fish_use_subcommand' -a %s -d '%s'\n", subCommand, option.helpMessage)
	}

	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("complete -c ssl-manager -s h -l %s -d '%s'\n", f.Name, f.Usage)
	})

	return nil
}
