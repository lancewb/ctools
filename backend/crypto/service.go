package crypto

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// CryptoService exposes key management, crypto primitives, and certificate utilities to the frontend.
type CryptoService struct {
	ctx context.Context
}

// NewCryptoService initializes a new CryptoService instance.
func NewCryptoService() *CryptoService {
	return &CryptoService{}
}

// SetContext sets the application context.
func (c *CryptoService) SetContext(ctx context.Context) {
	c.ctx = ctx
}

const (
	configDirOverrideEnv = "CTOOLS_CONFIG_DIR"
	keyStoreFile         = "crypto_keys.json"
	certStoreFile        = "crypto_certs.json"
	caStoreFile          = "crypto_ca.json"
)

// StoredKey represents a cryptographic key persisted in storage.
type StoredKey struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Algorithm  string            `json:"algorithm"` // RSA, ECC, SM2, SM9
	KeyType    string            `json:"keyType"`   // private, public
	Format     string            `json:"format"`    // pem, generated
	Usage      []string          `json:"usage"`     // sign, encrypt, etc.
	PrivatePEM string            `json:"privatePem,omitempty"`
	PublicPEM  string            `json:"publicPem,omitempty"`
	Extra      map[string]string `json:"extra,omitempty"` // Curve name, variant, etc.
	CreatedAt  time.Time         `json:"createdAt" ts_type:"string"`
}

// KeyParseRequest defines the input for parsing key material.
type KeyParseRequest struct {
	Name      string   `json:"name"`
	Algorithm string   `json:"algorithm"`
	Format    string   `json:"format"` // pem, hex, base64
	Data      string   `json:"data"`
	Usage     []string `json:"usage"`
	Variant   string   `json:"variant"` // master/sign/encrypt etc. (for SM9)
	Save      bool     `json:"save"`
}

// KeyParseResult contains the result of a key parsing operation.
type KeyParseResult struct {
	Stored     bool              `json:"stored"`
	Key        *StoredKey        `json:"key,omitempty"`
	PrivatePEM string            `json:"privatePem,omitempty"`
	PublicPEM  string            `json:"publicPem,omitempty"`
	Summary    map[string]string `json:"summary"`
}

// KeyGenRequest defines the parameters for generating a new key pair.
type KeyGenRequest struct {
	Name           string   `json:"name"`
	Algorithm      string   `json:"algorithm"`
	KeySize        int      `json:"keySize"`
	Curve          string   `json:"curve"`
	PublicExponent int      `json:"publicExponent"`
	Usage          []string `json:"usage"`
	Variant        string   `json:"variant"` // For SM9
	Save           bool     `json:"save"`
	UID            string   `json:"uid"` // For SM2/SM9 identity
}

// AsymmetricRequest defines the parameters for asymmetric crypto operations.
type AsymmetricRequest struct {
	Algorithm       string `json:"algorithm"` // RSA, ECC, SM2, SM9
	Operation       string `json:"operation"` // encrypt, decrypt, sign, verify
	PayloadIsHash   bool   `json:"payloadIsHash"`
	KeyID           string `json:"keyId"`
	PeerKeyID       string `json:"peerKeyId"`
	KeyData         string `json:"keyData"`
	KeyFormat       string `json:"keyFormat"`
	Payload         string `json:"payload"`
	PayloadFormat   string `json:"payloadFormat"`
	Signature       string `json:"signature"`
	SignatureFmt    string `json:"signatureFormat"`
	UID             string `json:"uid"` // User ID for SM2/SM9
	Padding         string `json:"padding"`
	OAEPHash        string `json:"oaepHash"`
	MGF1Hash        string `json:"mgf1Hash"`
	OutputFormat    string `json:"outputFormat"`
	KDF             string `json:"kdf"`
	SymmetricCipher string `json:"symmetricCipher"` // For ECIES
	MacAlgorithm    string `json:"macAlgorithm"`    // For ECIES
	EccMode         string `json:"eccMode"`         // C1C2C3 vs C1C3C2 etc.
}

// SymmetricRequest defines the parameters for symmetric crypto operations.
type SymmetricRequest struct {
	Algorithm        string `json:"algorithm"` // AES, SM4, DES, ChaCha20
	Mode             string `json:"mode"`      // CBC, ECB, GCM, CTR
	Padding          string `json:"padding"`   // PKCS7, Zero, None
	Operation        string `json:"operation"` // encrypt, decrypt
	Key              string `json:"key"`
	KeyFormat        string `json:"keyFormat"`
	IV               string `json:"iv"`
	IVFormat         string `json:"ivFormat"`
	Nonce            string `json:"nonce"`
	NonceFormat      string `json:"nonceFormat"`
	Input            string `json:"input"`
	InputFormat      string `json:"inputFormat"`
	Additional       string `json:"additionalData"` // AAD for GCM/Poly1305
	AdditionalFormat string `json:"additionalDataFormat"`
	OutputFormat     string `json:"outputFormat"`
}

