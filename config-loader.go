package main

import (
	"encoding/json"
	"errors"
	"os"
)

type LoaderConfig struct {
	CredentialsFileName   string            `json:"credentialsFileName"`
	GoogleDriveFolderId   string            `json:"googleDriveFolderId"`
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
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	config := &LoaderConfig{}
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
