package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type LoaderConfig struct {
	CredentialsFileName   string            `json:"credentialsFileName"`
	SubscriptionName      string            `json:"subscriptionName"`
	TopicName             string            `json:"topicName"`
	ProjectID             string            `json:"projectId"`
	TempFolder            string            `json:"tempFolder"`
	DestinationFolderRoot string            `json:"destinationFolderRoot"`
	RomTypeDestinations   map[string]string `json:"romTypeDestinations"`
}

const configFileName = "config.json"

func GetConfiguration() (*LoaderConfig, error) {
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		return nil, fmt.Errorf(
			"config file %s does not exist, please create it",
			configFileName)
	}

	configFile, err := os.Open(configFileName)
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

	err = validateConfig(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func validateConfig(config *LoaderConfig) error {
	var missingFields []string
	if config.CredentialsFileName == "" {
		missingFields = append(missingFields, "credentialsFileName")
	}

	if config.SubscriptionName == "" {
		missingFields = append(missingFields, "subscriptionName")
	}

	if config.TopicName == "" {
		missingFields = append(missingFields, "topicName")
	}

	if config.ProjectID == "" {
		missingFields = append(missingFields, "projectId")
	}

	if config.TempFolder == "" {
		missingFields = append(missingFields, "tempFolder")
	}

	if config.DestinationFolderRoot == "" {
		missingFields = append(missingFields, "destinationFolderRoot")
	}

	if len(missingFields) > 0 {
		return errors.New("missing fields: " + strings.Join(missingFields, ", "))
	}
	return nil
}
