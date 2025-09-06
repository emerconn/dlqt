package servicebus

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

func purgeQueueWithOptions(ctx context.Context, client *azservicebus.Client, queue string, options *azservicebus.ReceiverOptions, queueType string) error {
	log.Printf("creating %s receiver", queueType)
	receiver, err := client.NewReceiverForQueue(queue, options)
	if err != nil {
		return err
	}
	defer receiver.Close(ctx)

	totalPurged := 0
	batchSize := 100
	for {
		// check if any messages exist
		log.Printf("checking for messages in %s", queueType)
		peekedMessages, err := receiver.PeekMessages(ctx, 1, nil) // peek one message
		if err != nil {
			return fmt.Errorf("failed to peek messages from %s: %w", queueType, err)
		}
		if len(peekedMessages) == 0 {
			log.Printf("no messages found in %s", queueType)
			break
		}
		log.Printf("found messages in %s", queueType)

		// receive messages
		log.Printf("receiving messages from %s with batch size %d", queueType, batchSize)
		messages, err := receiver.ReceiveMessages(ctx, batchSize, nil)
		if err != nil {
			return fmt.Errorf("failed to receive messages from %s: %w", queueType, err)
		}
		log.Printf("received batch of %d messages from %s", len(messages), queueType)

		// complete received messages
		for _, message := range messages {
			err := receiver.CompleteMessage(ctx, message, nil)
			if err != nil {
				return fmt.Errorf("failed to complete message %s: %w", message.MessageID, err)
			} else {
				totalPurged++
			}
		}
		log.Printf("purged batch of %d messages from %s", len(messages), queueType)
	}

	if totalPurged > 0 {
		log.Printf("purged %d messages from %s", totalPurged, queueType)
	}
	return nil
}

func PurgeQueue(ctx context.Context, client *azservicebus.Client, queue string) error {
	return purgeQueueWithOptions(ctx, client, queue, nil, "queue")
}

func PurgeDeadLetterQueue(ctx context.Context, client *azservicebus.Client, queue string) error {
	options := &azservicebus.ReceiverOptions{
		SubQueue: azservicebus.SubQueueDeadLetter,
	}
	return purgeQueueWithOptions(ctx, client, queue, options, "dead-letter queue")
}
