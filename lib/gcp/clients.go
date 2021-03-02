package gcp

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
)

func GetDBClient(ctx context.Context, projectID string) *firestore.Client {
	dbClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore.NewClient: %v", err)

	}

	return dbClient
}

func GetPubSubClient(ctx context.Context, projectID string) *pubsub.Client {
	pubSubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("pubsub.NewClient: %v", err)
	}

	return pubSubClient
}
