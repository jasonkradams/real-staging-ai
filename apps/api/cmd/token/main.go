package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func main() {

	url := "https://dev-sleeping-pandas.us.auth0.com/oauth/token"

	payload := strings.NewReader("{\"client_id\":\"REDACTED_AUTH0_CLIENT_ID\",\"client_secret\":\"REDACTED_AUTH0_CLIENT_SECRET\",\"audience\":\"https://api.virtualstaging.local\",\"grant_type\":\"client_credentials\"}")

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		panic(err)
	}

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	// fmt.Println(res)
	fmt.Println(string(body))
}
