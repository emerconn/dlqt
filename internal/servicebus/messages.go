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

func RetriggerDeadLetterMessage(ctx context.Context, client *azservicebus.Client, queue string, messageID string) error {
	// Create receiver for dead-letter queue
	options := &azservicebus.ReceiverOptions{
		SubQueue: azservicebus.SubQueueDeadLetter,
	}
	receiver, err := client.NewReceiverForQueue(queue, options)
	if err != nil {
		return fmt.Errorf("failed to create DLQ receiver for queue '%s': %w", queue, err)
	}
	defer receiver.Close(ctx)

	// Create sender for main queue
	sender, err := client.NewSender(queue, nil)
	if err != nil {
		return fmt.Errorf("failed to create sender for queue '%s': %w", queue, err)
	}
	defer sender.Close(ctx)

	// Receive messages in batches until we find the specific message
	batchSize := 10
	maxBatches := 100 // Limit to avoid infinite loop
	for range maxBatches {
		messages, err := receiver.ReceiveMessages(ctx, batchSize, nil)
		if err != nil {
			return fmt.Errorf("failed to receive messages from DLQ: %w", err)
		}

		if len(messages) == 0 {
			break // No more messages
		}

		for _, message := range messages {
			if message.MessageID == messageID {
				// Found the message, create new message with same body
				newMessage := &azservicebus.Message{
					Body: message.Body,
				}

				// Send to main queue
				err = sender.SendMessage(ctx, newMessage, nil)
				if err != nil {
					return fmt.Errorf("failed to send retriggered message: %w", err)
				}

				// Complete the original DLQ message
				err = receiver.CompleteMessage(ctx, message, nil)
				if err != nil {
					return fmt.Errorf("failed to complete DLQ message: %w", err)
				}

				log.Printf("Successfully retriggered message %s from DLQ to main queue", messageID)
				return nil
			} else {
				// Not the target message, abandon it to put back in DLQ
				err = receiver.AbandonMessage(ctx, message, nil)
				if err != nil {
					log.Printf("failed to abandon message %s: %v", message.MessageID, err)
					// Continue anyway
				}
			}
		}
	}

	return fmt.Errorf("message with ID '%s' not found in DLQ for queue '%s' after checking %d messages", messageID, queue, batchSize*maxBatches)
}

// FetchDeadLetterMessage fetches one message from the dead letter queue
func FetchDeadLetterMessage(ctx context.Context, client *azservicebus.Client, queue string) (*azservicebus.ReceivedMessage, error) {
	// Create receiver for dead-letter queue
	options := &azservicebus.ReceiverOptions{
		SubQueue: azservicebus.SubQueueDeadLetter,
	}
	receiver, err := client.NewReceiverForQueue(queue, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create DLQ receiver for queue '%s': %w", queue, err)
	}
	defer receiver.Close(ctx)

	// Receive one message from DLQ
	messages, err := receiver.ReceiveMessages(ctx, 1, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages from DLQ: %w", err)
	}

	if len(messages) == 0 {
		return nil, nil // No messages available
	}

	message := messages[0]
	log.Printf("Fetched message %s from DLQ", message.MessageID)

	// Complete the message to remove it from the DLQ
	// err = receiver.CompleteMessage(ctx, message, nil)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to complete DLQ message: %w", err)
	// }

	return message, nil
}
