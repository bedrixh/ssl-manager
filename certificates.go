package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

func getPrivateKeyPath(dirPath string) string {
	return dirPath + "/key.pem"
}
func getPublicCertPath(dirPath string) string {
	return dirPath + "/cert.pem"
}

func GenerateCACert(certConfig *CertificateConfig) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	notBefore, notAfter := getValidFromAfter(certConfig)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		panic(err)
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
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &priv.PublicKey, priv)
	if err != nil {
		panic(err)
	}

	SaveCertToDisk(getPublicCertPath(certConfig.Path), certBytes)

	SaveKeyToDisk(getPrivateKeyPath(certConfig.Path), *priv)
}

func pemBlockFromKey(priv *ecdsa.PrivateKey) *pem.Block {
	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		panic(err)
	}
	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
}

func GenerateSSLCert(cert *CertificateConfig) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	pub := &priv.PublicKey

	SaveKeyToDisk(getPrivateKeyPath(cert.Path), *priv)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		panic(err)
	}

	notBefore, notAfter := getValidFromAfter(cert)

	certTemplate := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:  []string{cert.OrganizationName},
		},
		DNSNames:    cert.DNSNames,
		NotBefore:   notBefore,
		NotAfter:    notAfter,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	CAPrivateBytes := GetKeyFromDisk(getPrivateKeyPath(Config.CACertificate.Path))

	CACert := GetCertFromDisk(getPublicCertPath(Config.CACertificate.Path))

	certBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, &CACert, pub, CAPrivateBytes)
	if err != nil {
		panic(err)
	}
	SaveCertToDisk(getPublicCertPath(cert.Path), certBytes)

}

func GetCertFromDisk(path string) x509.Certificate {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	certBlock, _ := pem.Decode(data)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		panic("error decoding cert from disk")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		panic(err)
	}

	return *cert
}

func GetKeyFromDisk(path string) *ecdsa.PrivateKey {
	privatePEMBytes, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	privateBlock, _ := pem.Decode(privatePEMBytes)
	if privateBlock.Type != "EC PRIVATE KEY" {
		panic("error decoding key from disk")
	}

	PrivateKey, _ := x509.ParseECPrivateKey(privateBlock.Bytes)
	return PrivateKey
}

func SaveCertToDisk(path string, certBytes []byte) {
	certOut, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		panic(err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	defer certOut.Close()
}

func SaveKeyToDisk(path string, privateKey ecdsa.PrivateKey) {
	privateKeyFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		panic(err)
	}
	pem.Encode(privateKeyFile, pemBlockFromKey(&privateKey))
	defer privateKeyFile.Close()
}

func GetValidDaysRemaining(cert x509.Certificate) int64 {
	// 86400 is 1 day in seconds
	return (cert.NotAfter.Unix() - time.Now().Unix()) / 86400
}

func getValidFromAfter(certConfig *CertificateConfig) (time.Time, time.Time) {
	validFrom := time.Now()
	validTo := validFrom.Add(time.Duration(certConfig.ValidDays) * 24 * time.Hour)
	return validFrom, validTo
}

func getCertificateExists(certConfig *CertificateConfig) bool{
	fileInfo, err := os.Stat(getPublicCertPath(certConfig.Path))
	if err != nil {
		return false
	} 
	return fileInfo.IsDir() == false 
}

