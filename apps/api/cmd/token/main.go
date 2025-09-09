package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"
)

type config struct {
	Auth0 struct {
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		Audience     string `yaml:"audience"`
		GrantType    string `yaml:"grant_type"`
	} `yaml:"auth0"`
}

// make requestPayload a struct type
type requestPayload struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Audience     string `json:"audience"`
	GrantType    string `json:"grant_type"`
}

func main() {

	url := "https://dev-sleeping-pandas.us.auth0.com/oauth/token"

	cfg := loadConfig()

	auth0Config := &requestPayload{
		ClientID:     cfg.Auth0.ClientID,
		ClientSecret: cfg.Auth0.ClientSecret,
		Audience:     cfg.Auth0.Audience,
		GrantType:    cfg.Auth0.GrantType,
	}

	// convert auth0Config to json
	payload, err := json.Marshal(auth0Config)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			panic(err)
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}

func loadConfig() config {
	var cfg config
	data, err := os.ReadFile("secrets.yml")
	if err != nil {
		log.Fatal(err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}
	return cfg
}
