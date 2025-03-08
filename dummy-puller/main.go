package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Variables struct {
	ProjectID       string
	SubscriptionID  string
	maxNackAttempts int
	messageAgeLimit time.Duration
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {

		ctx := context.Background()
		variables, err := getEnvVariables()
		if err != nil {
			panic(err)
		}

		client, err := pubsub.NewClient(ctx, variables.ProjectID)
		if err != nil {
			panic(err)
		}
		defer client.Close()

		sub := client.Subscription(variables.SubscriptionID)
		fmt.Printf("Listening for messages on subscription %s...\n", variables.SubscriptionID)

		err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			msgAge := time.Since(msg.PublishTime)
			fmt.Printf("Received message: %s | Age: %v\n", msg.ID, msgAge)

			if msgAge > variables.messageAgeLimit {
				fmt.Printf("Message %s is older than %f minutes. Forcing Dead Letter...\n", msg.ID, variables.messageAgeLimit.Minutes())
				for i := 0; i < variables.maxNackAttempts; i++ {
					msg.Nack()
				}
				fmt.Printf("Message %s has been nacked %d forced to Dead Letter.\n", msg.ID, variables.maxNackAttempts)
				return
			}
			fmt.Printf("Message %s is younger than 10 minutes, ignoring it\n", msg.ID)
		})
		if err != nil {
			panic(err)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Server is running and listening on Cloud Run")
	})

	log.Printf("Server is listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

}

func getEnvVariables() (Variables, error) {
	projectId := os.Getenv("PROJECT_ID")
	if projectId == "" {
		return Variables{}, errors.New("PROJECT_ID environment variable is not set")
	}
	subscriptionID := os.Getenv("SUBSCRIPTION_ID")
	if subscriptionID == "" {
		return Variables{}, errors.New("SUBSCRIPTION_ID environment variable is not set")
	}

	maxNackAttemptsStr := os.Getenv("MAX_NACK_ATTEMPTS")
	maxNackAttempts, err := strconv.Atoi(maxNackAttemptsStr)
	if err != nil {
		maxNackAttempts = 5
	}

	messageAgeLimitStr := os.Getenv("MESSAGE_AGE_LIMIT")
	messageAgeLimitMinutes, err := strconv.Atoi(messageAgeLimitStr)
	if err != nil {
		messageAgeLimitMinutes = 10
	}
	messageAgeLimit := time.Duration(messageAgeLimitMinutes) * time.Minute

	return Variables{
		ProjectID:       projectId,
		SubscriptionID:  subscriptionID,
		maxNackAttempts: maxNackAttempts,
		messageAgeLimit: messageAgeLimit,
	}, nil
}
