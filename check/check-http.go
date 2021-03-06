package main

import (
	"crypto/tls"
	"net/http"
	"strconv"
	"time"

	"github.com/portertech/sensu-plugins-go/lib/check"
)

func main() {
	var (
		url      string
		redirect bool
		timeout  int
		insecure bool
		username string
		password string
	)

	c := check.New("CheckHTTP")
	c.Option.StringVarP(&url, "url", "u", "http://localhost/", "URL")
	c.Option.BoolVarP(&redirect, "redirect", "r", false, "REDIRECT")
	c.Option.IntVarP(&timeout, "timeout", "t", 15, "TIMEOUT")
	c.Option.StringVarP(&username, "username", "", "", "Username for basic authentication")
	c.Option.StringVarP(&password, "password", "", "", "Password for basic authentication")
	c.Option.BoolVarP(&insecure, "insecure", "k", false, "INSECURE (skips peer certificate validation)")

	c.Init()

	status, err := statusCode(url, timeout, insecure, username, password)
	if err != nil {
		c.Error(err)
	}

	switch {
	case status >= 400:
		c.Critical(strconv.Itoa(status))
	case status >= 300 && redirect:
		c.Ok(strconv.Itoa(status))
	case status >= 300:
		c.Warning(strconv.Itoa(status))
	default:
		c.Ok(strconv.Itoa(status))
	}
}

func statusCode(url string, timeout int, insecure bool, username string, password string) (int, error) {
	http.DefaultClient.Timeout = time.Duration(timeout) * time.Second

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	if len(username) > 0 || len(password) > 0 {
		request.SetBasicAuth(username, password)
	}

	response, err := transport.RoundTrip(request)
	if err != nil {
		return 0, err
	}

	return response.StatusCode, nil
}
