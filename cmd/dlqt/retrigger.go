package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/urfave/cli/v3"
)

type RetriggerRequest struct {
	Queue     string `json:"queue"`
	MessageID string `json:"messageId"`
}

func retriggerMessage(ctx context.Context, c *cli.Command) error {
	namespace := c.String("namespace")
	queue := c.String("queue")
	messageID := c.String("message-id")
	authServiceURL := c.String("auth-service-url")

	log.Println("namespace:", namespace)
	log.Println("queue:", queue)
	log.Println("message ID:", messageID)
	log.Println("auth service URL:", authServiceURL)

	// Get Azure AD token
	token, err := getAzureADToken()
	if err != nil {
		return fmt.Errorf("failed to get Azure AD token: %w", err)
	}

	// Prepare request to auth service
	reqBody := RetriggerRequest{
		Queue:     queue,
		MessageID: messageID,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request to auth service
	req, err := http.NewRequestWithContext(ctx, "POST", authServiceURL+"/retrigger", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call auth service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth service returned status %d: %s", resp.StatusCode, string(body))
	}

	log.Println("Message retriggered successfully")
	return nil
}

func getAzureADToken() (string, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return "", fmt.Errorf("failed to create credential: %w", err)
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	return token.Token, nil
}
