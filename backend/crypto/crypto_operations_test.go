package crypto

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestRunHashVectors(t *testing.T) {
	service := NewCryptoService()

	hashResult, err := service.RunHash(HashRequest{
		Algorithm:    "sha256",
		Input:        "abc",
		OutputFormat: "hex",
	})
	if err != nil {
		t.Fatalf("RunHash failed: %v", err)
	}
	if hashResult.Output != "BA7816BF8F01CFEA414140DE5DAE2223B00361A396177A9CB410FF61F20015AD" {
		t.Fatalf("unexpected SHA-256 digest: %s", hashResult.Output)
	}

	hmacResult, err := service.RunHash(HashRequest{
		Algorithm:    "sha256",
		Mode:         "hmac",
		Input:        "data",
		Key:          "secret",
		OutputFormat: "hex",
	})
	if err != nil {
		t.Fatalf("RunHash HMAC failed: %v", err)
	}
	if hmacResult.Output != "1B2C16B75BD2A870C114153CCDA5BCFCA63314BC722FA160D690DE133CCBB9DB" {
		t.Fatalf("unexpected HMAC digest: %s", hmacResult.Output)
	}
}

func TestRunSymmetricRoundTrips(t *testing.T) {
	service := NewCryptoService()
	cases := []SymmetricRequest{
		{Algorithm: "aes", Mode: "cbc", Padding: "pkcs7", Key: "0123456789abcdef", IV: "abcdef9876543210"},
		{Algorithm: "aes", Mode: "ecb", Padding: "pkcs7", Key: "0123456789abcdef"},
		{Algorithm: "aes", Mode: "ctr", Key: "0123456789abcdef", IV: "abcdef9876543210"},
		{Algorithm: "aes", Mode: "gcm", Key: "0123456789abcdef", Nonce: "123456789012", Additional: "aad"},
		{Algorithm: "sm4", Mode: "cbc", Padding: "pkcs7", Key: "0123456789abcdef", IV: "abcdef9876543210"},
		{Algorithm: "3des", Mode: "cbc", Padding: "pkcs7", Key: "0123456789abcdefghijklmn", IV: "12345678"},
		{Algorithm: "chacha20-poly1305", Key: "0123456789abcdefghijklmnopqrstuv", Nonce: "123456789012", Additional: "aad"},
	}

	for _, tc := range cases {
		t.Run(tc.Algorithm+"-"+tc.Mode, func(t *testing.T) {
			tc.Operation = "encrypt"
			tc.Input = "secret message"
			tc.OutputFormat = "base64"
			encrypted, err := service.RunSymmetric(tc)
			if err != nil {
				t.Fatalf("encrypt failed: %v", err)
			}
			if _, err := base64.StdEncoding.DecodeString(encrypted.Output); err != nil {
				t.Fatalf("expected base64 ciphertext: %v", err)
			}

			tc.Operation = "decrypt"
			tc.Input = encrypted.Output
			tc.InputFormat = "base64"
			tc.OutputFormat = "base64"
			decrypted, err := service.RunSymmetric(tc)
			if err != nil {
				t.Fatalf("decrypt failed: %v", err)
			}
			plain, err := base64.StdEncoding.DecodeString(decrypted.Output)
			if err != nil {
				t.Fatalf("decode plaintext: %v", err)
			}
			if string(plain) != "secret message" {
				t.Fatalf("unexpected plaintext: %q", plain)
			}
		})
	}
}

func TestCMACAndDiversification(t *testing.T) {
	service := NewCryptoService()
	cmac, err := service.RunSymmetric(SymmetricRequest{
		Algorithm:    "aes",
		Operation:    "cmac",
		Key:          "2b7e151628aed2a6abf7158809cf4f3c",
		KeyFormat:    "hex",
		Input:        "6bc1bee22e409f96e93d7e117393172a",
		InputFormat:  "hex",
		OutputFormat: "hex",
	})
	if err != nil {
		t.Fatalf("CMAC failed: %v", err)
	}
	if cmac.Output != "070A16B46B4D4144F79BDD9DD04A287C" {
		t.Fatalf("unexpected CMAC: %s", cmac.Output)
	}

	diversified, err := service.RunSymmetric(SymmetricRequest{
		Algorithm:    "aes",
		Operation:    "diversify16",
		Key:          "000102030405060708090a0b0c0d0e0f",
		KeyFormat:    "hex",
		Input:        "00112233445566778899aabbccddeeff",
		InputFormat:  "hex",
		OutputFormat: "hex",
	})
	if err != nil {
		t.Fatalf("diversification failed: %v", err)
	}
	if len(diversified.Output) != 32 {
		t.Fatalf("expected 16-byte diversification output, got %q", diversified.Output)
	}
}