// HashRequest defines the parameters for hashing operations.
type HashRequest struct {
	Algorithm    string `json:"algorithm"` // SHA256, SM3, MD5, etc.
	Mode         string `json:"mode"`      // hash, hmac
	Input        string `json:"input"`
	InputFormat  string `json:"inputFormat"`
	Key          string `json:"key"` // HMAC key
	KeyFormat    string `json:"keyFormat"`
	OutputFormat string `json:"outputFormat"`
}

// OperationResult contains the output of a cryptographic operation.
type OperationResult struct {
	Output   string            `json:"output,omitempty"`
	Verified bool              `json:"verified"`
	Details  map[string]string `json:"details,omitempty"`
}

// CertIssueRequest defines the parameters for issuing a certificate.
type CertIssueRequest struct {
	CommonName string `json:"commonName"`
	Algorithm  string `json:"algorithm"` // RSA, SM2
	KeySize    int    `json:"keySize"`
	ValidDays  int    `json:"validDays"`
	Usage      string `json:"usage"` // server, client
	Save       bool   `json:"save"`
}

// CertRecord represents a stored certificate.
type CertRecord struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Algorithm string            `json:"algorithm"`
	Usage     string            `json:"usage"`
	CertPEM   string            `json:"certPem"`
	KeyID     string            `json:"keyId,omitempty"` // ID of the associated private key
	Serial    string            `json:"serial"`
	NotBefore string            `json:"notBefore"`
	NotAfter  string            `json:"notAfter"`
	Subject   map[string]string `json:"subject"`
	Issuer    map[string]string `json:"issuer"`
	CreatedAt time.Time         `json:"createdAt" ts_type:"string"`
}

// CertIssueResult contains the generated certificate and keys.
type CertIssueResult struct {
	RootCA       *CertRecord  `json:"rootCa,omitempty"`
	Certificates []CertRecord `json:"certificates"`
	Keys         []*StoredKey `json:"keys"`
}

// CertExport bundles a certificate with its key for export.
type CertExport struct {
	Cert CertRecord `json:"cert"`
	Key  *StoredKey `json:"key,omitempty"`
}

// CertParseRequest defines the input for parsing a certificate.
type CertParseRequest struct {
	PEM string `json:"pem"`
}

// CertParseResult contains the parsed details of a certificate.
type CertParseResult struct {
	Subject            map[string]string `json:"subject"`
	Issuer             map[string]string `json:"issuer"`
	Serial             string            `json:"serial"`
	NotBefore          string            `json:"notBefore"`
	NotAfter           string            `json:"notAfter"`
	PublicKeyAlgorithm string            `json:"publicKeyAlgorithm"`
	SignatureAlgorithm string            `json:"signatureAlgorithm"`
	DNSNames           []string          `json:"dnsNames"`
	IPAddresses        []string          `json:"ipAddresses"`
	SANs               []string          `json:"sans"`
	KeyUsage           []string          `json:"keyUsage"`
	ExtKeyUsage        []string          `json:"extKeyUsage"`
	RawHex             string            `json:"rawHex"`
}

// DerParseRequest defines the input for parsing ASN.1 DER data.
type DerParseRequest struct {
	Name      string `json:"name"`
	HexString string `json:"hexString"`
	Base64    string `json:"base64"`
}

// DerNode represents a node in the ASN.1 DER structure.
type DerNode struct {
	Tag         int       `json:"tag"`
	Class       string    `json:"class"`
	Label       string    `json:"label,omitempty"`
	Constructed bool      `json:"constructed"`
	Length      int       `json:"length"`
	Value       string    `json:"value,omitempty"`
	Hex         string    `json:"hex"`
	Children    []DerNode `json:"children,omitempty"`
}

// DerParseResult contains the parsed DER tree.
type DerParseResult struct {
	Nodes []DerNode `json:"nodes"`
}

// ensureDataDir creates and returns the application data directory.
func (c *CryptoService) ensureDataDir() string {
	if override := strings.TrimSpace(os.Getenv(configDirOverrideEnv)); override != "" {
		_ = os.MkdirAll(override, 0o755)
		return override
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "."
	}
	root := filepath.Join(configDir, "WailsToolbox")
	os.MkdirAll(root, 0755)
	return root
}

// keyStorePath returns the file path for the key store.
func (c *CryptoService) keyStorePath() string {
	return filepath.Join(c.ensureDataDir(), keyStoreFile)
}

// certStorePath returns the file path for the certificate store.
func (c *CryptoService) certStorePath() string {
	return filepath.Join(c.ensureDataDir(), certStoreFile)
}

// caStorePath returns the file path for the CA store.
func (c *CryptoService) caStorePath() string {
	return filepath.Join(c.ensureDataDir(), caStoreFile)
}

// readKeys reads the list of stored keys from the file system.
func (c *CryptoService) readKeys() []StoredKey {
	path := c.keyStorePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return []StoredKey{}
	}
	var keys []StoredKey
	_ = json.Unmarshal(data, &keys)
	return keys
}

