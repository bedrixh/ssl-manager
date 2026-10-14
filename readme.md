# ssl‑manager

**ssl‑manager** – a small command‑line utility for creating and renewing SSL certificates.

## Features
- Reads configuration from json, toml, or yaml (examples are in root directory).  
- Generates self‑signed certificates and renews existing ones.

## Building
Go is required to build the project.
Clone the repo and build using GNU make.

```bash
git clone https://github.com/bedrixh/ssl-manager.git
cd ssl-manager
git switch "$(git tag -l 'v[0-9]*.[0-9]*.[0-9]*' --sort=creatordate | tail -n 1)" --detach
make build          # builds for your current OS/arch
# or
make compile        # builds for Linux, macOS (outputs to ./bin/)
```

The binary (`ssl-manager` or `ssl-manager.exe`) is written to ./bin/.


## Configuration
Default config file path is /etc/ssl-manager/ssl-manager.yaml.
Example configuration files are in example-config.toml and example-config.yaml.

## Usage  
The most important argument is `--help`
```md
Usage of bin/ssl-manager:
  -check-config
    	Checks config file passed in config argument
  -config string
    	Config file to be loaded on the start of the program (can be json, toml or yaml) (default "/etc/ssl-manager/conf.yaml")
  -daemon
    	ssl-manager runs as daemon and renews certificates automaticaly, other flags than config are ignored
  -force
    	Forces certificate generation, even when certificates already exists
  -gen-ca
    	Generates certification authority certificate and stores it on the disk
  -renew-certs
    	Creates missing certificates and renews certificates that will expire soon
  -version
    	Print version information and exit
```
**I am just a beginner in Go. Every pull request is welcome.**
