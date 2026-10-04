package certificates

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"

	"ssl-manager/config"
)

func GenerateCACert(certConfig *config.CertificateConfig) error {
	if certConfig == nil {
		return fmt.Errorf("GenerateCACert cannot accept nil CertConfig")
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return err
	}

	notBefore, notAfter := certConfig.GetValidFromTo()

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

	certBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	err = certConfig.SaveKeyToDisk(privateKey)
	if err != nil {
		return err
	}

	err = certConfig.SaveCertToDisk(certBytes)
	if err != nil {
		return err
	}

	return nil
}

func GenerateSSLCert(certConfig *config.CertificateConfig, caCertConfig *config.CertificateConfig) error {
	if certConfig == nil {
		return fmt.Errorf("GenerateSSLCert cannot accept nil CertConfig")
	}

	if caCertConfig == nil {
		return fmt.Errorf("GenerateSSLCert cannot accept nil caCertConfig")
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return err
	}
	publicKey := &privateKey.PublicKey

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	notBefore, notAfter := certConfig.GetValidFromTo()

	ipAddresses, err := certConfig.GetIPAddresses()
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

	CAPrivateBytes, err := caCertConfig.GetKeyFromDisk()
	if err != nil {
		return err
	}

	CACert, err := caCertConfig.GetCertFromDisk()
	if err != nil {
		return fmt.Errorf("error reading CA certificate from disk: %s", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, CACert, publicKey, CAPrivateBytes)
	if err != nil {
		return err
	}

	err = certConfig.SaveKeyToDisk(privateKey)
	if err != nil {
		return err
	}

	err = certConfig.SaveCertToDisk(certBytes)
	if err != nil {
		return err
	}

	return nil
}
