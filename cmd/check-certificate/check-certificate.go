package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/thomis/sensu-plugins-go/pkg/check"
)

// Config holds the configuration for the certificate check
type Config struct {
	Host    string
	Port    int
	Timeout int64
	Expiry  int64
}

// CertificateChecker performs TLS certificate validation
type CertificateChecker struct {
	config    Config
	dialer    *net.Dialer
	tlsConfig *tls.Config
}

// NewCertificateChecker creates a checker with default TLS config
func NewCertificateChecker(cfg Config) *CertificateChecker {
	return &CertificateChecker{
		config:    cfg,
		dialer:    &net.Dialer{Timeout: time.Duration(cfg.Timeout) * time.Second},
		tlsConfig: nil, // Use default TLS config for production
	}
}

// NewCertificateCheckerWithTLSConfig creates a checker with custom TLS config (for testing)
func NewCertificateCheckerWithTLSConfig(cfg Config, tlsConfig *tls.Config) *CertificateChecker {
	return &CertificateChecker{
		config:    cfg,
		dialer:    &net.Dialer{Timeout: time.Duration(cfg.Timeout) * time.Second},
		tlsConfig: tlsConfig,
	}
}

// Connect establishes a TLS connection to the target host
func (cc *CertificateChecker) Connect() (*tls.Conn, error) {
	address := cc.config.Host + ":" + strconv.Itoa(cc.config.Port)
	conn, err := tls.DialWithDialer(cc.dialer, "tcp", address, cc.tlsConfig)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// ValidateCertificate checks if the certificate is valid
func (cc *CertificateChecker) ValidateCertificate(conn *tls.Conn) (*x509.Certificate, error) {
	// Get connection state
	state := conn.ConnectionState()

	// Check if we have certificates
	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("unable to find or retrieve certificates")
	}

	cert := state.PeerCertificates[0]

	// check if hostname matches with certificate (only if not using InsecureSkipVerify)
	if cc.tlsConfig == nil || !cc.tlsConfig.InsecureSkipVerify {
		err := conn.VerifyHostname(cc.config.Host)
		if err != nil {
			return nil, err
		}
	} else {
		// Manual hostname verification for testing
		err := cert.VerifyHostname(cc.config.Host)
		if err != nil {
			return nil, err
		}
	}

	now := time.Now()

	// check date validity
	if now.Before(cert.NotBefore) {
		return nil, fmt.Errorf("certificate not before: %s UTC", cert.NotBefore.Format("2006-01-02 15:04:05"))
	}

	if now.After(cert.NotAfter) {
		return nil, fmt.Errorf("certificate not after: %s UTC", cert.NotAfter.Format("2006-01-02 15:04:05"))
	}

	return cert, nil
}

// CheckExpiry checks if the certificate is expiring soon
func (cc *CertificateChecker) CheckExpiry(cert *x509.Certificate) (bool, float64, string) {
	now := time.Now()
	days_left := cert.NotAfter.Sub(now).Hours() / 24

	if now.Add(time.Duration(cc.config.Expiry) * 24 * time.Hour).After(cert.NotAfter) {
		info := fmt.Sprintf("Certificate about to expire in less than %d days. Not After: %s (%0.1f days left)",
			cc.config.Expiry, cert.NotAfter.Format("2006-01-02 15:04:05")+" UTC",
			days_left)
		return true, days_left, info
	}

	return false, days_left, ""
}

// FormatOutput formats the certificate information for display
func (cc *CertificateChecker) FormatOutput(cert *x509.Certificate, days_left float64) string {
	address := cc.config.Host + ":" + strconv.Itoa(cc.config.Port)
	output := address + "\n\n"
	output += fmt.Sprintf("Issuer Name: %s\n", cert.Issuer)
	output += fmt.Sprintf("Not Before : %s\n",
		cert.NotBefore.Format("2006-01-02 15:04:05")+" UTC")
	output += fmt.Sprintf("Not After  : %s (%0.1f days left)\n",
		cert.NotAfter.Format("2006-01-02 15:04:05")+" UTC",
		days_left)
	output += fmt.Sprintf("Common Name: %s\n", cert.Issuer.CommonName)
	output += fmt.Sprintf("DNS Names  : %s\n", strings.Join(cert.DNSNames, ", "))
	return output
}

// Run executes the certificate check using the provided check instance for reporting
func (cc *CertificateChecker) Run(c *check.CheckStruct) {
	conn, err := cc.Connect()
	if err != nil {
		c.Error(err)
	}
	defer conn.Close()

	cert, err := cc.ValidateCertificate(conn)
	if err != nil {
		c.Error(err)
	}

	expiring, days_left, warning := cc.CheckExpiry(cert)
	if expiring {
		c.Warning(warning)
	}

	output := cc.FormatOutput(cert, days_left)
	c.Ok(output)
}

// SetupOptions configures the command-line flags
func SetupOptions(c *check.CheckStruct) *Config {
	var cfg Config

	c.Option.StringVarP(&cfg.Host, "host", "h", "localhost", "HOST")
	c.Option.IntVarP(&cfg.Port, "port", "P", 443, "PORT")
	c.Option.Int64VarP(&cfg.Timeout, "timeout", "t", 5, "TIMEOUT")
	c.Option.Int64VarP(&cfg.Expiry, "expiry", "e", 30, "EXPIRY warning in days")

	return &cfg
}

func main() {
	// Step 1: Create the check instance (handles Sensu lifecycle)
	c := check.New("CheckCertificate")

	// Step 2: Setup command-line options
	config := SetupOptions(c)

	// Step 3: Parse the command-line arguments
	c.Init()

	// Step 4: Create the certificate checker with the configuration
	checker := NewCertificateChecker(*config)

	// Step 5: Run the check and report results via the check instance
	checker.Run(c)
}
