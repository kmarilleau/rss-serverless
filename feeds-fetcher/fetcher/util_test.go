package fetcher

import (
	"context"
	"log"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
)

// Firestore
func getAllCollections(ctx context.Context, dbClient *firestore.Client) (c []*firestore.CollectionRef) {
	c, err := dbClient.Collections(ctx).GetAll()
	if err != nil {
		log.Fatal(err)
	}

	return
}

func getAllDocuments(ctx context.Context, c *firestore.CollectionRef) (d []*firestore.DocumentRef) {
	d, err := c.DocumentRefs(ctx).GetAll()
	if err != nil {
		log.Fatal(err)
	}

	return
}

func clearAllFirestoreData(ctx context.Context, dbClient *firestore.Client) {
	for _, collection := range getAllCollections(ctx, dbClient) {
		for _, document := range getAllDocuments(ctx, collection) {
			if _, err := document.Delete(ctx); err != nil {
				log.Fatal(err)
			}
		}
	}
}

// Others
func httpErrorStartsWithStatusCode(err error, statusCode int) bool {
	statusCodeStr := strconv.Itoa(statusCode)
	errStr := err.Error()

	return strings.HasPrefix(errStr, statusCodeStr)
}
