package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	DatabaseURL     string `json:"db_url"`
	CurrentUsername string `json:"current_user_name"`
}

// https://tutorialedge.net/golang/parsing-json-with-golang/

func (config *Config) SetUser(user string) error {
	(*config).CurrentUsername = user

	if err := write(config); err != nil {
		fmt.Println("Error writing Config struct to file: ", err)
		return err
	}
	return nil
}

func write(config *Config) error {
	byteValue, err := json.Marshal(config)
	if err != nil {
		fmt.Println("Error marshalling Config struct into JSON")
	}

	filepath, err := getConfigFilePath()
	if err != nil {
		fmt.Println("Error finding filepath for config file: ", err)
		return err
	}

	// os.Create - creates or truncates a file for writing
	jsonFile, err := os.Create(filepath)
	if err != nil {
		fmt.Println("Error opening file: ", err)
		return err
	}
	defer jsonFile.Close()

	writer := bufio.NewWriter(jsonFile)
	if _, err = writer.Write(byteValue); err != nil {
		fmt.Println("Error writing to file: ", err)
		return err
	}

	if err = writer.Flush(); err != nil {
		fmt.Println("Error flushing the writer object: ", err)
		return err
	}
	return nil
}

func Read() (Config, error) {
	filepath, err := getConfigFilePath()
	if err != nil {
		fmt.Println("Error finding filepath for config file: ", err)
		return Config{}, err
	}

	jsonFile, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Error opening file: ", err)
		return Config{}, err
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("Error reading file: ", err)
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return Config{}, err
	}

	return config, nil

	//decoder := json.NewDecoder()
	//if err := decoder.Decode(&config); err != nil {
	//	return Config{}, err
	//}
}

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	homeLocation, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error retrieving project directory homepath: ", err)
		return "", err
	}
	filepath := homeLocation + "/" + configFileName
	return filepath, nil
}
