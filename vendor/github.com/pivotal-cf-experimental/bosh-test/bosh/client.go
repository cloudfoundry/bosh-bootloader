package bosh

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var (
	client     = http.DefaultClient
	transport  = http.DefaultTransport
	bodyReader = ioutil.ReadAll
)

type Config struct {
	URL                 string
	Host                string
	DirectorCACert      string
	Username            string
	Password            string
	TaskPollingInterval time.Duration
	AllowInsecureSSL    bool
}

type Client struct {
	config Config
}

type Task struct {
	Id     int
	State  string
	Result string
}

func NewClient(config Config) Client {
	if config.TaskPollingInterval == time.Duration(0) {
		config.TaskPollingInterval = 5 * time.Second
	}

	if config.AllowInsecureSSL {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client = &http.Client{
			Transport: transport,
		}
	}

	return Client{
		config: config,
	}
}

func (c Client) GetConfig() Config {
	return c.config
}

func (c Client) rewriteURL(uri string) (string, error) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	parsedURL.Scheme = ""
	parsedURL.Host = ""

	return c.config.URL + parsedURL.String(), nil
}

func (c Client) checkTask(location string) (Task, error) {
	location, err := c.rewriteURL(location)
	if err != nil {
		return Task{}, err
	}

	var task Task
	request, err := http.NewRequest("GET", location, nil)
	if err != nil {
		return task, err
	}
	request.SetBasicAuth(c.config.Username, c.config.Password)

	response, err := transport.RoundTrip(request)
	if err != nil {
		return task, err
	}

	err = json.NewDecoder(response.Body).Decode(&task)
	if err != nil {
		return task, err
	}

	return task, nil
}

func (c Client) checkTaskStatus(location string) (int, error) {
	for {
		task, err := c.checkTask(location)
		if err != nil {
			return 0, err
		}

		switch task.State {
		case "done":
			return task.Id, nil
		case "error":
			taskOutputs, err := c.GetTaskOutput(task.Id)
			if err != nil {
				return task.Id, fmt.Errorf("failed to get full bosh task event log, bosh task failed with an error status %q", task.Result)
			}
			return task.Id, taskOutputs[len(taskOutputs)-1].Error
		case "errored":
			taskOutputs, err := c.GetTaskOutput(task.Id)
			if err != nil {
				return task.Id, fmt.Errorf("failed to get full bosh task event log, bosh task failed with an errored status %q", task.Result)
			}
			return task.Id, taskOutputs[len(taskOutputs)-1].Error
		case "cancelled":
			return task.Id, errors.New("bosh task was cancelled")
		default:
			time.Sleep(c.config.TaskPollingInterval)
		}
	}
}
