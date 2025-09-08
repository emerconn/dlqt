package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"dlqt/internal/msal"

	"github.com/urfave/cli/v3"
)

func retrigger(ctx context.Context, cmd *cli.Command) error {
	// set configs
	msalConfig := msal.MSALConfig{
		TenantID:  cmd.String("cmd-tenant-id"),
		ClientID:  cmd.String("cmd-client-id"),
		Scope:     "api://" + cmd.String("api-client-id") + "/dlq.read",
		CacheFile: "msal_cache.json",
	}
	apiConfig := msal.APIConfig{
		APIEndpoint: cmd.String("api-url") + "/retrigger",
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

	// prepare JSON payload
	payload := map[string]string{
		"message-id": cmd.String("message-id"),
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// create request and auth header
	req, err := http.NewRequest("PATCH", fullURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

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

	log.Printf("response body: %s", string(body))

	// TODO: handle response body

	return nil
}
