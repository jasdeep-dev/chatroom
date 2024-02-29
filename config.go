package main

import (
	"encoding/json"
	"os"
)

var Settings Config

type Config struct {
	MaxUsers           int    `json:"MaxUsers"`
	MessageSize        int    `json:"MessageSize"`
	DefaultTheme       string `json:"DefaultTheme"`
	HttpServer         string `json:"HttpServer"`
	JoinedMessage      string `json:"JoinedMessage"`
	WelcomeMessage     string `json:"WelcomeMessage"`
	WelcomeBackMessage string `json:"WelcomeBackMessage"`
}

func readConfigFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Parse JSON data into Config struct
	err = json.Unmarshal(data, &Settings)
	if err != nil {
		return err
	}

	return nil
}
