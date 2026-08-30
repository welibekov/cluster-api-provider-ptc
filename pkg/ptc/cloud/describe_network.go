package cloud

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/client/operations"
)

type Network struct {
	// Unique identifier of the network.
	ID string `json:"id,omitempty"`

	// Name of the network.
	Name string `json:"name,omitempty"`

	// Current status of the network (e.g., active, inactive).
	Status string `json:"status,omitempty"`
}

func (c *Client) DescribeNetworkByName(ctx context.Context, networkName string) (*Network, error) {
	authorizer, err := c.GetAuthorizer(ctx)
	if err != nil {
		return nil, err
	}

	params := operations.NewListNetworksParams()
	resp, err := c.apiClient.Operations.ListNetworksContext(ctx, params, authorizer)
	if err != nil {
		return nil, fmt.Errorf("error calling API with auth: %w", err)
	}

	payloadBytes, err := json.Marshal(resp.GetPayload())
	if err != nil {
		return nil, fmt.Errorf("failed to get network list: %w", err)
	}

	var networks []Network
	if err := json.Unmarshal(payloadBytes, &networks); err != nil {
		return nil, fmt.Errorf("failed to get network list: %w", err)
	}

	for _, network := range networks {
		if network.Name == networkName {
			return &network, nil
		}
	}

	return nil, fmt.Errorf("network %s not found", networkName)
}
