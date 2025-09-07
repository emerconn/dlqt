package main

import (
	"context"
	"fmt"
	"log"

	"dlqt/internal/msal"

	"github.com/urfave/cli/v3"
)

func fetch(ctx context.Context, cmd *cli.Command) error {
	config := msal.Config{
		TenantID:  "f09f69e2-b684-4c08-9195-f8f10f54154c",
		ClientID:  "e64205cb-fe54-4452-a6ab-7eec472bdfcc",
		Scope:     "api://074c5ac1-4ab2-4a8a-b811-2d7b8c4e419f/dlq.read",
		CacheFile: "msal_cache.json",
	}

	token, err := msal.GetToken(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	log.Printf("acquired token: %s\n", token)

	return nil
}
