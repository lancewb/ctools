package main

import (
	"context"
	"testing"

	"ctools/backend/crypto"
	"ctools/backend/network"
	"ctools/backend/other"
)

func TestAppStartupSetsServiceContexts(t *testing.T) {
	networkService := network.NewNetworkService()
	cryptoService := crypto.NewCryptoService()
	otherService := other.NewOtherService(cryptoService)
	app := NewApp(networkService, cryptoService, otherService)

	ctx := context.Background()
	app.startup(ctx)

	if app.ctx != ctx {
		t.Fatalf("expected app context to be set")
	}
	if app.networkService != networkService || app.cryptoService != cryptoService || app.otherService != otherService {
		t.Fatalf("expected app services to be retained")
	}
}
