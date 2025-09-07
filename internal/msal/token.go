package msal

import (
	"context"
	"fmt"
	"log"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

type MSALConfig struct {
	TenantID  string
	ClientID  string
	Scope     string
	CacheFile string // Add this for configurability
}

type APIConfig struct {
	APIEndpoint string
}

func GetToken(ctx context.Context, config *MSALConfig) (string, error) {
	// set up cache to persist tokens
	cacheAccessor := NewCacheAccessor(config.CacheFile)

	// create MSAL client
	client, err := public.New(config.ClientID, public.WithAuthority(fmt.Sprintf("https://login.microsoftonline.com/%s", config.TenantID)), public.WithCache(cacheAccessor))
	if err != nil {
		return "", fmt.Errorf("failed to create MSAL public client: %w", err)
	}

	// check for cached accounts
	accounts, err := client.Accounts(ctx)
	if err != nil {
		log.Println("no cached accounts, proceeding to interactive login")
	} else if len(accounts) > 0 {
		// silent token acquisition using the first account
		result, err := client.AcquireTokenSilent(ctx, []string{config.Scope}, public.WithSilentAccount(accounts[0]))
		if err == nil {
			return result.AccessToken, nil
		}
		log.Println("silent acquisition failed, proceeding to interactive")
	}

	// interactive token acquisition
	result, err := client.AcquireTokenInteractive(ctx, []string{config.Scope})
	if err != nil {
		return "", fmt.Errorf("failed to acquire token interactively: %w", err)
	}
	return result.AccessToken, nil
}
