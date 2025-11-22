package network

import "context"

type NetworkService struct {
	ctx context.Context
}

// NewNetworkService 构造函数
func NewNetworkService() *NetworkService {
	return &NetworkService{}
}

// SetContext 供 App 在启动时注入 Context
func (n *NetworkService) SetContext(ctx context.Context) {
	n.ctx = ctx
}
