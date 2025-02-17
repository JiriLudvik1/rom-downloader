package persistence

import (
	"rom-downloader/subscribing"
	"time"
)

type CompleteDownload struct {
	MessageId    string    `firestore:"messageId"`
	FileName     string    `firestore:"fileName"`
	BucketName   string    `firestore:"bucketName"`
	DownloadedAt time.Time `firestore:"downloadedAt"`
	IsDeleted    bool      `firestore:"isDeleted"`
}

func CompleteDownloadFromMessage(msg *subscribing.RomUploadedMessage) *CompleteDownload {
	return &CompleteDownload{
		MessageId:    msg.MessageId,
		FileName:     msg.File,
		BucketName:   msg.Bucket,
		DownloadedAt: time.Now().UTC(),
		IsDeleted:    false,
	}
}
