package main

import (
	"net"
	"strconv"
	"time"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/portertech/sensu-plugins-go/lib/check"
)

func main() {
	var (
		host    string
		port    int
		timeout int64
		expiry  int64
	)

	c := check.New("CheckCertificate")
	c.Option.StringVarP(&host, "host", "h", "localhost", "HOST")
	c.Option.IntVarP(&port, "port", "P", 443, "PORT")
	c.Option.Int64VarP(&timeout, "timeout", "t", 5, "TIMEOUT")
	c.Option.Int64VarP(&expiry, "expiry", "e", 30, "EXPIRY warning in days")
	c.Init()

	address := host + ":" + strconv.Itoa(port)

	dialer := &net.Dialer{Timeout: time.Duration(timeout)*time.Second}

	conn, err := tls.DialWithDialer(dialer, "tcp", address, nil)
	if err != nil {
		c.Error(err)
	}
	defer conn.Close()

	// check if hostname matches with certificate
  err = conn.VerifyHostname(host)
  if err != nil {
  	c.Error(err)
  }

  certs := conn.ConnectionState().PeerCertificates
  if len(certs) == 0 {
  	c.Error(fmt.Errorf("Unable to find or retrieve certificates"))
  }

  // check date validity
  cert  := certs[0]
  now := time.Now()
  if now.Before(cert.NotBefore) {
  	c.Error(fmt.Errorf("Certificate Not Before: " + cert.NotBefore.Format("2006-01-02 03:04:05")))
  }

  if now.After(cert.NotAfter) {
  	c.Error(fmt.Errorf("Certificate Not After: " + cert.NotAfter.Format("2006-01-02 03:04:05")))
  }

  // expiry warning in days
  days_left := cert.NotAfter.Sub(now).Hours() / 24
  if now.Add(time.Duration(expiry)*24*time.Hour).After(cert.NotAfter) {
  	info := fmt.Sprintf("Certificate about to expire in less than %d days. Not After: %s (%0.1f days left)",
  	expiry,cert.NotAfter.Format("2006-01	-02 03:04:05"),
  	days_left)
  	c.Warning(info)
  }

  output := address + "\n\n"
  output += fmt.Sprintf("Issuer Name: %s\n", cert.Issuer)
  output += fmt.Sprintf("Not Before : %s\n",
  	cert.NotBefore.Format("2006-01-02 03:04:05"))
  output += fmt.Sprintf("Not After  : %s (%0.1f days left)\n",
  	cert.NotAfter.Format("2006-01-02 03:04:05"),
  	days_left)
  output += fmt.Sprintf("Common Name: %s\n", cert.Issuer.CommonName)
  output += fmt.Sprintf("DNS Names  : %s\n", strings.Join(cert.DNSNames, ", "))

	c.Ok(output)
}
