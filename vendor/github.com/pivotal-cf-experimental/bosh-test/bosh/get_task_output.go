package bosh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type TaskOutput struct {
	Time     int64
	Error    TaskError
	Stage    string
	Tags     []string
	Total    int
	Task     string
	Index    int
	State    string
	Progress int
}

type TaskError struct {
	Code    int
	Message string
}

func (te TaskError) Error() string {
	return fmt.Sprintf("task error: %d has occurred: %s", te.Code, te.Message)
}

func (te TaskError) ErrorCode() int {
	return te.Code
}

func (c Client) GetTaskOutput(taskId int) ([]TaskOutput, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/tasks/%d/output?type=event", c.config.URL, taskId), nil)
	if err != nil {
		return []TaskOutput{}, err
	}
	request.SetBasicAuth(c.config.Username, c.config.Password)

	response, err := transport.RoundTrip(request)
	if err != nil {
		return []TaskOutput{}, err
	}

	body, err := bodyReader(response.Body)
	if err != nil {
		return []TaskOutput{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return []TaskOutput{}, fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	body = bytes.TrimSpace(body)
	parts := bytes.Split(body, []byte("\n"))

	var taskOutputs []TaskOutput
	for _, part := range parts {
		var taskOutput TaskOutput
		err = json.Unmarshal(part, &taskOutput)
		if err != nil {
			return []TaskOutput{}, err
		}

		taskOutputs = append(taskOutputs, taskOutput)
	}

	return taskOutputs, nil
}
