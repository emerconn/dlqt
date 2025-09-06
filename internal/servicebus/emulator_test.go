package servicebus

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/servicebus"
)

const (
	queueName           = "queue1"
	testMessage         = "Hello, Testcontainers!"
	deadLetterMessage   = "Dead letter test message"
	maxRetries          = 3
	maxDeliveryCount    = 10
	maxDeliveryAttempts = 15
	receiveTimeout      = 5 * time.Second
	deadLetterTimeout   = 10 * time.Second
	retryDelay          = 100 * time.Millisecond
)

type testHelper struct {
	t      *testing.T
	ctx    context.Context
	client *azservicebus.Client
}

var serviceBusConfig = `{
    "UserConfig": {
        "Namespaces": [{
            "Name": "sbemulatorns",
            "Queues": [{
                "Name": "queue1",
                "Properties": {
                    "DeadLetteringOnMessageExpiration": false,
                    "DefaultMessageTimeToLive": "PT1H",
                    "DuplicateDetectionHistoryTimeWindow": "PT20S",
                    "LockDuration": "PT1M",
                    "MaxDeliveryCount": 10,
                    "RequiresDuplicateDetection": false,
                    "RequiresSession": false
                }
            }]
        }],
        "Logging": {
            "Type": "File"
        }
    }
}`

// TestFixture encapsulates the test environment setup
type TestFixture struct {
	container testcontainers.Container
	client    *azservicebus.Client
}

// Setup initializes the test fixture
func setupTestFixture(t *testing.T, ctx context.Context) *TestFixture {
	t.Helper()

	container := setupServiceBusContainer(t, ctx)
	client := createServiceBusClient(t, ctx, container)

	return &TestFixture{
		container: container,
		client:    client,
	}
}

// Cleanup tears down the test fixture
func (tf *TestFixture) cleanup(t *testing.T) {
	t.Helper()
	if err := testcontainers.TerminateContainer(tf.container); err != nil {
		t.Logf("failed to terminate container: %v", err)
	}
}

func TestServiceBusEmulator(t *testing.T) {
	ctx := context.Background()
	fixture := setupTestFixture(t, ctx)
	defer fixture.cleanup(t)

	helper := &testHelper{
		t:      t,
		ctx:    ctx,
		client: fixture.client,
	}

	t.Run("SendMessage", helper.testSendMessage)
	t.Run("ReceiveMessage", helper.testReceiveMessage)
	t.Run("DeadLetterMessage", helper.testDeadLetterMessage)
}

func (h *testHelper) testSendMessage(t *testing.T) {
	h.sendMessage(testMessage)
}

func (h *testHelper) testReceiveMessage(t *testing.T) {
	// Send a message to receive
	h.sendMessage(testMessage)

	// Receive it
	receivedMessage := h.receiveMessage()

	if receivedMessage != testMessage {
		t.Errorf("expected message %q, got %q", testMessage, receivedMessage)
	}
}

func (h *testHelper) testDeadLetterMessage(t *testing.T) {
	uniqueMessage := deadLetterMessage + " - " + t.Name()

	// Send a message that we'll abandon repeatedly
	h.sendMessage(uniqueMessage)

	// Exceed max delivery count
	h.exceedMaxDeliveryCount(uniqueMessage)

	// Verify the message is in the dead letter queue
	receivedMessage := h.receiveDeadLetterMessage()

	if receivedMessage != uniqueMessage {
		t.Errorf("expected dead letter message %q, got %q", uniqueMessage, receivedMessage)
	}
}

