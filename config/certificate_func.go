package config

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

func (c *CertificateConfig) GetCertFromDisk() (*x509.Certificate, error) {
	bytes, err := os.ReadFile(c.GetCertPath())
	if err != nil {
		return nil, err
	}

	certBlock, _ := pem.Decode(bytes)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("error decoding cert from disk")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (c *CertificateConfig) GetKeyFromDisk() (*ecdsa.PrivateKey, error) {
	privatePEMBytes, err := os.ReadFile(c.GetKeyPath())
	if err != nil {
		return nil, err
	}
	privateBlock, _ := pem.Decode(privatePEMBytes)
	if privateBlock == nil || privateBlock.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("error decoding key from disk")
	}

	privateKey, err := x509.ParseECPrivateKey(privateBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func (c *CertificateConfig) SaveCertToDisk(certBytes []byte) error {
	publicCertFile, err := os.OpenFile(c.GetCertPath(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(c.CertPermissions))
	if err != nil {
		return err
	}
	defer publicCertFile.Close()

	err = pem.Encode(publicCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return err
	}

	err = publicCertFile.Chmod(os.FileMode(c.CertPermissions))
	if err != nil {
		return err
	}

	return nil
}

func (c *CertificateConfig) SaveKeyToDisk(privateKey *ecdsa.PrivateKey) error {
	privateKeyFile, err := os.OpenFile(c.GetKeyPath(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(c.KeyPermissions))
	if err != nil {
		return err
	}
	defer privateKeyFile.Close()

	privateKeyPemBlock, err := getPemBlockFromKey(privateKey)
	if err != nil {
		return err
	}
	err = pem.Encode(privateKeyFile, privateKeyPemBlock)
	if err != nil {
		return err
	}

	err = privateKeyFile.Chmod(os.FileMode(c.KeyPermissions))
	if err != nil {
		return err
	}

	return nil
}

func (c *CertificateConfig) GetValidDaysRemaining() (int64, error) {
	cert, err := c.GetCertFromDisk()
	if err != nil {
		return -1, fmt.Errorf("error reading certificate from disk: %s", err)
	}

	// 86400 is 1 day in seconds
	return (cert.NotAfter.Unix() - time.Now().Unix()) / 86400, nil
}

func (c *CertificateConfig) GetValidFromTo() (time.Time, time.Time) {
	validFrom := time.Now()
	validTo := validFrom.Add(time.Duration(c.ValidDays) * 24 * time.Hour)
	return validFrom, validTo
}

func getPemBlockFromKey(privateKey *ecdsa.PrivateKey) (*pem.Block, error) {
	bytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return &pem.Block{}, err
	}
	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: bytes}, nil
}
