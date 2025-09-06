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
		Name:                   "dlqtools",
		Usage:                  "Admin crud for the dlqt CLI tool",
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
		},
		Commands: []*cli.Command{
			{
				Name:  "seed",
				Usage: "Seed the dead-letter queue with test messages",
				Action: func(ctx context.Context, c *cli.Command) error {
					return seedMessages(ctx, c)
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
											return fmt.Errorf("num-messages value '%v' must be greater than 0", v)
										}
										return nil
									},
								},
							},
							{
								&cli.StringFlag{
									Name:     "json-messages",
									Usage:    "a JSON file containing messages to send",
									Required: false,
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
			{
				Name:  "purge",
				Usage: "Purge the queue and dead-letter queue of all messages",
				Action: func(ctx context.Context, c *cli.Command) error {
					return purgeMessages(ctx, c)
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
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
