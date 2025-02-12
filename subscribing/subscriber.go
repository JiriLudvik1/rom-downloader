package subscribing

import (
	"cloud.google.com/go/pubsub"
	"context"
	"google.golang.org/api/option"
	"log"
	"rom-downloader/config"
)

func StartSubscriber(ctx *context.Context, config *config.LoaderConfig) {
	client, err := pubsub.NewClient(
		*ctx,
		config.ProjectID,
		option.WithCredentialsFile(config.CredentialsFileName),
	)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	sub := client.Subscription(config.SubscriptionName)
	err = sub.Receive(*ctx, func(ctx context.Context, m *pubsub.Message) {
		log.Printf("Received message: %s", m.Data)
		m.Ack()
	})
	if err != nil {
		log.Fatalf("Failed to receive message: %v", err)
	}
}
