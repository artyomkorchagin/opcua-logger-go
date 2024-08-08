package api

import (
	"context"
	"fmt"
	"log"
	"main/types"

	"github.com/gopcua/opcua"
)

func FindServers(ctx context.Context, endpoint string) ([]types.EndpointConfig, error) {
	log.Println("Finding servers")
	servers, err := opcua.FindServers(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	endpoints := []types.EndpointConfig{}
	for _, server := range servers {
		endpoints = append(endpoints, types.EndpointConfig{
			Endpoint: server.DiscoveryURLs[0],
			Tags:     nil,
		})
	}
	return endpoints, nil
}

func FillEndpointConfig(endpointConfig *types.EndpointConfig, nodeList []NodeDef) error {
	for i, node := range nodeList {
		tag := types.Tag{
			ID:          fmt.Sprint(i + 1),
			Enabled:     1,
			Name:        node.BrowseName,
			Description: node.Description,
			Address:     node.NodeID.String(),
		}
		endpointConfig.Tags = append(endpointConfig.Tags, tag)
	}

	return nil
}
