package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/kmarilleau/rfe"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

type Feed struct {
	createdAt          time.Time
	fetchedAt          time.Time
	isCurrentlyFetched bool
	url                string
}

func createHTTPRequest(body string) (req *http.Request) {
	req = httptest.NewRequest("GET", "/", strings.NewReader(body))
	req.Header.Add("Content-Type", "application/json")
	return
}

func getResultData(rr *httptest.ResponseRecorder) (data Result) {
	data = Result{}
	body := rr.Body.String()
	json.Unmarshal([]byte(body), &data)

	return
}

func TestGetRequestData(t *testing.T) {
	var tts = []struct {
		name    string
		json    string
		want    RequestData
		wantErr string
	}{
		{name: "valid JSON", json: `{"url": "http://foo.bar", "documentId": "abcdef123"}`, want: RequestData{URL: "http://foo.bar", DocumentID: "abcdef123"}},
		{name: "invalid JSON", json: `{invalid json}`, wantErr: "invalid character 'i' looking for beginning of object key string"},
		{name: "required JSON fields not found", json: `{"example": "notvalid"}`, wantErr: RequiredJSONFieldsNotFound},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			req := createHTTPRequest(tt.json)

			got, err := getRequestData(req)

			assert.IsType(RequestData{}, got)
			assert.True(reflect.DeepEqual(tt.want, got), fmt.Sprintf("got: %+v\nwant: %+v", got, tt.want))

			if err != nil {
				assert.EqualError(err, tt.wantErr)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestFetchURLAndStoreItsContent(t *testing.T) {
	var tts = []struct {
		name       string
		url        string
		statusCode int
		body       string
		wantBody   string
		wantStatus string
	}{
		{name: "ok", url: "http://foo.bar", statusCode: 200, body: "spam", wantBody: "spam", wantStatus: "OK"},
		{name: "invalid url", url: "&", wantStatus: `Get "&": unsupported protocol scheme ""`},
		{name: "not found", url: "http://foo.bar", statusCode: 404, wantStatus: "404 Not Found"},
		{name: "internal server error", url: "http://foo.bar", statusCode: 500, wantStatus: "500 Internal Server Error"},
		{name: "service unavailable", url: "http://foo.bar", statusCode: 503, wantStatus: "503 Service Unavailable"},
	}

	ctx := context.Background()
	dbClient, err := firestore.NewClient(ctx, "test")
	if err != nil {
		log.Fatal(err)
	}
	defer dbClient.Close()

	createFeedDoc := func(url string) (doc *firestore.DocumentRef) {
		doc, _, err := dbClient.Collection(FeedsCollection).Add(ctx, Feed{
			createdAt: time.Now(),
			fetchedAt: time.Unix(0, 0),
			url:       url,
		})
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	getJSONRequestData := func(feedDocId string, url string) string {
		jsonBytes, _ := json.Marshal(RequestData{
			DocumentID:  feedDocId,
			URL:         url,
		})

		return string(jsonBytes)
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			feedDoc := createFeedDoc(tt.url)

			reqJSONStr := getJSONRequestData(feedDoc.ID, tt.url)
			req := createHTTPRequest(reqJSONStr)
			rr := httptest.NewRecorder()

			if tt.statusCode != 0 {
				defer gock.Off()
				gock.New(tt.url).
					Reply(tt.statusCode).
					BodyString(tt.body)
			}

			FetchURLAndStoreItsContent(rr, req)

			// Test Result Object
			gotResult := getResultData(rr)
			assert.Equal(tt.wantBody, gotResult.Body)
			assert.Equal(tt.wantStatus, gotResult.Status)

			// Test Feed Document
			feed := FeedDoc{}
			feed.LoadFromDocRef(ctx, feedDoc)

			assert.False(feed.Locked)
			assert.Equal(tt.wantStatus, feed.LastStatus)

			// Test Fetched Document
			if tt.wantStatus == LastStatusOk {
				fetchedDoc := dbClient.Collection(FetchedCollection).Doc(gotResult.FetchedDocID)
				fetched := FetchedDoc{}
				fetched.LoadFromDocRef(ctx, fetchedDoc)

				assert.Equal(feedDoc.ID, fetched.FeedRef)
				assert.Equal(tt.wantBody, fetched.Content)
			}

			clearAllFirestoreData(ctx, dbClient)
		})
	}
}

func TestMain(m *testing.M) {
	firestoreEmulator := rfe.FirestoreEmulator{Verbose: true}
	firestoreEmulator.Start()
	defer firestoreEmulator.Shutdown()

	m.Run()
}
