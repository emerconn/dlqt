package servicebus

import (
	"time"
)

// JSON-serializable version of a Service Bus message
type MessageResponse struct {
	Namespace                  string         `json:"namespace"`
	Queue                      string         `json:"queue"`
	MessageID                  string         `json:"messageID"`
	Body                       string         `json:"body"`
	ContentType                *string        `json:"contentType,omitempty"`
	CorrelationID              *string        `json:"correlationID,omitempty"`
	DeadLetterErrorDescription *string        `json:"deadLetterErrorDescription,omitempty"`
	DeadLetterReason           *string        `json:"deadLetterReason,omitempty"`
	DeadLetterSource           *string        `json:"deadLetterSource,omitempty"`
	DeliveryCount              uint32         `json:"deliveryCount"`
	EnqueuedSequenceNumber     *int64         `json:"enqueuedSequenceNumber,omitempty"`
	EnqueuedTime               *time.Time     `json:"enqueuedTime,omitempty"`
	ExpiresAt                  *time.Time     `json:"expiresAt,omitempty"`
	LockedUntil                *time.Time     `json:"lockedUntil,omitempty"`
	PartitionKey               *string        `json:"partitionKey,omitempty"`
	ReplyTo                    *string        `json:"replyTo,omitempty"`
	ReplyToSessionID           *string        `json:"replyToSessionID,omitempty"`
	ScheduledEnqueueTime       *time.Time     `json:"scheduledEnqueueTime,omitempty"`
	SequenceNumber             *int64         `json:"sequenceNumber,omitempty"`
	SessionID                  *string        `json:"sessionID,omitempty"`
	State                      int32          `json:"state"`
	Subject                    *string        `json:"subject,omitempty"`
	TimeToLive                 *time.Duration `json:"timeToLive,omitempty"`
	To                         *string        `json:"to,omitempty"`
	ApplicationProperties      map[string]any `json:"applicationProperties,omitempty"`
}