func setupServiceBusContainer(t *testing.T, ctx context.Context) testcontainers.Container {
	t.Helper()

	container, err := servicebus.Run(
		ctx,
		"mcr.microsoft.com/azure-messaging/servicebus-emulator:1.1.2",
		servicebus.WithAcceptEULA(),
		servicebus.WithConfig(strings.NewReader(serviceBusConfig)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	return container
}

type connectionStringGetter interface {
	ConnectionString(context.Context) (string, error)
}

func createServiceBusClient(t *testing.T, ctx context.Context, container testcontainers.Container) *azservicebus.Client {
	t.Helper()

	csg, ok := container.(connectionStringGetter)
	if !ok {
		t.Fatal("container does not implement ConnectionString method")
	}

	connectionString, err := csg.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client
}

func (h *testHelper) sendMessage(messageBody string) {
	h.t.Helper()

	sender, err := h.client.NewSender(queueName, nil)
	if err != nil {
		h.t.Fatalf("failed to create sender: %v", err)
	}
	defer sender.Close(h.ctx)

	message := &azservicebus.Message{Body: []byte(messageBody)}

	// Retry sending because queue might not be ready immediately
	for attempt := 0; attempt < maxRetries; attempt++ {
		if err = sender.SendMessage(h.ctx, message, nil); err == nil {
			return
		}
		if attempt < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	h.t.Fatalf("failed to send message after %d attempts: %v", maxRetries, err)
}

func (h *testHelper) receiveMessage() string {
	h.t.Helper()

	receiver, err := h.client.NewReceiverForQueue(queueName, nil)
	if err != nil {
		h.t.Fatalf("failed to create receiver: %v", err)
	}
	defer receiver.Close(h.ctx)

	messages, err := receiver.ReceiveMessages(h.ctx, 1, nil)
	if err != nil {
		h.t.Fatalf("failed to receive messages: %v", err)
	}
	if len(messages) == 0 {
		h.t.Fatal("no messages received")
	}

	message := messages[0]
	if err = receiver.CompleteMessage(h.ctx, message, nil); err != nil {
		h.t.Fatalf("failed to complete message: %v", err)
	}

	return string(message.Body)
}

func (h *testHelper) exceedMaxDeliveryCount(targetMessage string) {
	h.t.Helper()

	receiver, err := h.client.NewReceiverForQueue(queueName, nil)
	if err != nil {
		h.t.Fatalf("failed to create receiver: %v", err)
	}
	defer receiver.Close(h.ctx)

	for attempt := 0; attempt < maxDeliveryAttempts; attempt++ {
		receiveCtx, cancel := context.WithTimeout(h.ctx, receiveTimeout)
		messages, err := receiver.ReceiveMessages(receiveCtx, 1, nil)
		cancel()

		if err != nil {
			h.t.Logf("attempt %d: failed to receive messages: %v", attempt+1, err)
			continue
		}
		if len(messages) == 0 {
			h.t.Logf("attempt %d: no messages received, message may have been dead lettered", attempt+1)
			break
		}

		message := messages[0]
		messageBody := string(message.Body)

		// Only process our target message
		if messageBody != targetMessage {
			h.t.Logf("attempt %d: skipping unrelated message: %q", attempt+1, messageBody)
			if err = receiver.CompleteMessage(h.ctx, message, nil); err != nil {
				h.t.Logf("failed to complete unrelated message: %v", err)
			}
			continue
		}

		// Abandon target message to increase delivery count
		if err = receiver.AbandonMessage(h.ctx, message, nil); err != nil {
			h.t.Logf("attempt %d: failed to abandon message: %v", attempt+1, err)
		} else {
			h.t.Logf("attempt %d: abandoned message, delivery count: %d", attempt+1, message.DeliveryCount)
		}

		time.Sleep(retryDelay)
	}
}

func (h *testHelper) receiveDeadLetterMessage() string {
	h.t.Helper()

	receiver, err := h.client.NewReceiverForQueue(queueName, &azservicebus.ReceiverOptions{
		SubQueue: azservicebus.SubQueueDeadLetter,
	})
	if err != nil {
		h.t.Fatalf("failed to create dead letter receiver: %v", err)
	}
	defer receiver.Close(h.ctx)

	receiveCtx, cancel := context.WithTimeout(h.ctx, deadLetterTimeout)
	defer cancel()

	messages, err := receiver.ReceiveMessages(receiveCtx, 1, nil)
	if err != nil {
		h.t.Fatalf("failed to receive dead letter messages: %v", err)
	}
	if len(messages) == 0 {
		h.t.Fatal("no dead letter messages received")
	}

	message := messages[0]
	if err = receiver.CompleteMessage(h.ctx, message, nil); err != nil {
		h.t.Fatalf("failed to complete dead letter message: %v", err)
	}

	return string(message.Body)
}
