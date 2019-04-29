package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/portertech/sensu-plugins-go/lib/check"
)

func main() {
	var (
		url         string
		pidFile     string
		timeout     int
		checkStatus bool
	)

	c := check.New("CheckNGINX")
	c.Option.StringVarP(&url, "url", "u", "http://localhost/nginx-status", "NGINX status page URL")
	c.Option.StringVarP(&pidFile, "pidFile", "p", "/var/run/nginx.pid", "NGINX PID File")
	c.Option.IntVarP(&timeout, "timeout", "t", 15, "NGINX status page check timeout")
	c.Option.BoolVarP(&checkStatus, "checkStatus", "c", false, "Check NGINX status page")
	c.Init()

	processStatus, processErr := checkProcessRunning(pidFile)
        if ! processStatus {
                c.Critical(fmt.Sprintf("%v", processErr))
        }

	if checkStatus {
		connections, statusErr := nginxStatus(url, timeout)
		if statusErr != nil {
			c.Critical(fmt.Sprintf("%v", statusErr))
		}

		c.Ok(fmt.Sprintf("connections = %d | nginx_connections=%d", connections, connections))
	}
        c.Ok("OK")

}

func checkProcessRunning(pidFile string) (bool, error) {

	pidLine, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return false, fmt.Errorf("failed to read PID file %s, error: %s", pidFile, err)
	}

	pid, err := strconv.ParseInt(strings.TrimRight(string(pidLine), "\n"), 10, 64)
	if err != nil {
		return false, fmt.Errorf("failed to determine PID from PID file %s, error: %s", pidFile, err)
	}

        process, err := os.FindProcess(int(pid))
	if err != nil {
		return false, fmt.Errorf("failed to find process for PID %d, error: %s", pid, err)
	}

	signalErr := process.Signal(syscall.Signal(0))
	if signalErr == nil {
		return true, nil
	} else {
		return false, fmt.Errorf("failed to find process for PID %d, error: %s", pid, err)
	}

}

func nginxStatus(url string, timeout int) (int64, error) {

	http.DefaultClient.Timeout = time.Duration(timeout) * time.Second

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return -1, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return -1, err
	}

	if response.StatusCode != 200 {
		return -1, fmt.Errorf("Received HTTP status code %v from %s", response.StatusCode, url)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return -1, err
	}
	defer response.Body.Close()

	/*  Body should be similar to:

	    Active connections: 43
	    server accepts handled requests
	    7368 7368 10993
	    Reading: 0 Writing: 5 Waiting: 38 */

	lines := strings.Split(string(body), "\n")
	connections, err := strconv.ParseInt(strings.Fields(lines[0])[2], 10, 64)

	return connections, nil
}
