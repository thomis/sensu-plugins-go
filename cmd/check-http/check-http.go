package main

import (
	"crypto/tls"
	"net/http"
	"strconv"
	"time"

	"github.com/thomis/sensu-plugins-go/pkg/check"
)

type input struct {
	Url      string
	Timeout  int
	Insecure bool
	Username string
	Password string
}

func main() {
	var (
		input input
		redirect bool
	)

	c := check.New("CheckHTTP")
	c.Option.StringVarP(&input.Url, "url", "u", "http://localhost/", "URL")
	c.Option.IntVarP(&input.Timeout, "timeout", "t", 15, "TIMEOUT")
	c.Option.StringVarP(&input.Username, "username", "", "", "Username for basic authentication")
	c.Option.StringVarP(&input.Password, "password", "", "", "Password for basic authentication")
	c.Option.BoolVarP(&input.Insecure, "insecure", "k", false, "INSECURE (skips peer certificate validation)")

	c.Init()

	status, err := statusCode(input)
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

func statusCode(input input) (int, error) {
	c := http.Client{
		Timeout: time.Duration(input.Timeout) * time.Second,
		Transport: &http.Transport{
			ResponseHeaderTimeout: time.Duration(input.Timeout) * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: input.Insecure}}}

	request, err := http.NewRequest(http.MethodGet, input.Url, http.NoBody)
	if err != nil {
		return 0, err
	}

	if len(input.Username) > 0 || len(input.Password) > 0 {
		request.SetBasicAuth(input.Username, input.Password)
	}

	response, err := c.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	return response.StatusCode, nil
}
