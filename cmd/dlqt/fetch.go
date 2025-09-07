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
		TenantID:  cmd.String("cmd-tenant-id"),
		ClientID:  cmd.String("cmd-client-id"),
		Scope:     "api://" + cmd.String("api-client-id") + "/dlq.read",
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
