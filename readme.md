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
make build          # builds for your current OS/arch
# or
make compile        # builds for Linux, macOS, and Windows (outputs to ./bin/)
```

The binary (`ssl-manager` or `ssl-manager.exe`) is written to ./bin/.


## Configuration
Default config file path is /etc/ssl-manager/ssl-manager.yaml.
Example configuration files are in example-config.toml and example-config.yaml.

## Usage  
The most important argument is `--help`

`--config` to choose a configuration file. `--check-config` validates the file and prints the parsed configuration as JSON.
`--gen-cas` generates configured CAs, and `--renew-certs` checks and renews leaf certificates.
`--force` regenerates certificates that would otherwise be considered current (be aware, it can overwrite your CA certificates, it is better to delete/move the specific CA you ned to renew and run with `--gen-cas`).
`--daemon` performs an immediate renewal check at startup and repeats it at the configured `Daemon.RenewIntervalDays`.

**I am just a beginner in Go. Every pull request is welcome.**
