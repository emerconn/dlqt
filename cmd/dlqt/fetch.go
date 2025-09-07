package main

import (
	"context"
	"fmt"
	"log"

	"dlqt/internal/msal"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/urfave/cli/v3"
)

func fetch(ctx context.Context, cmd *cli.Command) error {
	tenantID := "f09f69e2-b684-4c08-9195-f8f10f54154c"
	clientID := "e64205cb-fe54-4452-a6ab-7eec472bdfcc"
	scope := "api://074c5ac1-4ab2-4a8a-b811-2d7b8c4e419f/dlq.read"

	// set up cache to persist tokens
	cacheFile := "msal_cache.json"
	cacheAccessor := msal.NewCacheAccessor(cacheFile)

	// create MSAL client
	client, err := public.New(clientID, public.WithAuthority(fmt.Sprintf("https://login.microsoftonline.com/%s", tenantID)), public.WithCache(cacheAccessor))
	if err != nil {
		return fmt.Errorf("failed to create MSAL public client: %w", err)
	}

	// check for cached accounts
	accounts, err := client.Accounts(ctx)
	if err != nil {
		log.Println("no cached accounts, proceeding to interactive login")
	} else if len(accounts) > 0 {
		// silent token acquisition using the first account
		result, err := client.AcquireTokenSilent(ctx, []string{scope}, public.WithSilentAccount(accounts[0]))
		if err == nil {
			token := result.AccessToken
			log.Printf("token acquired silently: %s", token)
			return nil
		}
		log.Println("silent acquisition failed, proceeding to interactive")
	}

	// interactive token acquisition
	result, err := client.AcquireTokenInteractive(ctx, []string{scope})
	if err != nil {
		return fmt.Errorf("failed to acquire token interactively: %w", err)
	}

	token := result.AccessToken
	log.Printf("token acquired interactively: %s", token)

	return nil
}
