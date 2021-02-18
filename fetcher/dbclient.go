package fetcher

import (
	"cloud.google.com/go/firestore"
	"context"
	"log"
)

func createClient(ctx context.Context, projectID string) (client *firestore.Client) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return
}
