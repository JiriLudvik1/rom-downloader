package subscribing

import "time"

type RomUploadedMessage struct {
	MessageId string
	Bucket    string    `json:"bucket"`
	File      string    `json:"file"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
}