func TestKeyGenerationParsingAndAsymmetricOperations(t *testing.T) {
	t.Setenv("CTOOLS_CONFIG_DIR", t.TempDir())
	service := NewCryptoService()

	rsaKey, err := service.GenerateKeyPair(KeyGenRequest{Algorithm: "rsa", KeySize: 1024, Save: true, Name: "rsa"})
	if err != nil {
		t.Fatalf("Generate RSA failed: %v", err)
	}
	if _, err := service.ParseKey(KeyParseRequest{Algorithm: "rsa", Format: "pem", Data: rsaKey.PrivatePEM}); err != nil {
		t.Fatalf("Parse RSA failed: %v", err)
	}
	rsaCipher, err := service.RunAsymmetric(AsymmetricRequest{
		Algorithm:    "rsa",
		Operation:    "encrypt",
		Padding:      "pkcs1",
		KeyData:      rsaKey.PublicPEM,
		KeyFormat:    "pem",
		Payload:      "hello",
		OutputFormat: "base64",
	})
	if err != nil {
		t.Fatalf("RSA encrypt failed: %v", err)
	}
	rsaPlain, err := service.RunAsymmetric(AsymmetricRequest{
		Algorithm:     "rsa",
		Operation:     "decrypt",
		Padding:       "pkcs1",
		KeyData:       rsaKey.PrivatePEM,
		KeyFormat:     "pem",
		Payload:       rsaCipher.Output,
		PayloadFormat: "base64",
		OutputFormat:  "base64",
	})
	if err != nil {
		t.Fatalf("RSA decrypt failed: %v", err)
	}
	assertBase64Text(t, rsaPlain.Output, "hello")

	rsaSig, err := service.RunAsymmetric(AsymmetricRequest{
		Algorithm:    "rsa",
		Operation:    "sign",
		KeyData:      rsaKey.PrivatePEM,
		KeyFormat:    "pem",
		Payload:      "signed",
		OutputFormat: "base64",
	})
	if err != nil {
		t.Fatalf("RSA sign failed: %v", err)
	}
	rsaVerify, err := service.RunAsymmetric(AsymmetricRequest{
		Algorithm:    "rsa",
		Operation:    "verify",
		KeyData:      rsaKey.PublicPEM,
		KeyFormat:    "pem",
		Payload:      "signed",
		Signature:    rsaSig.Output,
		SignatureFmt: "base64",
	})
	if err != nil {
		t.Fatalf("RSA verify failed: %v", err)
	}
	if !rsaVerify.Verified {
		t.Fatalf("expected RSA signature to verify")
	}

	eccKey, err := service.GenerateKeyPair(KeyGenRequest{Algorithm: "ecc", Curve: "p-256", Save: true, Name: "ecc"})
	if err != nil {
		t.Fatalf("Generate ECC failed: %v", err)
	}
	if _, err := service.ParseKey(KeyParseRequest{Algorithm: "ecc", Format: "pem", Data: eccKey.PrivatePEM}); err != nil {
		t.Fatalf("Parse ECC failed: %v", err)
	}
	eccSig, err := service.RunAsymmetric(AsymmetricRequest{
		Algorithm:    "ecc",
		Operation:    "sign",
		KeyData:      eccKey.PrivatePEM,
		KeyFormat:    "pem",
		Payload:      "signed",
		OutputFormat: "base64",
	})
	if err != nil {
		t.Fatalf("ECC sign failed: %v", err)
	}
	eccVerify, err := service.RunAsymmetric(AsymmetricRequest{
		Algorithm:    "ecc",
		Operation:    "verify",
		KeyData:      eccKey.PublicPEM,
		KeyFormat:    "pem",
		Payload:      "signed",
		Signature:    eccSig.Output,
		SignatureFmt: "base64",
	})
	if err != nil || !eccVerify.Verified {
		t.Fatalf("expected ECC signature to verify, result=%+v err=%v", eccVerify, err)
	}
	eccCipher, err := service.RunAsymmetric(AsymmetricRequest{
		Algorithm:    "ecc",
		Operation:    "encrypt",
		KeyData:      eccKey.PublicPEM,
		KeyFormat:    "pem",
		Payload:      "hello",
		OutputFormat: "base64",
	})
	if err != nil {
		t.Fatalf("ECC encrypt failed: %v", err)
	}
	eccPlain, err := service.RunAsymmetric(AsymmetricRequest{
		Algorithm:     "ecc",
		Operation:     "decrypt",
		KeyData:       eccKey.PrivatePEM,
		KeyFormat:     "pem",
		Payload:       eccCipher.Output,
		PayloadFormat: "base64",
		OutputFormat:  "base64",
	})
	if err != nil {
		t.Fatalf("ECC decrypt failed: %v", err)
	}
	assertBase64Text(t, eccPlain.Output, "hello")

	sm2Key, err := service.GenerateKeyPair(KeyGenRequest{Algorithm: "sm2", Save: true, Name: "sm2"})
	if err != nil {
		t.Fatalf("Generate SM2 failed: %v", err)
	}
	if _, err := service.ParseKey(KeyParseRequest{Algorithm: "sm2", Format: "pem", Data: sm2Key.PrivatePEM}); err != nil {
		t.Fatalf("Parse SM2 failed: %v", err)
	}
	sm2Sig, err := service.RunAsymmetric(AsymmetricRequest{
		Algorithm:    "sm2",
		Operation:    "sign",
		KeyData:      sm2Key.PrivatePEM,
		KeyFormat:    "pem",
		Payload:      "signed",
		OutputFormat: "base64",
	})
	if err != nil {
		t.Fatalf("SM2 sign failed: %v", err)
	}
	sm2Verify, err := service.RunAsymmetric(AsymmetricRequest{
		Algorithm:    "sm2",
		Operation:    "verify",
		KeyData:      sm2Key.PublicPEM,
		KeyFormat:    "pem",
		Payload:      "signed",
		Signature:    sm2Sig.Output,
		SignatureFmt: "base64",
	})
	if err != nil || !sm2Verify.Verified {
		t.Fatalf("expected SM2 signature to verify, result=%+v err=%v", sm2Verify, err)
	}
	sm2Cipher, err := service.RunAsymmetric(AsymmetricRequest{
		Algorithm:    "sm2",
		Operation:    "encrypt",
		KeyData:      sm2Key.PublicPEM,
		KeyFormat:    "pem",
		Payload:      "hello",
		OutputFormat: "base64",
	})
	if err != nil {
		t.Fatalf("SM2 encrypt failed: %v", err)
	}
	sm2Plain, err := service.RunAsymmetric(AsymmetricRequest{
		Algorithm:     "sm2",
		Operation:     "decrypt",
		KeyData:       sm2Key.PrivatePEM,
		KeyFormat:     "pem",
		Payload:       sm2Cipher.Output,
		PayloadFormat: "base64",
		OutputFormat:  "base64",
	})
	if err != nil {
		t.Fatalf("SM2 decrypt failed: %v", err)
	}
	assertBase64Text(t, sm2Plain.Output, "hello")

	sm9Key, err := service.GenerateKeyPair(KeyGenRequest{Algorithm: "sm9", Variant: "sign-master", Save: true, Name: "sm9"})
	if err != nil {
		t.Fatalf("Generate SM9 failed: %v", err)
	}
	if !strings.Contains(sm9Key.PrivatePEM, "SM9 SIGN MASTER PRIVATE KEY") {
		t.Fatalf("unexpected SM9 PEM: %s", sm9Key.PrivatePEM)
	}
}

