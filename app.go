package main

import (
	"context"
	"ctools/backend/network"
)

// App struct
type App struct {
	ctx            context.Context
	networkService *network.NetworkService
}

// NewApp creates a new App application struct
func NewApp(netService *network.NetworkService) *App {
	return &App{
		networkService: netService,
	}
}

// startup is called when the app starts. The context is saved
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.networkService.SetContext(ctx)
}
