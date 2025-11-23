package main

import (
	"context"
	"ctools/backend/crypto"
	"ctools/backend/network"
	"ctools/backend/other"
)

// App represents the main application structure.
// It holds references to various services and the application context.
type App struct {
	ctx            context.Context
	networkService *network.NetworkService
	cryptoService  *crypto.CryptoService
	otherService   *other.OtherService
}

// NewApp initializes and returns a new App instance.
//
// netService is the service handling network operations.
// cryptoService is the service handling cryptographic operations.
// otherService is the service handling other miscellaneous operations.
//
// Returns a pointer to the App struct.
func NewApp(netService *network.NetworkService, cryptoService *crypto.CryptoService, otherService *other.OtherService) *App {
	return &App{
		networkService: netService,
		cryptoService:  cryptoService,
		otherService:   otherService,
	}
}

// startup is called when the app starts.
// It saves the application context and passes it to the registered services.
//
// ctx is the Wails application context.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.networkService.SetContext(ctx)
	a.cryptoService.SetContext(ctx)
	a.otherService.SetContext(ctx)
}
