# to use the cli tool you can use arguments
```bash
  -config string
    	Config file to be loaded on the start of the program (can be json, toml or yaml) (default "./conf.toml")
  -force
    	Forces certificate generation, even when certificates already exists
  -gen-ca
    	Generates certification authority certificate and stores it on the disk
  -renew-certs
    	Creates missing certificates and renews certificates that will expire soon

```

# when creating config file be aware that config files are not yet validated. It is possible to use TOML, YAML or JSON, whitchever fits your need the best
```toml
[CACertificate]
Name = "CA"
Path = "./CA2/"
OrganizationName = "John Doe"
CommonName = "John Doe"
Email = "johndoe@example.com""
ValidDays = 79000
RenewThresholdDays = 30

[[Certificates]]
Name = "service-api2"
Path = "./service-api2"
IPs = ["10.0.0.5"]
DNSNames = ["api.example.com"]

[[Certificates]]
Name = "web-frontend2"
Path = "./web-frontend2"
IPs = []
DNSNames = ["www.example.com", "example.com"]

[CertificatesDefaults]
ValidDays = 20
RenewThresholdDays = 7
OrganizationName = "John Doe"
CommonName = "John Doe"
Email = "johndoe@example.com"
```
