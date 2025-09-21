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
		Version:                "v0.3.0",
		Usage:                  "Developer tool for interacting with Azure Service Bus DLQ",
		EnableShellCompletion:  true,
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "namespace",
				Aliases:  []string{"n"},
				Usage:    "the Service Bus namespace",
				Sources:  cli.EnvVars("AZURE_SERVICEBUS_NAMESPACE"),
				Required: true,
			},
			&cli.StringFlag{
				Name:     "queue",
				Aliases:  []string{"q"},
				Usage:    "the Service Bus queue name",
				Sources:  cli.EnvVars("AZURE_SERVICEBUS_QUEUE"),
				Required: true,
			},
			&cli.StringFlag{
				Name:     "api-url",
				Usage:    "the API service URL",
				Sources:  cli.EnvVars("API_URL"),
				Required: false,
			},
			&cli.StringFlag{
				Name:     "api-client-id",
				Usage:    "the API client ID",
				Sources:  cli.EnvVars("API_AZURE_CLIENT_ID"),
				Required: false,
			},
			&cli.StringFlag{
				Name:     "cmd-client-id",
				Usage:    "the CMD client ID",
				Sources:  cli.EnvVars("CMD_AZURE_CLIENT_ID"),
				Required: false,
			},
			&cli.StringFlag{
				Name:     "cmd-tenant-id",
				Usage:    "the CMD tenant ID",
				Sources:  cli.EnvVars("CMD_AZURE_TENANT_ID"),
				Required: false,
			},
		},
		Commands: []*cli.Command{
			// seed
			{
				Name:  "seed",
				Usage: "Seed the dead-letter queue with test messages",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return seedMessages(ctx, cmd)
				},
				MutuallyExclusiveFlags: []cli.MutuallyExclusiveFlags{
					{
						Required: true,
						Flags: [][]cli.Flag{
							{
								&cli.IntFlag{
									Name:     "num-messages",
									Aliases:  []string{"m"},
									Usage:    "the number of messages to send",
									Required: false,
									Action: func(ctx context.Context, cmd *cli.Command, v int) error {
										if v <= 0 {
											return fmt.Errorf("num-messages must be greater than 0, got %d", v)
										} else if v > 2048 {
											return fmt.Errorf("num-messages must be less than or equal to 2048, got %d", v)
										}
										return nil
									},
								},
							},
						},
					},
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:     "no-dlq",
						Usage:    "do not dead-letter messages",
						Required: false,
					},
				},
			},
			// purge
			{
				Name:  "purge",
				Usage: "Purge the queue and dead-letter queue of all messages",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return purgeMessages(ctx, cmd)
				},
				MutuallyExclusiveFlags: []cli.MutuallyExclusiveFlags{
					{
						Required: false,
						Flags: [][]cli.Flag{
							{
								&cli.BoolFlag{
									Name:     "no-queue",
									Usage:    "do not purge the normal queue",
									Required: false,
								},
							},
							{
								&cli.BoolFlag{
									Name:     "no-dlq",
									Usage:    "do not purge the dead-letter queue",
									Required: false,
								},
							},
						},
					},
				},
			},
			// fetch
			{
				Name:  "fetch",
				Usage: "Fetch one message from the dead letter queue",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return fetch(ctx, cmd)
				},
			},
			// retrigger
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
