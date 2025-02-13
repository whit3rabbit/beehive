package setup

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// SSLConfig holds configuration for SSL certificate generation
type SSLConfig struct {
	Organization     string
	ValidityDays     int
	KeySize          int
	CertificateFile  string
	PrivateKeyFile   string
	AdditionalHosts  []string
	AdditionalIPs    []string
}

// DefaultSSLConfig returns the default SSL configuration
func DefaultSSLConfig(certFile, keyFile string) *SSLConfig {
	return &SSLConfig{
		Organization:    "Manager",
		ValidityDays:   365,
		KeySize:        2048,
		CertificateFile: certFile,
		PrivateKeyFile:  keyFile,
		AdditionalHosts: []string{},
		AdditionalIPs:   []string{},
	}
}

// GenerateSSLCertificate generates a self-signed SSL certificate
func GenerateSSLCertificate(config *Config) error {
	sslConfig := DefaultSSLConfig(config.TLSCertFile, config.TLSKeyFile)
	return generateCertificate(sslConfig)
}

// generateCertificate handles the actual certificate generation process
func generateCertificate(config *SSLConfig) error {
	// Create certificate directory if it doesn't exist
	certDir := filepath.Dir(config.CertificateFile)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, config.KeySize)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Prepare certificate template
	template, err := createCertificateTemplate(config)
	if err != nil {
		return fmt.Errorf("failed to create certificate template: %w", err)
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		template,
		template,
		&privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Write certificate to file
	if err := writeCertificateToFile(config.CertificateFile, derBytes); err != nil {
		return fmt.Errorf("failed to write certificate file: %w", err)
	}

	// Write private key to file
	if err := writePrivateKeyToFile(config.PrivateKeyFile, privateKey); err != nil {
		return fmt.Errorf("failed to write private key file: %w", err)
	}

	return nil
}

// createCertificateTemplate creates an x509 certificate template with appropriate settings
func createCertificateTemplate(config *SSLConfig) (*x509.Certificate, error) {
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Prepare IP addresses
	ipAddresses := []net.IP{net.ParseIP("127.0.0.1")}
	for _, ip := range config.AdditionalIPs {
		if parsedIP := net.ParseIP(ip); parsedIP != nil {
			ipAddresses = append(ipAddresses, parsedIP)
		}
	}

	// Prepare DNS names
	dnsNames := []string{"localhost"}
	dnsNames = append(dnsNames, config.AdditionalHosts...)

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{config.Organization},
			CommonName:   "localhost",
		},
		NotBefore:             now,
		NotAfter:              now.Add(time.Duration(config.ValidityDays) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           ipAddresses,
		DNSNames:             dnsNames,
	}

	return template, nil
}

// generateSerialNumber generates a random serial number for the certificate
func generateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}
	return serialNumber, nil
}

// writeCertificateToFile writes the certificate to a PEM file
func writeCertificateToFile(filename string, derBytes []byte) error {
	certOut, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}); err != nil {
		return fmt.Errorf("failed to write certificate file: %w", err)
	}

	return nil
}

// writePrivateKeyToFile writes the private key to a PEM file
func writePrivateKeyToFile(filename string, privateKey *rsa.PrivateKey) error {
	keyOut, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer keyOut.Close()

	if err := pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		return fmt.Errorf("failed to write private key file: %w", err)
	}

	return nil
}

// ValidateSSLCertificate checks if the certificate and key files exist and are valid
func ValidateSSLCertificate(config *Config) error {
	// Check certificate file
	certData, err := os.ReadFile(config.TLSCertFile)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Parse certificate
	certBlock, _ := pem.Decode(certData)
	if certBlock == nil {
		return fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check private key file
	keyData, err := os.ReadFile(config.TLSKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}

	// Parse private key
	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil {
		return fmt.Errorf("failed to decode private key PEM")
	}

	_, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	// Check certificate expiration
	if time.Now().After(cert.NotAfter) {
		return fmt.Errorf("certificate has expired")
	}

	return nil
}

// AddHostToSSLCertificate regenerates the SSL certificate with an additional host
func AddHostToSSLCertificate(config *Config, hostname string) error {
	sslConfig := DefaultSSLConfig(config.TLSCertFile, config.TLSKeyFile)
	sslConfig.AdditionalHosts = append(sslConfig.AdditionalHosts, hostname)
	return generateCertificate(sslConfig)
}

// AddIPToSSLCertificate regenerates the SSL certificate with an additional IP address
func AddIPToSSLCertificate(config *Config, ip string) error {
	sslConfig := DefaultSSLConfig(config.TLSCertFile, config.TLSKeyFile)
	sslConfig.AdditionalIPs = append(sslConfig.AdditionalIPs, ip)
	return generateCertificate(sslConfig)
}