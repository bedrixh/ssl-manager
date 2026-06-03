package main

import (
	"flag"
	"fmt"

	"github.com/BurntSushi/toml"
)

var Config *Configuration

func main() {
	argConfFilePtr := flag.String("config", "./conf.toml", "Config file to be loaded on the start of the program (can be json or toml)")
	argGenCAPtr := flag.Bool("genca", false, "Generates certification authority certificate and stores it on the disk")
	flag.Parse()

	Config = LoadConfig(*argConfFilePtr)
	bytes, _ := toml.Marshal(Config)
	fmt.Println(string(bytes))
	fmt.Println(GetValidDaysRemaining(GetCertFromDisk("./CA/cert.pem")))

	if *argGenCAPtr == true {
		GenerateCACert(Config)
	}

}
