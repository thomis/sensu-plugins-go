package main

import (
	"net"
	"strconv"
	"time"

	"github.com/thomis/sensu-plugins-go/pkg/check"
)

func main() {
	var (
		host    string
		port    int
		timeout int64
	)

	c := check.New("CheckPing")
	c.Option.StringVarP(&host, "host", "h", "localhost", "HOST")
	c.Option.IntVarP(&port, "port", "P", 22, "PORT")
	c.Option.Int64VarP(&timeout, "timeout", "t", 5, "TIMEOUT")
	c.Init()

	address, err := checkConnection(host, port, timeout)
	if err != nil {
		c.Error(err)
		return
	}

	c.Ok(address)
}

// checkConnection attempts a TCP connection to host:port within the timeout
// (seconds). It returns the dialed address and any connection error.
func checkConnection(host string, port int, timeout int64) (string, error) {
	address := host + ":" + strconv.Itoa(port)

	conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Second)
	if err != nil {
		return address, err
	}
	defer conn.Close()

	return address, nil
}
