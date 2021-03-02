package fetcher

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
)

const FeedsCollection = "feeds"
const FetchedCollection = "fetched_feeds"

const RequiredJSONFieldsNotFound = "Required JSON fields not found"
const LastStatusOk = "OK"

var env = getEnv()
var projectID string
var dbClient *firestore.Client
var globctx context.Context

type RequestData struct {
	DocumentID string `json:"documentId"`
	URL        string `json:"url"`
}

type Result struct {
	Body         string `json:"body"`
	Status       string `json:"status"`
	FetchedDocID string `json:"fetchedDocId"`
}

type Doc interface {
	LoadFromDocRef()
}

type FeedDoc struct {
	Locked     bool      `firestore:"currentlyFetched"`
	LastStatus string    `firestore:"lastFetchStatus"`
	FetchedAT  time.Time `firestore:"fetchedAt"`
}

func (s *FeedDoc) LoadFromDocRef(ctx context.Context, doc *firestore.DocumentRef) error {
	docSnap, err := doc.Get(ctx)
	if err != nil {
		return err
	}
	err = docSnap.DataTo(&s)
	return err
}

type FetchedDoc struct {
	FeedRef string `firestore:"feedRed"`
	Content string `firestore:"content"`
}

func (s *FetchedDoc) LoadFromDocRef(ctx context.Context, doc *firestore.DocumentRef) error {
	docSnap, err := doc.Get(ctx)
	if err != nil {
		return err
	}
	err = docSnap.DataTo(&s)
	return err
}

func getProjectID() string {
	switch env {
	case "PROD":
		return os.Getenv("GOOGLE_CLOUD_PROJECT")
	default:
		return "test-project"
	}

}

func getDBClient(ctx context.Context) *firestore.Client {
	dbClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore.NewClient: %v", err)

	}

	return dbClient
}

func getRequestData(req *http.Request) (RequestData, error) {
	data := RequestData{}
	err := json.NewDecoder(req.Body).Decode(&data)

	if err == nil && (stringIsEmpty(data.URL) || stringIsEmpty(data.DocumentID)) {
		err = errors.New(RequiredJSONFieldsNotFound)
	}

	return data, err
}

func writeResultOn(result Result, w http.ResponseWriter) {
	json.NewEncoder(w).Encode(result)
}

func unlockFeedDoc(ctx context.Context, d *firestore.DocumentRef) error {
	field := getFieldTag(FeedDoc{}, "Locked", "firestore")
	fieldsUpdate := []firestore.Update{{Path: field, Value: false}}

	_, err := d.Update(ctx, fieldsUpdate)
	return err
}

func lockFeedDoc(ctx context.Context, d *firestore.DocumentRef) error {
	field := getFieldTag(FeedDoc{}, "Locked", "firestore")
	fieldsUpdate := []firestore.Update{{Path: field, Value: true}}

	_, err := d.Update(ctx, fieldsUpdate)
	return err
}

func updateFeedLastFetchStatus(ctx context.Context, d *firestore.DocumentRef, status string) error {
	field := getFieldTag(FeedDoc{}, "LastStatus", "firestore")
	fieldsUpdate := []firestore.Update{{Path: field, Value: status}}

	_, err := d.Update(ctx, fieldsUpdate)
	return err
}

func updateFeedFetchTime(ctx context.Context, d *firestore.DocumentRef) error {
	field := getFieldTag(FeedDoc{}, "FetchedAt", "firestore")
	fieldsUpdate := []firestore.Update{{Path: field, Value: time.Now()}}

	_, err := d.Update(ctx, fieldsUpdate)
	return err
}

func createFetchedDoc(ctx context.Context, content FetchedDoc) (*firestore.DocumentRef, error) {
	doc, _, err := dbClient.Collection(FetchedCollection).Add(ctx, content)
	return doc, err
}

func init() {
	ctx := context.Background()
	projectID = getProjectID()

	if env == "PROD" {
		dbClient = getDBClient(ctx)
	}
}

func FetchURLAndStoreItsContent(w http.ResponseWriter, req *http.Request) {
	reqCtx := req.Context()
	result := Result{}

	if env == "TEST" {
		dbClient = getDBClient(reqCtx)
	}

	requestData, err := getRequestData(req)
	if err != nil {
		result.Status = err.Error()
		writeResultOn(result, w)
		return
	}

	feedDoc := dbClient.Collection(FeedsCollection).Doc(requestData.DocumentID)

	lockFeedDoc(reqCtx, feedDoc)
	defer unlockFeedDoc(reqCtx, feedDoc)

	fetchedContent, err := fetchURLBody(requestData.URL)
	if err != nil {
		result.Status = err.Error()
		writeResultOn(result, w)
		updateFeedLastFetchStatus(reqCtx, feedDoc, err.Error())
		return
	}
	result.Body = fetchedContent

	fetchedDoc, err := createFetchedDoc(reqCtx, FetchedDoc{FeedRef: requestData.DocumentID, Content: fetchedContent})
	if err != nil {
		result.Status = err.Error()
		writeResultOn(result, w)
		updateFeedLastFetchStatus(reqCtx, feedDoc, err.Error())
		return
	}
	result.FetchedDocID = fetchedDoc.ID

	updateFeedLastFetchStatus(reqCtx, feedDoc, LastStatusOk)
	updateFeedFetchTime(reqCtx, feedDoc)

	result.Status = LastStatusOk
	writeResultOn(result, w)
}
