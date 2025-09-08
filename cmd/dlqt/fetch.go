package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

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
	req.Header.Set("Content-Type", "application/json") // Optional, adjust as needed

	// execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		return fmt.Errorf("failed to fetch data: %d %s", resp.StatusCode, string(body))
	}

	log.Println("request sent successfully")

	// TODO: handle response body

	return nil
}
