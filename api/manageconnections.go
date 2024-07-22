package api

import "main/types"

func AddNewEndpoint(endpoint string) types.EndpointConfig {
	conn := types.EndpointConfig{
		Endpoint: endpoint,
		Tags:     nil,
		Interval: 30,
	}
	return conn
}
