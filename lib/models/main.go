package models

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
)

const FeedsCollection = "feeds"
const FetchedCollection = "fetched_feeds"

type Doc interface {
	LoadFromDocRef()
}

type FeedDoc struct {
	URL             string    `firestore:"url"`
	CreatedAt       time.Time `firestore:"createdAt"`
	FetchedAt       time.Time `firestore:"fetchedAt"`
	LastFetchStatus string    `firestore:"lastFetchStatus"`
	Locked          bool      `firestore:"currentlyFetched"`
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
	FeedRef string `firestore:"feedRef"`
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
