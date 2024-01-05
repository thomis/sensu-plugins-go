package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/thomis/sensu-plugins-go/pkg/check"
)

type item struct {
	Host string
	Port int
	Timeout int64
	Expiry int64
}

func main() {
	var (
		item item
	)

	c := check.New("CheckCertificate")
	c.Option.StringVarP(&item.Host, "host", "h", "localhost", "HOST")
	c.Option.IntVarP(&item.Port, "port", "P", 443, "PORT")
	c.Option.Int64VarP(&item.Timeout, "timeout", "t", 5, "TIMEOUT")
	c.Option.Int64VarP(&item.Expiry, "expiry", "e", 30, "EXPIRY warning in days")
	c.Init()

	address := item.Host + ":" + strconv.Itoa(item.Port)

	dialer := &net.Dialer{Timeout: time.Duration(item.Timeout) * time.Second}

	conn, err := tls.DialWithDialer(dialer, "tcp", address, nil)
	if err != nil {
		c.Error(err)
	}
	defer conn.Close()

	// check if hostname matches with certificate
	err = conn.VerifyHostname(item.Host)
	if err != nil {
		c.Error(err)
	}

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		c.Error(fmt.Errorf("unable to find or retrieve certificates"))
	}

	// check date validity
	cert := certs[0]
	now := time.Now()
	if now.Before(cert.NotBefore) {
		c.Error(fmt.Errorf("Certificate Not Before: " + cert.NotBefore.Format("2006-01-02 15:04:05") + " UTC"))
	}

	if now.After(cert.NotAfter) {
		c.Error(fmt.Errorf("Certificate Not After: " + cert.NotAfter.Format("2006-01-02 15:04:05") + " UTC"))
	}

	// expiry warning in days
	days_left := cert.NotAfter.Sub(now).Hours() / 24
	if now.Add(time.Duration(item.Expiry) * 24 * time.Hour).After(cert.NotAfter) {
		info := fmt.Sprintf("Certificate about to expire in less than %d days. Not After: %s (%0.1f days left)",
			item.Expiry, cert.NotAfter.Format("2006-01-02 15:04:05")+" UTC",
			days_left)
		c.Warning(info)
	}

	output := address + "\n\n"
	output += fmt.Sprintf("Issuer Name: %s\n", cert.Issuer)
	output += fmt.Sprintf("Not Before : %s\n",
		cert.NotBefore.Format("2006-01-02 15:04:05")+" UTC")
	output += fmt.Sprintf("Not After  : %s (%0.1f days left)\n",
		cert.NotAfter.Format("2006-01-02 15:04:05")+" UTC",
		days_left)
	output += fmt.Sprintf("Common Name: %s\n", cert.Issuer.CommonName)
	output += fmt.Sprintf("DNS Names  : %s\n", strings.Join(cert.DNSNames, ", "))

	c.Ok(output)
}
