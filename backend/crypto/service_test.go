package crypto

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExtractPEMOrDERAcceptsDERInput(t *testing.T) {
	rawHex := "a1b2c3d4"
	block, der, err := extractPEMOrDER(rawHex, "hex")
	if err != nil {
		t.Fatalf("expected DER data to be accepted: %v", err)
	}
	if block != nil {
		t.Fatalf("expected no PEM block for DER input")
	}
	if hex.EncodeToString(der) != strings.ToLower(rawHex) {
		t.Fatalf("unexpected DER result: %x", der)
	}
}

func TestExtractPEMOrDERRaisesOnExtraPEMData(t *testing.T) {
	pemData := "-----BEGIN TEST-----\nAQID\n-----END TEST-----\nTRAILING"
	if _, _, err := extractPEMOrDER(pemData, "pem"); err == nil {
		t.Fatalf("expected error for trailing data after PEM block")
	}
}

func TestCryptoServiceGenerateKeyPairPersistsToCustomDir(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CTOOLS_CONFIG_DIR", tempDir)

	svc := NewCryptoService()
	_, err := svc.GenerateKeyPair(KeyGenRequest{
		Algorithm: "rsa",
		KeySize:   512,
		Save:      true,
		Name:      "unit-test",
	})
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	keys := svc.ListStoredKeys()
	if len(keys) != 1 {
		t.Fatalf("expected 1 stored key, got %d", len(keys))
	}
	if keys[0].PrivatePEM == "" || keys[0].PublicPEM == "" {
		t.Fatal("stored key should contain PEM material")
	}

	storePath := filepath.Join(tempDir, keyStoreFile)
	info, err := os.Stat(storePath)
	if err != nil {
		t.Fatalf("expected key store file: %v", err)
	}
	if runtime.GOOS != "windows" {
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Fatalf("expected key store to have 0600 permissions, got %v", perm)
		}
	}
}
