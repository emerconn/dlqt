package servicebus

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

func GetClient(namespace string) (*azservicebus.Client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	client, err := azservicebus.NewClient(namespace, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Service Bus client for namespace '%s': %w", namespace, err)
	}
	return client, nil
}
