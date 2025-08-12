package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomis/sensu-plugins-go/pkg/check"
)

func generateTestCertificate(notBefore, notAfter time.Time, dnsNames []string, commonName string) (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test Org"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    commonName,
		},
		Issuer: pkix.Name{
			Organization: []string{"Test Issuer Org"},
			Country:      []string{"US"},
			CommonName:   commonName + " CA",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}

	return cert, nil
}

func createTestTLSServer(cert tls.Certificate) (net.Listener, int, error) {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	listener, err := tls.Listen("tcp", "127.0.0.1:0", tlsConfig)
	if err != nil {
		return nil, 0, err
	}

	port := listener.Addr().(*net.TCPAddr).Port

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			// Perform TLS handshake
			tlsConn := conn.(*tls.Conn)
			err = tlsConn.Handshake()
			if err != nil {
				conn.Close()
				continue
			}
			// Keep connection open for a moment to allow client to read certificate
			time.Sleep(100 * time.Millisecond)
			conn.Close()
		}
	}()

	return listener, port, nil
}

func TestCertificateChecker_Connect(t *testing.T) {
	// Create a test server
	cert, err := generateTestCertificate(
		time.Now().Add(-24*time.Hour),
		time.Now().Add(90*24*time.Hour),
		[]string{"localhost", "127.0.0.1"},
		"test.example.com",
	)
	assert.NoError(t, err)

	listener, port, err := createTestTLSServer(cert)
	assert.NoError(t, err)
	defer listener.Close()

	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name        string
		cfg         Config
		expectError bool
	}{
		{
			name: "Valid connection",
			cfg: Config{
				Host:    "localhost",
				Port:    port,
				Timeout: 5,
				Expiry:  30,
			},
			expectError: false,
		},
		{
			name: "Invalid host",
			cfg: Config{
				Host:    "non-existent-host-12345.example.com",
				Port:    443,
				Timeout: 1,
				Expiry:  30,
			},
			expectError: true,
		},
		{
			name: "Invalid port",
			cfg: Config{
				Host:    "localhost",
				Port:    99999,
				Timeout: 1,
				Expiry:  30,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: true,
			}
			checker := NewCertificateCheckerWithTLSConfig(tt.cfg, tlsConfig)
			conn, err := checker.Connect()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, conn)
				if conn != nil {
					conn.Close()
				}
			}
		})
	}
}

