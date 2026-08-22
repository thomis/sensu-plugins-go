package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/thomis/sensu-plugins-go/pkg/check"
)

// bodyLimit caps how much of the response body is read for pattern matching.
const bodyLimit = 10 * 1024 * 1024

type input struct {
	Url      string
	Timeout  int
	Insecure bool
	Username string
	Password string
	Pattern  string
}

func main() {
	var (
		input    input
		redirect bool
	)

	c := check.New("CheckHTTP")
	c.Option.StringVarP(&input.Url, "url", "u", "http://localhost/", "URL")
	c.Option.IntVarP(&input.Timeout, "timeout", "t", 15, "TIMEOUT")
	c.Option.StringVarP(&input.Username, "username", "", "", "Username for basic authentication")
	c.Option.StringVarP(&input.Password, "password", "", "", "Password for basic authentication")
	c.Option.BoolVarP(&input.Insecure, "insecure", "k", false, "INSECURE (skips peer certificate validation)")
	c.Option.StringVarP(&input.Pattern, "pattern", "p", "", "PATTERN (regular expression the response body must match, critical if not found)")

	c.Init()

	pattern, err := compilePattern(input.Pattern)
	if err != nil {
		c.Error(err)
	}

	status, body, err := fetch(input)
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
	case pattern != nil && !pattern.MatchString(body):
		c.Critical(fmt.Sprintf("%d, pattern [%s] not found in response body", status, input.Pattern))
	case pattern != nil:
		c.Ok(fmt.Sprintf("%d, pattern [%s] found in response body", status, input.Pattern))
	default:
		c.Ok(strconv.Itoa(status))
	}
}

func compilePattern(pattern string) (*regexp.Regexp, error) {
	if pattern == "" {
		return nil, nil
	}
	return regexp.Compile(pattern)
}

func fetch(input input) (int, string, error) {
	c := http.Client{
		Timeout: time.Duration(input.Timeout) * time.Second,
		Transport: &http.Transport{
			ResponseHeaderTimeout: time.Duration(input.Timeout) * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: input.Insecure}}}

	request, err := http.NewRequest(http.MethodGet, input.Url, http.NoBody)
	if err != nil {
		return 0, "", err
	}

	if len(input.Username) > 0 || len(input.Password) > 0 {
		request.SetBasicAuth(input.Username, input.Password)
	}

	response, err := c.Do(request)
	if err != nil {
		return 0, "", err
	}
	defer response.Body.Close()

	if input.Pattern == "" {
		return response.StatusCode, "", nil
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, bodyLimit))
	if err != nil {
		return 0, "", err
	}

	return response.StatusCode, string(body), nil
}
