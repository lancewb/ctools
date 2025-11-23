package main

import (
	"context"
	"ctools/backend/crypto"
	"ctools/backend/network"
	"ctools/backend/other"
)

// App struct
type App struct {
	ctx            context.Context
	networkService *network.NetworkService
	cryptoService  *crypto.CryptoService
	otherService   *other.OtherService
}

// NewApp creates a new App application struct
func NewApp(netService *network.NetworkService, cryptoService *crypto.CryptoService, otherService *other.OtherService) *App {
	return &App{
		networkService: netService,
		cryptoService:  cryptoService,
		otherService:   otherService,
	}
}

// startup is called when the app starts. The context is saved
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.networkService.SetContext(ctx)
	a.cryptoService.SetContext(ctx)
	a.otherService.SetContext(ctx)
}
