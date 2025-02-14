package subscribing

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"google.golang.org/api/option"
	"log"
	"rom-downloader/config"
)

func StartSubscriber(
	ctx context.Context,
	config *config.LoaderConfig,
	messages chan<- RomUploadedMessage,
) {
	client, err := pubsub.NewClient(
		ctx,
		config.ProjectID,
		option.WithCredentialsFile(config.CredentialsFileName),
	)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Error closing client: %v", err)
		}
	}()

	sub := client.Subscription(config.SubscriptionName)
	log.Printf("Created subscriber for subscription: %s", sub.ID())

	err = sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		log.Printf("Received message: [%s] %s", m.ID, m.Data)
		var message RomUploadedMessage
		if err := json.Unmarshal(m.Data, &message); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			m.Nack()
			return
		}
		message.MessageId = m.ID
		messages <- message
		m.Ack()
	})
	if err != nil {
		log.Fatalf("Failed to receive message: %v", err)
	}
}