func TestCertificatesAndDERParsing(t *testing.T) {
	t.Setenv("CTOOLS_CONFIG_DIR", t.TempDir())
	service := NewCryptoService()

	issued, err := service.IssueCertificate(CertIssueRequest{
		CommonName: "unit.example",
		Algorithm:  "rsa",
		KeySize:    2048,
		ValidDays:  1,
		Usage:      "server",
	})
	if err != nil {
		t.Fatalf("IssueCertificate failed: %v", err)
	}
	if len(issued.Certificates) != 1 || len(issued.Keys) != 1 {
		t.Fatalf("unexpected issue result: %+v", issued)
	}

	parsed, err := service.ParseCertificate(CertParseRequest{PEM: issued.Certificates[0].CertPEM})
	if err != nil {
		t.Fatalf("ParseCertificate failed: %v", err)
	}
	if parsed.Subject["CN"] != "unit.example" {
		t.Fatalf("unexpected certificate subject: %+v", parsed.Subject)
	}

	exported, err := service.ExportCertificate(issued.Certificates[0].ID)
	if err != nil {
		t.Fatalf("ExportCertificate failed: %v", err)
	}
	if exported.Key == nil || exported.Key.PrivatePEM == "" {
		t.Fatalf("expected exported certificate to include key")
	}

	if _, err := service.ParseDER(DerParseRequest{HexString: "3003020105"}); err != nil {
		t.Fatalf("ParseDER failed: %v", err)
	}

	remainingCerts := service.DeleteCertificate(issued.Certificates[0].ID)
	for _, cert := range remainingCerts {
		if cert.ID == issued.Certificates[0].ID {
			t.Fatalf("certificate was not deleted")
		}
	}
	remainingKeys := service.DeleteStoredKey(issued.Keys[0].ID)
	for _, key := range remainingKeys {
		if key.ID == issued.Keys[0].ID {
			t.Fatalf("key was not deleted")
		}
	}

	sm2Issued, err := service.IssueCertificate(CertIssueRequest{
		CommonName: "sm2.example",
		Algorithm:  "sm2",
		ValidDays:  1,
		Usage:      "server",
	})
	if err != nil {
		t.Fatalf("IssueCertificate SM2 failed: %v", err)
	}
	if len(sm2Issued.Certificates) != 2 || len(sm2Issued.Keys) != 2 {
		t.Fatalf("unexpected SM2 issue result: %+v", sm2Issued)
	}
	if _, err := service.ParseCertificate(CertParseRequest{PEM: sm2Issued.Certificates[0].CertPEM}); err != nil {
		t.Fatalf("ParseCertificate SM2 failed: %v", err)
	}
}

func assertBase64Text(t *testing.T, encoded, expected string) {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode base64: %v", err)
	}
	if string(data) != expected {
		t.Fatalf("expected %q, got %q", expected, data)
	}
}
