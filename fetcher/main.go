package fetcher

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type httpError struct {
	StatusCode int
	Status     string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Status)
}


func FetchURLBody(url string) (string, error) {
	resp, err := http.Get(url)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", &httpError{StatusCode: resp.StatusCode, Status: resp.Status}
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	return string(bodyBytes), err
}