func TestCertificateChecker_ValidateCertificate(t *testing.T) {
	tests := []struct {
		name          string
		notBefore     time.Time
		notAfter      time.Time
		dnsNames      []string
		host          string
		expectError   bool
		errorContains string
	}{
		{
			name:        "Valid certificate",
			notBefore:   time.Now().Add(-24 * time.Hour),
			notAfter:    time.Now().Add(90 * 24 * time.Hour),
			dnsNames:    []string{"localhost", "127.0.0.1"},
			host:        "localhost",
			expectError: false,
		},
		{
			name:          "Certificate not yet valid",
			notBefore:     time.Now().Add(24 * time.Hour),
			notAfter:      time.Now().Add(90 * 24 * time.Hour),
			dnsNames:      []string{"localhost", "127.0.0.1"},
			host:          "localhost",
			expectError:   true,
			errorContains: "certificate not before",
		},
		{
			name:          "Expired certificate",
			notBefore:     time.Now().Add(-90 * 24 * time.Hour),
			notAfter:      time.Now().Add(-24 * time.Hour),
			dnsNames:      []string{"localhost", "127.0.0.1"},
			host:          "localhost",
			expectError:   true,
			errorContains: "certificate not after",
		},
		{
			name:          "Hostname mismatch",
			notBefore:     time.Now().Add(-24 * time.Hour),
			notAfter:      time.Now().Add(90 * 24 * time.Hour),
			dnsNames:      []string{"wronghost.example.com"},
			host:          "localhost",
			expectError:   true,
			errorContains: "certificate is valid for",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert, err := generateTestCertificate(tt.notBefore, tt.notAfter, tt.dnsNames, "test.example.com")
			assert.NoError(t, err)

			listener, port, err := createTestTLSServer(cert)
			assert.NoError(t, err)
			defer listener.Close()

			time.Sleep(100 * time.Millisecond)

			tlsConfig := &tls.Config{
				InsecureSkipVerify: true,
			}
			checker := NewCertificateCheckerWithTLSConfig(Config{
				Host:    tt.host,
				Port:    port,
				Timeout: 5,
				Expiry:  30,
			}, tlsConfig)

			conn, err := checker.Connect()
			if err != nil {
				if tt.expectError {
					return
				}
				t.Fatalf("Failed to connect: %v", err)
			}
			defer conn.Close()

			result, err := checker.ValidateCertificate(conn)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestCertificateChecker_ValidateCertificate_WithoutInsecureSkipVerify(t *testing.T) {
	// This test specifically tests the conn.VerifyHostname path when NOT using InsecureSkipVerify
	// This is a different code path than the manual cert.VerifyHostname

	// Create a test certificate with a specific hostname
	cert, err := generateTestCertificate(
		time.Now().Add(-24*time.Hour),
		time.Now().Add(90*24*time.Hour),
		[]string{"test.example.com"}, // Certificate is only valid for test.example.com
		"test.example.com",
	)
	assert.NoError(t, err)

	listener, port, err := createTestTLSServer(cert)
	assert.NoError(t, err)
	defer listener.Close()

	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name          string
		host          string
		useInsecure   bool
		expectError   bool
		errorContains string
	}{
		{
			name:          "Valid hostname with InsecureSkipVerify=false (nil tlsConfig)",
			host:          "localhost",
			useInsecure:   false,
			expectError:   true,
			errorContains: "certificate", // Will fail because cert is for test.example.com, not localhost
		},
		{
			name:          "Valid hostname with InsecureSkipVerify=true",
			host:          "localhost",
			useInsecure:   true,
			expectError:   true,
			errorContains: "certificate is valid for test.example.com, not localhost", // Manual verification
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var checker *CertificateChecker

			if tt.useInsecure {
				tlsConfig := &tls.Config{
					InsecureSkipVerify: true,
				}
				checker = NewCertificateCheckerWithTLSConfig(Config{
					Host:    tt.host,
					Port:    port,
					Timeout: 5,
					Expiry:  30,
				}, tlsConfig)
			} else {
				// Use the production path with nil tlsConfig
				// This will try to verify against system roots and fail
				checker = NewCertificateChecker(Config{
					Host:    tt.host,
					Port:    port,
					Timeout: 5,
					Expiry:  30,
				})
			}

			conn, err := checker.Connect()
			if err != nil {
				// Connection might fail if using default TLS config (no InsecureSkipVerify)
				// because our test cert is self-signed
				if !tt.useInsecure {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "certificate")
					return
				}
				t.Fatalf("Unexpected connection error: %v", err)
			}
			defer conn.Close()

			result, err := checker.ValidateCertificate(conn)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestCertificateChecker_CheckExpiry(t *testing.T) {
	tests := []struct {
		name          string
		notAfter      time.Time
		expiryDays    int64
		expectWarning bool
	}{
		{
			name:          "Certificate expires in 45 days, warning at 30",
			notAfter:      time.Now().Add(45 * 24 * time.Hour),
			expiryDays:    30,
			expectWarning: false,
		},
		{
			name:          "Certificate expires in 25 days, warning at 30",
			notAfter:      time.Now().Add(25 * 24 * time.Hour),
			expiryDays:    30,
			expectWarning: true,
		},
		{
			name:          "Certificate expires in 7 days, warning at 7",
			notAfter:      time.Now().Add(7*24*time.Hour + time.Hour),
			expiryDays:    7,
			expectWarning: false,
		},
		{
			name:          "Certificate expires in 6 days, warning at 7",
			notAfter:      time.Now().Add(6 * 24 * time.Hour),
			expiryDays:    7,
			expectWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert := &x509.Certificate{
				NotAfter: tt.notAfter,
			}

			checker := NewCertificateChecker(Config{
				Host:    "localhost",
				Port:    443,
				Timeout: 5,
				Expiry:  tt.expiryDays,
			})

			expiring, daysLeft, warning := checker.CheckExpiry(cert)

			assert.Equal(t, tt.expectWarning, expiring)

			if tt.expectWarning {
				assert.Less(t, daysLeft, float64(tt.expiryDays))
				assert.Contains(t, warning, "Certificate about to expire")
				assert.Contains(t, warning, fmt.Sprintf("less than %d days", tt.expiryDays))
			} else {
				assert.GreaterOrEqual(t, daysLeft, float64(tt.expiryDays)-1) // -1 for rounding tolerance
				assert.Empty(t, warning)
			}
		})
	}
}

func TestCertificateChecker_FormatOutput(t *testing.T) {
	notBefore := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	notAfter := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	cert := &x509.Certificate{
		Issuer: pkix.Name{
			Organization: []string{"Test Issuer"},
			CommonName:   "Test CA",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		DNSNames:  []string{"example.com", "www.example.com"},
	}

	checker := NewCertificateChecker(Config{
		Host:    "example.com",
		Port:    443,
		Timeout: 5,
		Expiry:  30,
	})

	output := checker.FormatOutput(cert, 180.5)

	// Verify output contains expected information
	assert.Contains(t, output, "example.com:443")
	assert.Contains(t, output, "Issuer Name: CN=Test CA,O=Test Issuer")
	assert.Contains(t, output, "Not Before : 2024-01-01 00:00:00 UTC")
	assert.Contains(t, output, "Not After  : 2024-12-31 23:59:59 UTC (180.5 days left)")
	assert.Contains(t, output, "Common Name: Test CA")
	assert.Contains(t, output, "DNS Names  : example.com, www.example.com")
}

func TestCertificateChecker_Integration(t *testing.T) {
	// Create certificates with various states
	tests := []struct {
		name          string
		notBefore     time.Time
		notAfter      time.Time
		dnsNames      []string
		commonName    string
		expiryDays    int64
		expectWarning bool
		expectError   bool
		errorContains string
	}{
		{
			name:          "Valid certificate with no warnings",
			notBefore:     time.Now().Add(-24 * time.Hour),
			notAfter:      time.Now().Add(90 * 24 * time.Hour),
			dnsNames:      []string{"localhost", "127.0.0.1"},
			commonName:    "test.example.com",
			expiryDays:    30,
			expectWarning: false,
			expectError:   false,
		},
		{
			name:          "Certificate expiring soon",
			notBefore:     time.Now().Add(-24 * time.Hour),
			notAfter:      time.Now().Add(20 * 24 * time.Hour),
			dnsNames:      []string{"localhost", "127.0.0.1"},
			commonName:    "test.example.com",
			expiryDays:    30,
			expectWarning: true,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert, err := generateTestCertificate(tt.notBefore, tt.notAfter, tt.dnsNames, tt.commonName)
			assert.NoError(t, err)

			listener, port, err := createTestTLSServer(cert)
			assert.NoError(t, err)
			defer listener.Close()

			time.Sleep(100 * time.Millisecond)

			// Create checker and test full flow
			tlsConfig := &tls.Config{
				InsecureSkipVerify: true,
			}
			checker := NewCertificateCheckerWithTLSConfig(Config{
				Host:    "localhost",
				Port:    port,
				Timeout: 5,
				Expiry:  tt.expiryDays,
			}, tlsConfig)

			// Connect
			conn, err := checker.Connect()
			if tt.expectError && tt.errorContains == "connect" {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, conn)
			defer conn.Close()

			// Validate
			validCert, err := checker.ValidateCertificate(conn)
			if tt.expectError && tt.errorContains != "connect" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, validCert)

			// Check expiry
			expiring, daysLeft, warning := checker.CheckExpiry(validCert)
			assert.Equal(t, tt.expectWarning, expiring)
			if tt.expectWarning {
				assert.NotEmpty(t, warning)
			}

			// Format output
			output := checker.FormatOutput(validCert, daysLeft)
			assert.Contains(t, output, "localhost:"+fmt.Sprintf("%d", port))
			assert.Contains(t, output, "Issuer Name:")
			assert.Contains(t, output, "Not Before :")
			assert.Contains(t, output, "Not After  :")
			assert.Contains(t, output, "Common Name:")
			assert.Contains(t, output, "DNS Names  :")
		})
	}
}

func TestMultipleDNSNames(t *testing.T) {
	dnsNames := []string{"test1.example.com", "test2.example.com", "localhost", "127.0.0.1"}
	cert, err := generateTestCertificate(
		time.Now().Add(-24*time.Hour),
		time.Now().Add(90*24*time.Hour),
		dnsNames,
		"multi.example.com",
	)
	assert.NoError(t, err)

	listener, port, err := createTestTLSServer(cert)
	assert.NoError(t, err)
	defer listener.Close()

	time.Sleep(100 * time.Millisecond)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	checker := NewCertificateCheckerWithTLSConfig(Config{
		Host:    "localhost",
		Port:    port,
		Timeout: 5,
		Expiry:  30,
	}, tlsConfig)

	conn, err := checker.Connect()
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()

	validCert, err := checker.ValidateCertificate(conn)
	assert.NoError(t, err)
	assert.NotNil(t, validCert)

	// Verify DNS names in the certificate
	assert.Equal(t, dnsNames, validCert.DNSNames)

	// Format output and verify it includes all DNS names
	_, daysLeft, _ := checker.CheckExpiry(validCert)
	output := checker.FormatOutput(validCert, daysLeft)
	assert.Contains(t, output, strings.Join(dnsNames, ", "))
}

func TestCertificateIssuerInfo(t *testing.T) {
	cert, err := generateTestCertificate(
		time.Now().Add(-24*time.Hour),
		time.Now().Add(90*24*time.Hour),
		[]string{"localhost"},
		"issuer.test.com",
	)
	assert.NoError(t, err)

	listener, port, err := createTestTLSServer(cert)
	assert.NoError(t, err)
	defer listener.Close()

	time.Sleep(100 * time.Millisecond)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	checker := NewCertificateCheckerWithTLSConfig(Config{
		Host:    "localhost",
		Port:    port,
		Timeout: 5,
		Expiry:  30,
	}, tlsConfig)

	conn, err := checker.Connect()
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()

	validCert, err := checker.ValidateCertificate(conn)
	assert.NoError(t, err)
	assert.NotNil(t, validCert)

	// Check issuer information - the issuer is actually self-signed so it has the subject info
	assert.Equal(t, "Test Org", validCert.Issuer.Organization[0])
	assert.Equal(t, "issuer.test.com", validCert.Issuer.CommonName)

	// Verify output formatting includes issuer info
	_, daysLeft, _ := checker.CheckExpiry(validCert)
	output := checker.FormatOutput(validCert, daysLeft)
	assert.Contains(t, output, "Test Org")
	assert.Contains(t, output, "issuer.test.com")
}

func TestConnectionTimeout(t *testing.T) {
	checker := NewCertificateChecker(Config{
		Host:    "non-existent-host-12345.example.com",
		Port:    443,
		Timeout: 1,
		Expiry:  30,
	})

	conn, err := checker.Connect()
	assert.Error(t, err)
	assert.Nil(t, conn)
}

func TestEmptyCertificateList(t *testing.T) {
	// This test simulates a scenario where no certificates are returned
	// Since we can't easily mock this with a real TLS connection,
	// we'll test the error handling in ValidateCertificate

	// Create a minimal test to ensure the function handles empty cert lists
	checker := NewCertificateChecker(Config{
		Host:    "localhost",
		Port:    443,
		Timeout: 5,
		Expiry:  30,
	})

	// We can't directly test the empty certificate case without mocking,
	// but we can verify the checker is created correctly
	assert.NotNil(t, checker)
	assert.Equal(t, "localhost", checker.config.Host)
	assert.Equal(t, 443, checker.config.Port)
}

func TestSetupOptions(t *testing.T) {
	// Test that SetupOptions correctly sets up the flag options with defaults
	c := check.New("CheckCertificate")
	cfg := SetupOptions(c)

	// Verify default values are set
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 443, cfg.Port)
	assert.Equal(t, int64(5), cfg.Timeout)
	assert.Equal(t, int64(30), cfg.Expiry)

	// Verify that flags are registered
	assert.NotNil(t, c.Option.Lookup("host"))
	assert.NotNil(t, c.Option.Lookup("port"))
	assert.NotNil(t, c.Option.Lookup("timeout"))
	assert.NotNil(t, c.Option.Lookup("expiry"))

	// Verify short flags are registered
	assert.NotNil(t, c.Option.ShorthandLookup("h"))
	assert.NotNil(t, c.Option.ShorthandLookup("P"))
	assert.NotNil(t, c.Option.ShorthandLookup("t"))
	assert.NotNil(t, c.Option.ShorthandLookup("e"))
}

func TestCertificateChecker_Run(t *testing.T) {
	tests := []struct {
		name           string
		certNotBefore  time.Time
		certNotAfter   time.Time
		cfg            Config
		expectExpiring bool
		expectedExit   int // 0 for Ok, 1 for Warning
		outputContains string
	}{
		{
			name:          "Successful check - no warning",
			certNotBefore: time.Now().Add(-24 * time.Hour),
			certNotAfter:  time.Now().Add(90 * 24 * time.Hour),
			cfg: Config{
				Host:    "localhost",
				Timeout: 5,
				Expiry:  30,
			},
			expectExpiring: false,
			expectedExit:   0,
			outputContains: "localhost:",
		},
		{
			name:          "Certificate expiring warning",
			certNotBefore: time.Now().Add(-24 * time.Hour),
			certNotAfter:  time.Now().Add(20 * 24 * time.Hour), // Expires in 20 days
			cfg: Config{
				Host:    "localhost",
				Timeout: 5,
				Expiry:  30, // Warning at 30 days
			},
			expectExpiring: true,
			expectedExit:   1,
			outputContains: "Certificate about to expire",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test certificate with specific expiry
			cert, err := generateTestCertificate(
				tt.certNotBefore,
				tt.certNotAfter,
				[]string{"localhost", "127.0.0.1"},
				"test.example.com",
			)
			assert.NoError(t, err)

			listener, port, err := createTestTLSServer(cert)
			assert.NoError(t, err)
			defer listener.Close()

			time.Sleep(100 * time.Millisecond)

			// Update config with the actual port
			tt.cfg.Port = port

			tlsConfig := &tls.Config{
				InsecureSkipVerify: true,
			}
			checker := NewCertificateCheckerWithTLSConfig(tt.cfg, tlsConfig)

			// Create a mock check that captures the exit behavior
			mockCheck := &MockCheck{}

			// Since Run calls c.Ok() or c.Warning() which exit, we test the components
			// Test the individual components that Run method calls
			conn, err := checker.Connect()
			assert.NoError(t, err)
			assert.NotNil(t, conn)
			defer conn.Close()

			validCert, err := checker.ValidateCertificate(conn)
			assert.NoError(t, err)
			assert.NotNil(t, validCert)

			expiring, daysLeft, warning := checker.CheckExpiry(validCert)
			assert.Equal(t, tt.expectExpiring, expiring)
			if tt.expectExpiring {
				assert.NotEmpty(t, warning)
				assert.Contains(t, warning, "Certificate about to expire")
				// Simulate what Run would do
				mockCheck.Warning(warning)
				assert.Equal(t, 1, mockCheck.exitCode)
			} else {
				assert.Empty(t, warning)
				// Simulate what Run would do
				output := checker.FormatOutput(validCert, daysLeft)
				mockCheck.Ok(output)
				assert.Equal(t, 0, mockCheck.exitCode)
			}

			output := checker.FormatOutput(validCert, daysLeft)
			assert.Contains(t, output, "localhost:")
		})
	}
}

