package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

func GenerateCACert(certConfig *CertificateConfig) error {
	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return err
	}

	err = SaveKeyToDisk(certConfig.GetKeyPath(), priv)
	if err != nil {
		return err
	}

	notBefore, notAfter := getValidFromAfter(certConfig)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	certTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{certConfig.OrganizationName},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageAny},
		BasicConstraintsValid: true,
		IsCA:                  true,
		SignatureAlgorithm:    x509.ECDSAWithSHA512,
		PublicKeyAlgorithm:    x509.ECDSA,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	err = SaveCertToDisk(certConfig.GetCertPath(), certBytes)
	if err != nil {
		return err
	}

	return nil
}

func GenerateSSLCert(certConfig *CertificateConfig) error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return err
	}
	publicKey := &privateKey.PublicKey

	err = SaveKeyToDisk(certConfig.GetKeyPath(), privateKey)
	if err != nil {
		return err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	notBefore, notAfter := getValidFromAfter(certConfig)

	ipAddresses, err := certConfig.GetIPAdresses()
	if err != nil {
		return err
	}

	certTemplate := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{certConfig.OrganizationName},
		},
		DNSNames:           certConfig.DNSNames,
		IPAddresses:        ipAddresses,
		NotBefore:          notBefore,
		NotAfter:           notAfter,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:           x509.KeyUsageDigitalSignature,
		SignatureAlgorithm: x509.ECDSAWithSHA512,
		PublicKeyAlgorithm: x509.ECDSA,
		IsCA:               false,
	}

	CAPrivateBytes, err := GetKeyFromDisk(Config.CACertificate.GetKeyPath())
	if err != nil {
		return err
	}

	CACert, err := GetCertFromDisk(Config.CACertificate.GetCertPath())
	if err != nil {
		return fmt.Errorf("error reading CA certificate from disk: %s", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, CACert, publicKey, CAPrivateBytes)
	if err != nil {
		return err
	}

	err = SaveCertToDisk(certConfig.GetCertPath(), certBytes)
	if err != nil {
		return err
	}

	return nil
}

func getPemBlockFromKey(privateKey *ecdsa.PrivateKey) (*pem.Block, error) {
	bytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return &pem.Block{}, err
	}
	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: bytes}, nil
}

func GetCertFromDisk(path string) (*x509.Certificate, error) {
	bytes, err := os.ReadFile(path)
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

func GetKeyFromDisk(path string) (*ecdsa.PrivateKey, error) {
	privatePEMBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	privateBlock, _ := pem.Decode(privatePEMBytes)
	if privateBlock == nil || privateBlock.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("error decoding key from disk: %s", path)
	}

	privateKey, err := x509.ParseECPrivateKey(privateBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func SaveCertToDisk(path string, certBytes []byte) error {
	certOut, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return err
	}

	defer certOut.Close()
	return nil
}

func SaveKeyToDisk(path string, privateKey *ecdsa.PrivateKey) error {
	privateKeyFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}

	privateKeyPemBlock, err := getPemBlockFromKey(privateKey)
	if err != nil {
		return err
	}
	pem.Encode(privateKeyFile, privateKeyPemBlock)
	defer privateKeyFile.Close()

	return nil
}

func GetValidDaysRemaining(certConfig *CertificateConfig) (int64, error) {
	// 86400 is 1 day in seconds
	cert, err := GetCertFromDisk(certConfig.GetCertPath())
	if err != nil {
		return 0, fmt.Errorf("error reading certificate from disk: %s", err)
	}

	return (cert.NotAfter.Unix() - time.Now().Unix()) / 86400, nil
}

func getValidFromAfter(certConfig *CertificateConfig) (time.Time, time.Time) {
	validFrom := time.Now()
	validTo := validFrom.Add(time.Duration(certConfig.ValidDays) * 24 * time.Hour)
	return validFrom, validTo
}
