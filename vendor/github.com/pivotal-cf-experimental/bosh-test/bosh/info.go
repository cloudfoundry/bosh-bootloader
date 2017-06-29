package bosh

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type DirectorInfo struct {
	UUID string
	CPI  string
}

func (c Client) Info() (DirectorInfo, error) {
	response, err := client.Get(fmt.Sprintf("%s/info", c.config.URL))
	if err != nil {
		return DirectorInfo{}, err
	}

	if response.StatusCode != http.StatusOK {
		body, err := bodyReader(response.Body)
		if err != nil {
			return DirectorInfo{}, err
		}
		defer response.Body.Close()

		return DirectorInfo{}, fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	info := DirectorInfo{}
	err = json.NewDecoder(response.Body).Decode(&info)
	if err != nil {
		return DirectorInfo{}, err
	}

	return info, nil
}
