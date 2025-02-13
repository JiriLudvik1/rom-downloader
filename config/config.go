package config

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

type LoaderConfig struct {
	CredentialsFileName   string            `json:"credentialsFileName"`
	GoogleDriveFolderId   string            `json:"googleDriveFolderId"`
	SubscriptionName      string            `json:"subscriptionName"`
	TopicName             string            `json:"topicName"`
	ProjectID             string            `json:"projectId"`
	TempFolder            string            `json:"tempFolder"`
	DestinationFolderRoot string            `json:"destinationFolderRoot"`
	RomTypeDestinations   map[string]string `json:"romTypeDestinations"`
}

func GetConfiguration() (*LoaderConfig, error) {
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		return nil, errors.New("config.json file not found")
	}

	configFile, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer func() {
		err = configFile.Close()
		if err != nil {
			log.Printf("Error closing config file: %v", err)
		}
	}()

	decoder := json.NewDecoder(configFile)
	config := &LoaderConfig{}
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
