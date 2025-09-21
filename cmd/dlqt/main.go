package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"
)

type timestampWriter struct{}

func (w *timestampWriter) Write(p []byte) (n int, err error) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	message := fmt.Sprintf("%s %s", timestamp, string(p))
	return os.Stderr.Write([]byte(message))
}

func main() {
	// use RFC3339 timestamp for log messages
	log.SetFlags(0)
	log.SetOutput(&timestampWriter{})

	// load env vars from .env file
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	cmd := &cli.Command{
		Name:                   "dlqt",
		Version:                "v0.2.4",
		Usage:                  "Developer tool for interacting with Azure Service Bus DLQ",
		EnableShellCompletion:  true,
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "namespace",
				Usage:    "the Service Bus namespace",
				Sources:  cli.EnvVars("AZURE_SERVICEBUS_NAMESPACE"),
				Required: true,
			},
			&cli.StringFlag{
				Name:     "queue",
				Usage:    "the Service Bus queue name",
				Sources:  cli.EnvVars("AZURE_SERVICEBUS_QUEUE"),
				Required: true,
			},
			&cli.StringFlag{
				Name:     "api-url",
				Usage:    "the API service URL",
				Sources:  cli.EnvVars("API_URL"),
				Required: true,
			},
			&cli.StringFlag{
				Name:     "api-client-id",
				Usage:    "the API client ID",
				Sources:  cli.EnvVars("API_AZURE_CLIENT_ID"),
				Required: true,
			},
			&cli.StringFlag{
				Name:     "cmd-client-id",
				Usage:    "the CMD client ID",
				Sources:  cli.EnvVars("CMD_AZURE_CLIENT_ID"),
				Required: true,
			},
			&cli.StringFlag{
				Name:     "cmd-tenant-id",
				Usage:    "the CMD tenant ID",
				Sources:  cli.EnvVars("CMD_AZURE_TENANT_ID"),
				Required: true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "fetch",
				Usage: "Fetch one message from the dead letter queue",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return fetch(ctx, cmd)
				},
			},
			{
				Name:  "retrigger",
				Usage: "Retrigger one message from the dead letter queue",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return retrigger(ctx, cmd)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "message-id",
						Usage:    "the message ID to retrigger",
						Required: true,
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
