package bosh

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (c Client) TaskResult(taskId int) (map[string]interface{}, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/tasks/%d/output?type=result", c.config.URL, taskId), nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(c.config.Username, c.config.Password)

	response, err := transport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	body, err := bodyReader(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response %s:\n%s", response.Status, body)
	}

	result := map[string]interface{}{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
