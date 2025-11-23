package crypto

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
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

func NewCryptoService() *CryptoService {
	return &CryptoService{}
}

func (c *CryptoService) SetContext(ctx context.Context) {
	c.ctx = ctx
}

const (
	keyStoreFile  = "crypto_keys.json"
	certStoreFile = "crypto_certs.json"
	caStoreFile   = "crypto_ca.json"
)

type StoredKey struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Algorithm  string            `json:"algorithm"`
	KeyType    string            `json:"keyType"`
	Format     string            `json:"format"`
	Usage      []string          `json:"usage"`
	PrivatePEM string            `json:"privatePem,omitempty"`
	PublicPEM  string            `json:"publicPem,omitempty"`
	Extra      map[string]string `json:"extra,omitempty"`
	CreatedAt  time.Time         `json:"createdAt" ts_type:"string"`
}

type KeyParseRequest struct {
	Name      string   `json:"name"`
	Algorithm string   `json:"algorithm"`
	Format    string   `json:"format"`
	Data      string   `json:"data"`
	Usage     []string `json:"usage"`
	Variant   string   `json:"variant"` // master/sign/encrypt etc.
	Save      bool     `json:"save"`
}

type KeyParseResult struct {
	Stored     bool              `json:"stored"`
	Key        *StoredKey        `json:"key,omitempty"`
	PrivatePEM string            `json:"privatePem,omitempty"`
	PublicPEM  string            `json:"publicPem,omitempty"`
	Summary    map[string]string `json:"summary"`
}

type KeyGenRequest struct {
	Name      string   `json:"name"`
	Algorithm string   `json:"algorithm"`
	KeySize   int      `json:"keySize"`
	Curve     string   `json:"curve"`
	Usage     []string `json:"usage"`
	Variant   string   `json:"variant"`
	Save      bool     `json:"save"`
	UID       string   `json:"uid"`
}

type AsymmetricRequest struct {
	Algorithm       string `json:"algorithm"`
	Operation       string `json:"operation"`
	KeyID           string `json:"keyId"`
	PeerKeyID       string `json:"peerKeyId"`
	KeyData         string `json:"keyData"`
	KeyFormat       string `json:"keyFormat"`
	Payload         string `json:"payload"`
	PayloadFormat   string `json:"payloadFormat"`
	Signature       string `json:"signature"`
	SignatureFmt    string `json:"signatureFormat"`
	UID             string `json:"uid"`
	Padding         string `json:"padding"`
	OAEPHash        string `json:"oaepHash"`
	MGF1Hash        string `json:"mgf1Hash"`
	OutputFormat    string `json:"outputFormat"`
	KDF             string `json:"kdf"`
	SymmetricCipher string `json:"symmetricCipher"`
	MacAlgorithm    string `json:"macAlgorithm"`
	EccMode         string `json:"eccMode"`
}

type SymmetricRequest struct {
	Algorithm        string `json:"algorithm"`
	Mode             string `json:"mode"`
	Padding          string `json:"padding"`
	Operation        string `json:"operation"`
	Key              string `json:"key"`
	KeyFormat        string `json:"keyFormat"`
	IV               string `json:"iv"`
	IVFormat         string `json:"ivFormat"`
	Nonce            string `json:"nonce"`
	NonceFormat      string `json:"nonceFormat"`
	Input            string `json:"input"`
	InputFormat      string `json:"inputFormat"`
	Additional       string `json:"additionalData"`
	AdditionalFormat string `json:"additionalDataFormat"`
	OutputFormat     string `json:"outputFormat"`
}

type HashRequest struct {
	Algorithm    string `json:"algorithm"`
	Mode         string `json:"mode"` // hash/hmac
	Input        string `json:"input"`
	InputFormat  string `json:"inputFormat"`
	Key          string `json:"key"`
	KeyFormat    string `json:"keyFormat"`
	OutputFormat string `json:"outputFormat"`
}

