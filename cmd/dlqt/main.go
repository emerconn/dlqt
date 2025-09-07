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
				Aliases:  []string{"u"},
				Usage:    "the API service URL",
				Sources:  cli.EnvVars("DLQT_API_URL"),
				Required: true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "fetch",
				Usage: "Fetch one message from the dead letter queue",
				Action: func(ctx context.Context, c *cli.Command) error {
					return fetch(ctx, c)
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
