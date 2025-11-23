package network

import "context"

// NetworkService handles network-related operations such as pinging, HTTP requests, and SSH connections.
type NetworkService struct {
	ctx context.Context
}

// NewNetworkService initializes a new NetworkService instance.
func NewNetworkService() *NetworkService {
	return &NetworkService{}
}

// SetContext sets the application context.
func (n *NetworkService) SetContext(ctx context.Context) {
	n.ctx = ctx
}