type OperationResult struct {
	Output   string            `json:"output,omitempty"`
	Verified bool              `json:"verified"`
	Details  map[string]string `json:"details,omitempty"`
}

type CertIssueRequest struct {
	CommonName string `json:"commonName"`
	Algorithm  string `json:"algorithm"`
	KeySize    int    `json:"keySize"`
	ValidDays  int    `json:"validDays"`
	Usage      string `json:"usage"`
	Save       bool   `json:"save"`
}

type CertRecord struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Algorithm string            `json:"algorithm"`
	Usage     string            `json:"usage"`
	CertPEM   string            `json:"certPem"`
	KeyID     string            `json:"keyId,omitempty"`
	Serial    string            `json:"serial"`
	NotBefore string            `json:"notBefore"`
	NotAfter  string            `json:"notAfter"`
	Subject   map[string]string `json:"subject"`
	Issuer    map[string]string `json:"issuer"`
	CreatedAt time.Time         `json:"createdAt" ts_type:"string"`
}

type CertIssueResult struct {
	RootCA       *CertRecord  `json:"rootCa,omitempty"`
	Certificates []CertRecord `json:"certificates"`
	Keys         []*StoredKey `json:"keys"`
}

type CertExport struct {
	Cert CertRecord `json:"cert"`
	Key  *StoredKey `json:"key,omitempty"`
}

type CertParseRequest struct {
	PEM string `json:"pem"`
}

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

type DerParseRequest struct {
	Name      string `json:"name"`
	HexString string `json:"hexString"`
	Base64    string `json:"base64"`
}

type DerNode struct {
	Tag         int       `json:"tag"`
	Class       string    `json:"class"`
	Constructed bool      `json:"constructed"`
	Length      int       `json:"length"`
	Hex         string    `json:"hex"`
	Children    []DerNode `json:"children,omitempty"`
}

type DerParseResult struct {
	Nodes []DerNode `json:"nodes"`
}

func (c *CryptoService) ensureDataDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "."
	}
	root := filepath.Join(configDir, "WailsToolbox")
	os.MkdirAll(root, 0755)
	return root
}

func (c *CryptoService) keyStorePath() string {
	return filepath.Join(c.ensureDataDir(), keyStoreFile)
}

func (c *CryptoService) certStorePath() string {
	return filepath.Join(c.ensureDataDir(), certStoreFile)
}

func (c *CryptoService) caStorePath() string {
	return filepath.Join(c.ensureDataDir(), caStoreFile)
}

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

func (c *CryptoService) writeKeys(keys []StoredKey) {
	data, _ := json.MarshalIndent(keys, "", "  ")
	_ = os.WriteFile(c.keyStorePath(), data, 0644)
}

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

func (c *CryptoService) writeCerts(certs []CertRecord) {
	data, _ := json.MarshalIndent(certs, "", "  ")
	_ = os.WriteFile(c.certStorePath(), data, 0644)
}

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

func base64Decode(data string) ([]byte, error) {
	normalized := strings.ReplaceAll(data, "\n", "")
	normalized = strings.ReplaceAll(normalized, "\r", "")
	normalized = strings.TrimSpace(normalized)
	return base64.StdEncoding.DecodeString(normalized)
}

func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

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

func (c *CryptoService) ListStoredKeys() []StoredKey {
	keys := c.readKeys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].CreatedAt.After(keys[j].CreatedAt)
	})
	return keys
}

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

func (c *CryptoService) ExportStoredKey(id string) (StoredKey, error) {
	key, err := c.findKey(id)
	if err != nil {
		return StoredKey{}, err
	}
	return *key, nil
}

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

func normalizeOutputFormat(format string) string {
	switch strings.ToLower(format) {
	case "base64":
		return "base64"
	default:
		return "hex"
	}
}

func encodeOutputBytes(data []byte, format string) string {
	switch normalizeOutputFormat(format) {
	case "base64":
		return encodeBase64(data)
	default:
		return strings.ToUpper(hex.EncodeToString(data))
	}
}
