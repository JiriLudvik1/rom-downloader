package persistence

import (
	"cloud.google.com/go/firestore"
	"context"
	"google.golang.org/api/option"
	"log"
	"rom-downloader/config"
)

type FirestoreService struct {
	client *firestore.Client
	ctx    context.Context
	config *config.LoaderConfig
}

const completeDownloadCollection = "complete"

func NewFirestoreService(ctx context.Context, config *config.LoaderConfig) (*FirestoreService, error) {
	client, err := firestore.NewClient(ctx, config.ProjectID, option.WithCredentialsFile(config.CredentialsFileName))
	if err != nil {
		return nil, err
	}

	service := &FirestoreService{client: client, ctx: ctx, config: config}
	go service.closeOnContext()
	return service, nil
}

func (s *FirestoreService) CreateCompleteDownloadDoc(download *CompleteDownload) error {
	err := s.writeDocument(completeDownloadCollection, nil, download)
	if err != nil {
		return err
	}

	log.Printf("Created success document to firestore for file %s", download.FileName)
	return nil
}

func (s *FirestoreService) writeDocument(collectionName string, documentId *string, data interface{}) error {
	var docRef *firestore.DocumentRef

	if documentId == nil {
		docRef = s.client.Collection(collectionName).NewDoc()
	} else {
		docRef = s.client.Collection(collectionName).Doc(*documentId)
	}

	_, err := docRef.Set(s.ctx, data)
	return err
}

func (s *FirestoreService) closeOnContext() {
	<-s.ctx.Done()

	err := s.client.Close()
	if err != nil {
		log.Printf("Error closing Firestore client: %v", err)
	} else {
		log.Println("Firestore client closed gracefully")
	}
}
