package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/portertech/sensu-plugins-go/lib/check"
)

type request struct {
	url      string
	// redirect bool # unused and detected by linter
	timeout  time.Duration
	insecure bool
	username string
	password string
	method   string
	body     string
	code     int
	pattern  string
	proxyURL string
	noProxy  bool
}

func main() {
	request := request{}

	c := check.New("check-http-json")
	c.Option.SortFlags = false
	c.Option.StringVarP(&request.url, "url", "u", "http://localhost/", "URL")
	c.Option.DurationVarP(&request.timeout, "timeout", "t", 15*time.Second, "Timeout")
	c.Option.StringVarP(&request.username, "username", "", "", "Username for basic authentication")
	c.Option.StringVarP(&request.password, "password", "", "", "Password for basic authentication")
	c.Option.BoolVarP(&request.insecure, "insecure", "k", false, "INSECURE (skips peer certificate validation)")
	c.Option.StringVarP(&request.method, "method", "m", "GET", "HTTP methods such as GET, POST, PUT, DELETE, PATCH etc.")
	c.Option.StringVarP(&request.body, "body", "b", "", "Body string to pass with request")
	c.Option.StringVarP(&request.pattern, "pattern", "p", "", "Regular expression pattern to match against response body (See https://github.com/google/re2/wiki/Syntax)")
	c.Option.StringVarP(&request.proxyURL, "proxy-url", "", "", "Proxy URL which can include a PORT")
	c.Option.BoolVarP(&request.noProxy, "no-proxy", "", false, "Do not use http proxy (also not from environment)")
	c.Option.IntVarP(&request.code, "code", "c", 200, "Expected response code")
	c.Init()

	status, response, err := send(&request)
	if err != nil {
		c.Error(err)
	}

	switch {
	case status == "CRITICAL":
		c.Critical(response)
	default:
		c.Ok(response)
	}
}

func send(request *request) (string, string, error) {
	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: request.insecure},
	}

	if request.noProxy {
		tr.Proxy = nil
	}

	if len(request.proxyURL) > 0 {
		proxyURL, err := url.Parse(request.proxyURL)
		if err != nil {
			return "", "", err
		}
		tr.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{Transport: tr, Timeout: request.timeout}

	r, err := http.NewRequest(request.method, request.url, strings.NewReader(request.body))
	if err != nil {
		return "", "", err
	}
	r.Header.Set("Content-Type", "application/json")

	if len(request.username) > 0 || len(request.password) > 0 {
		r.SetBasicAuth(request.username, request.password)
	}

	start := time.Now()
	resp, err := client.Do(r)
	if err != nil {
		return "CRITICAL", "", err
	}
	took := float64(time.Since(start)) / float64(time.Millisecond)
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != request.code {
		return "CRITICAL", fmt.Sprintf("Status code [%d], body [%s]", resp.StatusCode, responseBody), nil
	}

	if len(request.pattern) > 0 {
		// pattern match aginst body
		re, err := regexp.Compile(request.pattern)
		if err != nil {
			return "", "", err
		}

		if !re.Match([]byte(responseBody)) {
			return "CRITICAL", fmt.Sprintf("Status code [%d], pattern [%s] doesn't match with [%s]", resp.StatusCode, request.pattern, responseBody), nil
		}
	}

	return "OK", fmt.Sprintf("Status code [%d], took [%0.1f ms]", resp.StatusCode, took), nil
}
