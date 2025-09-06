package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/urfave/cli/v3"
)

type CheckAuthRequest struct {
	Namespace string `json:"namespace"`
	Queue     string `json:"queue"`
}

type CheckAuthResponse struct {
	Authorized bool   `json:"authorized"`
	UserID     string `json:"userID"`
	Namespace  string `json:"namespace"`
	Queue      string `json:"queue"`
}

func checkAuth(ctx context.Context, c *cli.Command) error {
	namespace := c.String("namespace")
	queue := c.String("queue")
	authServiceURL := c.String("auth-service-url")

	log.Println("namespace:", namespace)
	log.Println("queue:", queue)
	log.Println("auth service URL:", authServiceURL)

	// Get Azure AD token
	token, err := getAzureADToken()
	if err != nil {
		return fmt.Errorf("failed to get Azure AD token: %w", err)
	}

	// Prepare request to auth service
	reqBody := CheckAuthRequest{
		Namespace: namespace,
		Queue:     queue,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request to auth service
	req, err := http.NewRequestWithContext(ctx, "POST", authServiceURL+"/check-auth", bytes.NewBuffer(jsonData))
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		var authResp CheckAuthResponse
		if err := json.Unmarshal(body, &authResp); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		log.Printf("✅ Authorization successful!")
		log.Printf("   User ID: %s", authResp.UserID)
		log.Printf("   Namespace: %s", authResp.Namespace)
		log.Printf("   Queue: %s", authResp.Queue)
		return nil
	} else if resp.StatusCode == http.StatusForbidden {
		var authResp CheckAuthResponse
		if err := json.Unmarshal(body, &authResp); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		log.Printf("❌ Authorization failed!")
		log.Printf("   User ID: %s", authResp.UserID)
		log.Printf("   Namespace: %s", authResp.Namespace)
		log.Printf("   Queue: %s", authResp.Queue)
		return fmt.Errorf("user not authorized for this resource")
	} else {
		return fmt.Errorf("auth service returned status %d: %s", resp.StatusCode, string(body))
	}
}

func getAzureADToken() (string, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return "", fmt.Errorf("failed to create credential: %w", err)
	}

	// Request token using the app client ID as scope (will be output from Terraform)
	// TODO: Replace with actual client ID from Terraform output
	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{"795ab387-3200-4ae3-81e3-a096b787e155/.default"}, // This will be replaced with actual ID
	})
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	return token.Token, nil
}
