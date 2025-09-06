package servicebus

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

func SendMessageBatch(ctx context.Context, client *azservicebus.Client, queue string, messages []string) error {
	log.Println("creating sender")
	sender, err := client.NewSender(queue, nil)
	if err != nil {
		return fmt.Errorf("failed to create sender for queue '%s': %w", queue, err)
	}
	defer sender.Close(ctx)

	log.Println("creating message batch")
	batch, err := sender.NewMessageBatch(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to create message batch: %w", err)
	}

	log.Printf("adding %d messages to batch", len(messages))
	for _, message := range messages {
		log.Printf("adding message to batch: %s", message)
		err := batch.AddMessage(&azservicebus.Message{Body: []byte(message)}, nil)

		if errors.Is(err, azservicebus.ErrMessageTooLarge) {
			log.Print("message batch is full")
		}
	}

	log.Printf("sending batch of %d messages", len(messages))
	if err := sender.SendMessageBatch(ctx, batch, nil); err != nil {
		return fmt.Errorf("failed to send message batch: %w", err)
	}

	log.Printf("successfully sent batch of %d messages", len(messages))
	return nil
}

func DeadLetterMessages(ctx context.Context, client *azservicebus.Client, queue string, count int) error {
	deadLetterOptions := &azservicebus.DeadLetterOptions{
		ErrorDescription: to.Ptr("exampleErrorDescription"),
		Reason:           to.Ptr("exampleReason"),
	}

	log.Println("creating receiver")
	receiver, err := client.NewReceiverForQueue(queue, nil)
	if err != nil {
		return fmt.Errorf("failed to create receiver for queue '%s': %w", queue, err)
	}
	defer receiver.Close(ctx)

	// dead-letter messages in batches until count is reached
	maxMessages := 100
	receivedMessages := 0
	for receivedMessages < count {
		log.Printf("receiving %d messages", maxMessages)
		messages, err := receiver.ReceiveMessages(ctx, maxMessages, nil)
		if err != nil {
			return fmt.Errorf("failed to receive messages: %w", err)
		}

		log.Printf("received %d messages", len(messages))
		receivedMessages += len(messages)

		log.Printf("dead-lettering messages")
		for _, message := range messages {
			log.Printf("dead-lettering message: %s", message.MessageID)
			err := receiver.DeadLetterMessage(ctx, message, deadLetterOptions)
			if err != nil {
				return fmt.Errorf("failed to dead-letter message '%s': %w", message.MessageID, err)
			}
		}
	}

	log.Printf("dead-lettered %d messages", receivedMessages)
	return nil
}