// writeKeys writes the list of keys to the file system.
func (c *CryptoService) writeKeys(keys []StoredKey) {
	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		log.Printf("crypto: unable to marshal key store: %v", err)
		return
	}
	if err := os.WriteFile(c.keyStorePath(), data, 0600); err != nil {
		log.Printf("crypto: unable to persist key store: %v", err)
	}
}

// readCerts reads the list of stored certificates from the file system.
func (c *CryptoService) readCerts() []CertRecord {
	path := c.certStorePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return []CertRecord{}
	}
	var certs []CertRecord
	_ = json.Unmarshal(data, &certs)
	return certs
}

// writeCerts writes the list of certificates to the file system.
func (c *CryptoService) writeCerts(certs []CertRecord) {
	data, err := json.MarshalIndent(certs, "", "  ")
	if err != nil {
		log.Printf("crypto: unable to marshal certificate store: %v", err)
		return
	}
	if err := os.WriteFile(c.certStorePath(), data, 0644); err != nil {
		log.Printf("crypto: unable to persist certificate store: %v", err)
	}
}

// decodeData decodes a payload based on the specified format.
func decodeData(format, payload string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "pem":
		block, _ := pem.Decode([]byte(payload))
		if block == nil {
			return nil, errors.New("unable to decode PEM block")
		}
		return block.Bytes, nil
	case "base64", "b64":
		return base64Decode(payload)
	case "hex":
		return hex.DecodeString(strings.ReplaceAll(strings.TrimSpace(payload), " ", ""))
	case "raw", "bytes":
		return []byte(payload), nil
	default:
		return []byte(payload), nil
	}
}

// base64Decode decodes a Base64 string, handling newlines and carriage returns.
func base64Decode(data string) ([]byte, error) {
	normalized := strings.ReplaceAll(data, "\n", "")
	normalized = strings.ReplaceAll(normalized, "\r", "")
	normalized = strings.TrimSpace(normalized)
	return base64.StdEncoding.DecodeString(normalized)
}

// encodeBase64 encodes bytes to a Base64 string.
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// findKey retrieves a stored key by its ID.
func (c *CryptoService) findKey(id string) (*StoredKey, error) {
	if id == "" {
		return nil, errors.New("missing key id")
	}
	for _, k := range c.readKeys() {
		if k.ID == id {
			return &k, nil
		}
	}
	return nil, errors.New("key not found")
}

// saveKey persists a key to storage.
func (c *CryptoService) saveKey(key StoredKey) StoredKey {
	keys := c.readKeys()
	found := false
	for i, existing := range keys {
		if existing.ID == key.ID {
			keys[i] = key
			found = true
			break
		}
	}
	if !found {
		keys = append(keys, key)
	}
	c.writeKeys(keys)
	return key
}

// ListStoredKeys returns all keys stored in the local file system.
//
// Returns a slice of StoredKey sorted by creation date (newest first).
func (c *CryptoService) ListStoredKeys() []StoredKey {
	keys := c.readKeys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].CreatedAt.After(keys[j].CreatedAt)
	})
	return keys
}

// DeleteStoredKey removes a key from storage by its ID.
//
// id: The unique identifier of the key to delete.
// Returns the updated list of stored keys.
func (c *CryptoService) DeleteStoredKey(id string) []StoredKey {
	keys := c.readKeys()
	result := make([]StoredKey, 0, len(keys))
	for _, k := range keys {
		if k.ID == id {
			continue
		}
		result = append(result, k)
	}
	c.writeKeys(result)
	return result
}

// ExportStoredKey retrieves a key from storage for export.
//
// id: The unique identifier of the key.
// Returns the StoredKey or an error if not found.
func (c *CryptoService) ExportStoredKey(id string) (StoredKey, error) {
	key, err := c.findKey(id)
	if err != nil {
		return StoredKey{}, err
	}
	return *key, nil
}

// ExportCertificate retrieves a certificate and its associated key (if available) from storage.
//
// id: The unique identifier of the certificate.
// Returns a CertExport struct containing the certificate and optional key.
func (c *CryptoService) ExportCertificate(id string) (CertExport, error) {
	var record *CertRecord
	for _, cert := range c.readCerts() {
		if cert.ID == id {
			tmp := cert
			record = &tmp
			break
		}
	}
	if record == nil {
		return CertExport{}, errors.New("certificate not found")
	}
	var key *StoredKey
	if record.KeyID != "" {
		if stored, err := c.findKey(record.KeyID); err == nil {
			key = stored
		}
	}
	return CertExport{
		Cert: *record,
		Key:  key,
	}, nil
}

// normalizeOutputFormat standardizes the output format string.
// Defaults to "hex" if not "base64".
func normalizeOutputFormat(format string) string {
	switch strings.ToLower(format) {
	case "base64":
		return "base64"
	default:
		return "hex"
	}
}

// encodeOutputBytes encodes byte data into the specified format (hex or base64).
func encodeOutputBytes(data []byte, format string) string {
	switch normalizeOutputFormat(format) {
	case "base64":
		return encodeBase64(data)
	default:
		return strings.ToUpper(hex.EncodeToString(data))
	}
}
