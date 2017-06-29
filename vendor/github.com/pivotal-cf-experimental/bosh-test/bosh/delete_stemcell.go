package bosh

import (
	"fmt"
	"net/http"
)

type InUse struct {
	Use bool
}

type errorStatus interface {
	ErrorCode() int
}

func (i InUse) Error() string {
	return "stemcell is in use"
}

func (i InUse) InUse() bool {
	return i.Use
}

func (c Client) DeleteStemcell(name, version string) error {
	request, err := http.NewRequest("DELETE", fmt.Sprintf("%s/stemcells/%s/%s", c.config.URL, name, version), nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "text/yaml")
	request.SetBasicAuth(c.config.Username, c.config.Password)

	response, err := transport.RoundTrip(request)
	if err != nil {
		return err
	}

	body, err := bodyReader(response.Body)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusFound {
		return fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	_, err = c.checkTaskStatus(response.Header.Get("Location"))
	if err != nil {
		if errStatus, ok := err.(errorStatus); ok {
			if errStatus.ErrorCode() == 50004 {
				return InUse{true}
			}
		}

		return err
	}

	return nil
}
