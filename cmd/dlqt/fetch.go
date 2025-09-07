package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"dlqt/internal/msal"

	"github.com/urfave/cli/v3"
)

func fetch(ctx context.Context, cmd *cli.Command) error {
	msalConfig := msal.MSALConfig{
		TenantID:  "f09f69e2-b684-4c08-9195-f8f10f54154c",
		ClientID:  "e64205cb-fe54-4452-a6ab-7eec472bdfcc",
		Scope:     "api://074c5ac1-4ab2-4a8a-b811-2d7b8c4e419f/dlq.read",
		CacheFile: "msal_cache.json",
	}
	apiConfig := msal.APIConfig{
		APIEndpoint: cmd.String("api-url") + "/fetch",
	}

	// get JWT
	token, err := msal.GetToken(ctx, &msalConfig)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	log.Printf("token: %s\n", token)

	// create request and auth header
	req, err := http.NewRequest("GET", apiConfig.APIEndpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json") // Optional, adjust as needed

	// TODO: add request body with service bus message ID

	// execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch data: %s", resp.Status)
	}

	log.Println("request sent successfully")

	// TODO: handle response body

	return nil
}