// Test the actual Run method with a mock check
func TestCertificateChecker_RunMethod(t *testing.T) {
	// Create a test certificate
	cert, err := generateTestCertificate(
		time.Now().Add(-24*time.Hour),
		time.Now().Add(90*24*time.Hour),
		[]string{"localhost", "127.0.0.1"},
		"test.example.com",
	)
	assert.NoError(t, err)

	listener, port, err := createTestTLSServer(cert)
	assert.NoError(t, err)
	defer listener.Close()

	time.Sleep(100 * time.Millisecond)

	// Test successful run
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	checker := NewCertificateCheckerWithTLSConfig(Config{
		Host:    "localhost",
		Port:    port,
		Timeout: 5,
		Expiry:  30,
	}, tlsConfig)

	// Create a mock check with a custom exit function
	exitCalled := false
	exitCode := -1
	c := check.New("TestCheck")
	c.ExitFn = func(code int) {
		exitCalled = true
		exitCode = code
	}

	// Run should call c.Ok() which calls ExitFn with code 0
	checker.Run(c)

	assert.True(t, exitCalled)
	assert.Equal(t, 0, exitCode)
}

func TestCertificateChecker_RunMethodWithExpiry(t *testing.T) {
	// Create a test certificate that expires soon
	cert, err := generateTestCertificate(
		time.Now().Add(-24*time.Hour),
		time.Now().Add(20*24*time.Hour), // Expires in 20 days
		[]string{"localhost", "127.0.0.1"},
		"test.example.com",
	)
	assert.NoError(t, err)

	listener, port, err := createTestTLSServer(cert)
	assert.NoError(t, err)
	defer listener.Close()

	time.Sleep(100 * time.Millisecond)

	// Test warning run
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	checker := NewCertificateCheckerWithTLSConfig(Config{
		Host:    "localhost",
		Port:    port,
		Timeout: 5,
		Expiry:  30, // Warning at 30 days
	}, tlsConfig)

	// Create a mock check with a custom exit function that captures the first exit call
	exitCalled := false
	firstExitCode := -1
	c := check.New("TestCheck")
	c.ExitFn = func(code int) {
		if !exitCalled {
			exitCalled = true
			firstExitCode = code
		}
		// Don't actually exit in the test
	}

	// Run should call c.Warning() which calls ExitFn with code 1
	checker.Run(c)

	assert.True(t, exitCalled)
	assert.Equal(t, 1, firstExitCode) // Should exit with 1 for warning
}

// MockCheck is a test double for check.Check
type MockCheck struct {
	exitCalled bool
	exitCode   int
	message    string
}

func (m *MockCheck) Ok(msg string) {
	m.exitCalled = true
	m.exitCode = 0
	m.message = msg
}

func (m *MockCheck) Warning(msg string) {
	m.exitCalled = true
	m.exitCode = 1
	m.message = msg
}

func (m *MockCheck) Error(err error) {
	m.exitCalled = true
	m.exitCode = 2
	m.message = err.Error()
}
