package fetcher

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

var httpTestCases = []struct {
	name       string
	url        string
	statusCode int
	body       string
	wantBody   string
	wantErr    string
}{
	{name: "ok", url: "http://foo.bar", statusCode: 200, body: "spam", wantBody: "spam"},
	{name: "invalid url", url: "&", wantErr: `Get "&": unsupported protocol scheme ""`},
	{name: "not found", url: "http://foo.bar", statusCode: 404},
	{name: "internal server error", url: "http://foo.bar", statusCode: 500},
	{name: "service unavailable", url: "http://foo.bar", statusCode: 503},
}

func TestHttpGet(t *testing.T) {
	for _, tt := range httpTestCases {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			if tt.statusCode != 0 {
				defer gock.Off()
				gock.New(tt.url).
					Reply(tt.statusCode).
					BodyString(tt.body)
			}

			resp, err := http.Get(tt.url)
			if tt.wantErr != "" {
				assert.Error(err)
				assert.EqualError(err, tt.wantErr, "The two HTTP errors should be the same.")
			} else {
				assert.NotNil(resp.Body)

				bodyBytes, _ := io.ReadAll(resp.Body)
				body := string(bodyBytes)

				assert.NoError(err)
				assert.Equal(tt.statusCode, resp.StatusCode, "The two status codes should be the same.")
				assert.Equal(tt.wantBody, body, "The two body contents should be the same.")
			}
		})
	}
}

func TestFetchUrlBody(t *testing.T) {
	for _, tt := range httpTestCases {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			if tt.statusCode != 0 {
				defer gock.Off()
				gock.New(tt.url).
					Reply(tt.statusCode).
					BodyString(tt.body)
			}

			actualBody, err := fetchURLBody(tt.url)

			if tt.wantErr != "" {
				assert.Error(err)
				assert.EqualError(err, tt.wantErr, "The two HTTP errors should be the same.")
			} else if tt.statusCode != 200 {
				assert.Error(err)
				assert.True(httpErrorStartsWithStatusCode(err, tt.statusCode))
			} else {
				assert.NotNil(actualBody)
				assert.NoError(err)
			}
		})
	}
}
