package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/thomis/sensu-plugins-go/pkg/check"
)

type alivenessStruct struct {
	Status string
}

type connection struct {
	Host     string
	Port     int
	Vhost    string
	User     string
	Password string
	Timeout  int
}

func main() {
	var (
		connection connection
	)

	c := check.New("CheckRabbitMQ")
	c.Option.StringVarP(&connection.Host, "host", "h", "localhost", "HOST")
	c.Option.IntVarP(&connection.Port, "port", "P", 15672, "PORT")
	c.Option.StringVarP(&connection.Vhost, "vhost", "v", "%2F", "VHOST")
	c.Option.StringVarP(&connection.User, "user", "u", "guest", "USER")
	c.Option.StringVarP(&connection.Password, "password", "p", "guest", "PASSWORD")
	c.Option.IntVarP(&connection.Timeout, "timeout", "t", 10, "TIMEOUT")
	c.Init()

	status, err := alivenessTest(connection)
	if err != nil {
		c.Error(err)
	}

	switch status {
	case "ok":
		c.Ok("RabbitMQ server is alive")
	default:
		c.Warning("Object Not Found")
	}
}

func alivenessTest(connection connection) (string, error) {
	var aliveness alivenessStruct
	http.DefaultClient.Timeout = time.Duration(connection.Timeout) * time.Second

	request := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Host:   connection.Host + ":" + strconv.Itoa(connection.Port),
			Scheme: "http",
			Opaque: "/api/aliveness-test/" + connection.Vhost,
		},
		Header: http.Header{
			"User-Agent": {"godoc-example/0.1"},
		},
	}
	request.SetBasicAuth(connection.User, connection.Password)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	json.Unmarshal(body, &aliveness)
	return aliveness.Status, nil
}
