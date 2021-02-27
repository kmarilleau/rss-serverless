package fetcher

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
)

const FeedsCollection = "feeds"
const FetchedCollection = "fetched_feeds"

const RequiredJSONFieldsNotFound = "Required JSON fields not found"
const lastStatusOk = "OK"

type RequestData struct {
	Environment string `json:"environment"`
	DocumentID  string `json:"documentId"`
	URL         string `json:"url"`
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
	Locked     bool   `firestore:"currentlyFetched"`
	LastStatus string `firestore:"lastFetchStatus"`
	FetchedAt  bool   `firestore:"fetchedAt"`
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

func getProjectID(environment string) string {
	switch environment {
	case "test":
		return "test"
	default:
		return os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

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

func createFetchedDoc(ctx context.Context, dbClient *firestore.Client, content FetchedDoc) (*firestore.DocumentRef, error) {
	doc, _, err := dbClient.Collection(FetchedCollection).Add(ctx, content)
	return doc, err
}

func FetchURLAndStoreItsContent(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	result := Result{}

	requestData, err := getRequestData(req)
	if err != nil {
		result.Status = err.Error()
		writeResultOn(result, w)
		return
	}

	projectID := getProjectID(requestData.Environment)

	dbClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		result.Status = err.Error()
		writeResultOn(result, w)
		return
	}
	defer dbClient.Close()

	feedDoc := dbClient.Collection(FeedsCollection).Doc(requestData.DocumentID)

	lockFeedDoc(ctx, feedDoc)
	defer unlockFeedDoc(ctx, feedDoc)

	fetchedContent, err := fetchURLBody(requestData.URL)
	result.Body = fetchedContent
	if err != nil {
		result.Status = err.Error()
		writeResultOn(result, w)
		updateFeedLastFetchStatus(ctx, feedDoc, err.Error())
		return
	}

	fetchedDoc, err := createFetchedDoc(ctx, dbClient, FetchedDoc{FeedRef: requestData.DocumentID, Content: fetchedContent})
	result.FetchedDocID = fetchedDoc.ID
	if err != nil {
		result.Status = err.Error()
		writeResultOn(result, w)
		updateFeedLastFetchStatus(ctx, feedDoc, err.Error())
		return
	}

	updateFeedLastFetchStatus(ctx, feedDoc, lastStatusOk)
	updateFeedFetchTime(ctx, feedDoc)

	result.Status = lastStatusOk
	writeResultOn(result, w)
}
