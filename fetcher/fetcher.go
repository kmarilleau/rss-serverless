package fetcher

import (
	"errors"
	"io/ioutil"
	"net/http"
)

func fetchURLBody(url string) (string, error) {
	resp, err := http.Get(url)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	return string(bodyBytes), err
}
