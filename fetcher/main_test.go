package fetcher

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"net/url"
	"testing"
)

const BODY = "egg"

var uris = []string{
	// Valid URIs
	"http://foo.com/blah_blah",
	"http://www.example.com/wpstyle/?p=364",
	"https://www.example.com/foo/?bar=baz&inga=42&quux",
	"http://223.255.255.254",
	// Invalid URIs
	"htp://foo.com",
	"//foo.com",
	"https:2&quux",
}

var httpResponses = []struct {
	Status     string
	StatusCode int
}{
	{"OK", 200},
	{"Not Found", 404},
	{"Gone", 410},
	{"Internal Server Error", 500},
	{"Service Unavailable", 503},
}

type FetchURLTestCase struct {
	TestCase       string
	URI            string
	URIIsValid     bool
	HTTPStatusCode int
	Body           string
}

func URIIsValid(uri string) bool {
	parsedURI, err := url.ParseRequestURI(uri)

	if err != nil {
		return false
	}

	switch parsedURI.Scheme {
	case "http", "https":
		return true
	default:
		return false
	}
}

func generateFetchURLTestCase() []FetchURLTestCase {
	tts := []FetchURLTestCase{}
	for _, uri := range uris {
		var testCaseMsgStart string
		if URIIsValid(uri) {
			testCaseMsgStart = fmt.Sprintf("Valid URL(%s)", uri)
		} else {
			testCaseMsgStart = fmt.Sprintf("Invalid URL(%s)", uri)
		}

		for _, http := range httpResponses {
			testCaseMsg := fmt.Sprintf("%s %d %s", testCaseMsgStart, http.StatusCode, http.Status)
			tts = append(tts, FetchURLTestCase{
				TestCase:       testCaseMsg,
				URI:            uri,
				URIIsValid:     URIIsValid(uri),
				HTTPStatusCode: http.StatusCode,
				Body:           BODY,
			})
		}
	}

	return tts
}

func TestFetchUrlBody(t *testing.T) {
	for _, tt := range generateFetchURLTestCase() {
		t.Run(tt.TestCase, func(t *testing.T) {
			assert := assert.New(t)
			defer gock.Off()

			gock.New(tt.URI).
				Reply(tt.HTTPStatusCode).
				BodyString(BODY)

			got, err := FetchURLBody(tt.URI)

			if tt.URIIsValid {
				switch tt.HTTPStatusCode {
				case 200:
					assert.Equal(BODY, got)
					assert.NoError(err)
				default:
					assert.Equal("", got)
					assert.Error(err)
					assert.IsType(&httpError{}, err)
				}
			} else {
				assert.Equal("", got)
				assert.Error(err)
			}
		})
	}
}
