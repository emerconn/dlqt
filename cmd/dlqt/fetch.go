package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"

	"dlqt/internal/msal"
	"dlqt/internal/servicebus"

	"github.com/urfave/cli/v3"
)

func fetch(ctx context.Context, cmd *cli.Command) error {
	// set configs
	msalConfig := msal.MSALConfig{
		TenantID:  cmd.String("cmd-tenant-id"),
		ClientID:  cmd.String("cmd-client-id"),
		Scope:     "api://" + cmd.String("api-client-id") + "/dlq.read",
		CacheFile: "msal_cache.json",
	}
	apiConfig := msal.APIConfig{
		APIEndpoint: cmd.String("api-url") + "/fetch",
	}

	// add URL query parameters
	params := url.Values{}
	params.Add("namespace", cmd.String("namespace"))
	params.Add("queue", cmd.String("queue"))
	fullURL := apiConfig.APIEndpoint + "?" + params.Encode()

	// get JWT
	token, err := msal.GetToken(ctx, &msalConfig)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	log.Printf("token: %s\n", token)

	// create request and auth header
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// check HTTP status code
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read error response body: %w", err)
		}
		return fmt.Errorf("failed to fetch data: %s", string(body))
	}

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// unescape HTML entities in the response
	unescapedBody := html.UnescapeString(string(body))
	log.Printf("unescaped response body: %s", unescapedBody)

	// parse JSON response
	var message servicebus.MessageResponse
	err = json.Unmarshal([]byte(unescapedBody), &message)
	if err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// TODO: handle parsed message

	return nil
}
