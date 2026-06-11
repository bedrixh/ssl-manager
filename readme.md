# ssl‑manager

**ssl‑manager** – a small command‑line utility for creating and renewing SSL certificates.

## Features
- Reads configuration from json, toml, or yaml (examples are in root directory).  
- Generates self‑signed certificates and renews existing ones.

## Building
Go is required to build the project
Clone the repo and build

```bash
git clone https://github.com/yourusername/ssl-manager.git
cd ssl-manager
make build          # builds for your current OS/arch
# or
make compile        # builds for Linux, macOS, and Windows (outputs to ./bin/)
```

The binary (`ssl-manager` or `ssl-manager.exe`) appears in the ./bin/.


## Configuration
There are example configuration files, conf.json, conf.toml and conf.yaml (not yet, only the example for toml is already done).
Default config file path is /etc/ssl-manager/conf.toml

## Usage  
The most important argument is `--help`:

```bash
./ssl-manager --help
```

**I am just a beginner in Go. Every pull request is welcome.**
