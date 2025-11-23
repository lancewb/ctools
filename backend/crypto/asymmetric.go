package crypto

import (
	stdcrypto "crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"io"
	"math/big"
	"strings"

	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/sm9"
	"github.com/emmansun/gmsm/smx509"
	"golang.org/x/crypto/hkdf"
)

type keyMaterial struct {
	stored     *StoredKey
	privatePEM string
	publicPEM  string
}

type eccOptions struct {
	mode      string
	symmetric string
	mac       string
	kdf       stdcrypto.Hash
	macHash   stdcrypto.Hash
}

// RunAsymmetric performs an asymmetric cryptographic operation based on the request.
//
// req: The AsymmetricRequest containing algorithm, operation, and parameters.
// Returns an OperationResult with the output or an error.
func (c *CryptoService) RunAsymmetric(req AsymmetricRequest) (OperationResult, error) {
	switch strings.ToLower(req.Algorithm) {
	case "rsa":
		return c.runRSAOperation(req)
	case "ecc":
		return c.runECCOperation(req)
	case "sm2":
		return c.runSM2Operation(req)
	case "sm9":
		return c.runSM9Operation(req)
	default:
		return OperationResult{}, fmt.Errorf("unsupported algorithm: %s", req.Algorithm)
	}
}

func (c *CryptoService) runRSAOperation(req AsymmetricRequest) (OperationResult, error) {
	mat, err := c.resolveKeyMaterial(req.KeyID, req.KeyData, req.KeyFormat)
	if err != nil {
		return OperationResult{}, err
	}
	op := strings.ToLower(req.Operation)
	payload, err := decodeBlob(req.Payload, req.PayloadFormat)
	if err != nil {
		return OperationResult{}, err
	}
	outputFormat := normalizeOutputFormat(req.OutputFormat)
	padding := strings.ToLower(req.Padding)
	if padding == "" {
		padding = "oaep"
	}
	oaepHash, err := resolveHashAlgorithm(req.OAEPHash, stdcrypto.SHA256)
	if err != nil {
		return OperationResult{}, err
	}
	mgfHash, err := resolveHashAlgorithm(req.MGF1Hash, oaepHash)
	if err != nil {
		return OperationResult{}, err
	}

	switch op {
	case "encrypt":
		pub := mat.publicPEM
		if pub == "" && mat.privatePEM != "" {
			key, err := parseRSAPrivate(mat.privatePEM)
			if err != nil {
				return OperationResult{}, err
			}
			pubBytes, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
			pub = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
		}
		if pub == "" {
			return OperationResult{}, errors.New("missing RSA public key")
		}
		pubKey, err := parseRSAPublic(pub)
		if err != nil {
			return OperationResult{}, err
		}
		var ciphertext []byte
		if padding == "pkcs1" {
			ciphertext, err = rsa.EncryptPKCS1v15(rand.Reader, pubKey, payload)
		} else {
			ciphertext, err = encryptRSAOAEP(pubKey, payload, nil, oaepHash, mgfHash)
		}
		if err != nil {
			return OperationResult{}, err
		}
		return OperationResult{
			Output: encodeOutputBytes(ciphertext, outputFormat),
			Details: map[string]string{
				"base64": encodeBase64(ciphertext),
			},
		}, nil
	case "decrypt":
		if mat.privatePEM == "" {
			return OperationResult{}, errors.New("missing RSA private key")
		}
		priv, err := parseRSAPrivate(mat.privatePEM)
		if err != nil {
			return OperationResult{}, err
		}
		var plaintext []byte
		if padding == "pkcs1" {
			plaintext, err = rsa.DecryptPKCS1v15(rand.Reader, priv, payload)
		} else {
			plaintext, err = priv.Decrypt(rand.Reader, payload, &rsa.OAEPOptions{
				Hash:    oaepHash,
				MGFHash: mgfHash,
			})
		}
		if err != nil {
			return OperationResult{}, err
		}
		return OperationResult{
			Output: encodeOutputBytes(plaintext, outputFormat),
			Details: map[string]string{
				"text":   string(plaintext),
				"base64": encodeBase64(plaintext),
			},
		}, nil
	case "sign":
		if mat.privatePEM == "" {
			return OperationResult{}, errors.New("missing RSA private key")
		}
		priv, err := parseRSAPrivate(mat.privatePEM)
		if err != nil {
			return OperationResult{}, err
		}
		digest := sha256.Sum256(payload)
		sig, err := rsa.SignPSS(rand.Reader, priv, stdcrypto.SHA256, digest[:], nil)
		if err != nil {
			return OperationResult{}, err
		}
		return OperationResult{
			Output: encodeOutputBytes(sig, outputFormat),
			Details: map[string]string{
				"base64": encodeBase64(sig),
			},
		}, nil
	case "verify":
		pub := mat.publicPEM
		if pub == "" && mat.privatePEM != "" {
			priv, err := parseRSAPrivate(mat.privatePEM)
			if err != nil {
				return OperationResult{}, err
			}
			pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
			pub = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
		}
		if pub == "" {
			return OperationResult{}, errors.New("missing RSA public key")
		}
		pubKey, err := parseRSAPublic(pub)
		if err != nil {
			return OperationResult{}, err
		}
		signature, err := decodeBlob(req.Signature, req.SignatureFmt)
		if err != nil {
			return OperationResult{}, err
		}
		digest := sha256.Sum256(payload)
		if err := rsa.VerifyPSS(pubKey, stdcrypto.SHA256, digest[:], signature, nil); err != nil {
			return OperationResult{Verified: false}, nil
		}
		return OperationResult{Verified: true}, nil
	default:
		return OperationResult{}, fmt.Errorf("unsupported RSA operation: %s", req.Operation)
	}
}

func (c *CryptoService) runECCOperation(req AsymmetricRequest) (OperationResult, error) {
	mat, err := c.resolveKeyMaterial(req.KeyID, req.KeyData, req.KeyFormat)
	if err != nil {
		return OperationResult{}, err
	}
	opts, err := resolveECCOptions(req)
	if err != nil {
		return OperationResult{}, err
	}
	op := strings.ToLower(req.Operation)
	payload, err := decodeBlob(req.Payload, req.PayloadFormat)
	if err != nil {
		return OperationResult{}, err
	}
	outputFormat := normalizeOutputFormat(req.OutputFormat)

	switch op {
	case "encrypt":
		pub, err := ensureECCPublic(mat)
		if err != nil {
			return OperationResult{}, err
		}
		output, meta, err := eccEncrypt(pub, payload, opts)
		if err != nil {
			return OperationResult{}, err
		}
		return OperationResult{
			Output: encodeOutputBytes(output, outputFormat),
			Details: map[string]string{
				"base64":          encodeBase64(output),
				"mode":            meta.Mode,
				"cipher":          meta.Symmetric,
				"mac":             meta.Mac,
				"ephemeralLength": fmt.Sprintf("%d", meta.ephemeralLength),
				"nonceLength":     fmt.Sprintf("%d", meta.nonceLength),
			},
		}, nil
	case "decrypt":
		priv, err := parseECCPrivate(mat.privatePEM)
		if err != nil {
			return OperationResult{}, err
		}
		plaintext, err := eccDecrypt(priv, payload, opts)
		if err != nil {
			return OperationResult{}, err
		}
		return OperationResult{
			Output: encodeOutputBytes(plaintext, outputFormat),
			Details: map[string]string{
				"text":   string(plaintext),
				"base64": encodeBase64(plaintext),
				"mode":   opts.mode,
				"cipher": opts.symmetric,
			},
		}, nil
	case "sign":
		priv, err := parseECCPrivate(mat.privatePEM)
		if err != nil {
			return OperationResult{}, err
		}
		hashBytes := sha256.Sum256(payload)
		sig, err := ecdsa.SignASN1(rand.Reader, priv, hashBytes[:])
		if err != nil {
			return OperationResult{}, err
		}
		return OperationResult{
			Output: encodeOutputBytes(sig, outputFormat),
			Details: map[string]string{
				"base64": encodeBase64(sig),
			},
		}, nil
	case "verify":
		pub, err := ensureECCPublic(mat)
		if err != nil {
			return OperationResult{}, err
		}
		hashBytes := sha256.Sum256(payload)
		signature, err := decodeBlob(req.Signature, req.SignatureFmt)
		if err != nil {
			return OperationResult{}, err
		}
		ok := ecdsa.VerifyASN1(pub, hashBytes[:], signature)
		return OperationResult{Verified: ok}, nil
	default:
		return OperationResult{}, fmt.Errorf("unsupported ECC operation: %s", req.Operation)
	}
}

func (c *CryptoService) runSM2Operation(req AsymmetricRequest) (OperationResult, error) {
	mat, err := c.resolveKeyMaterial(req.KeyID, req.KeyData, req.KeyFormat)
	if err != nil {
		return OperationResult{}, err
	}
	op := strings.ToLower(req.Operation)
	payload, err := decodeBlob(req.Payload, req.PayloadFormat)
	if err != nil {
		return OperationResult{}, err
	}
	outputFormat := normalizeOutputFormat(req.OutputFormat)

	switch op {
	case "encrypt":
		pub, err := ensureSM2Public(mat)
		if err != nil {
			return OperationResult{}, err
		}
		ct, err := sm2.Encrypt(rand.Reader, pub, payload, nil)
		if err != nil {
			return OperationResult{}, err
		}
		return OperationResult{
			Output: encodeOutputBytes(ct, outputFormat),
			Details: map[string]string{
				"base64": encodeBase64(ct),
			},
		}, nil
	case "decrypt":
		priv, err := parseSM2Private(mat.privatePEM)
		if err != nil {
			return OperationResult{}, err
		}
		plaintext, err := sm2.Decrypt(priv, payload)
		if err != nil {
			return OperationResult{}, err
		}
		return OperationResult{
			Output: encodeOutputBytes(plaintext, outputFormat),
			Details: map[string]string{
				"text":   string(plaintext),
				"base64": encodeBase64(plaintext),
			},
		}, nil
	case "sign":
		priv, err := parseSM2Private(mat.privatePEM)
		if err != nil {
			return OperationResult{}, err
		}
		uid := []byte(req.UID)
		if len(uid) == 0 {
			uid = []byte("1234567812345678")
		}
		opts := sm2.NewSM2SignerOption(true, uid)
		sig, err := priv.Sign(rand.Reader, payload, opts)
		if err != nil {
			return OperationResult{}, err
		}
		return OperationResult{
			Output: encodeOutputBytes(sig, outputFormat),
			Details: map[string]string{
				"base64": encodeBase64(sig),
			},
		}, nil
	case "verify":
		pub, err := ensureSM2Public(mat)
		if err != nil {
			return OperationResult{}, err
		}
		signature, err := decodeBlob(req.Signature, req.SignatureFmt)
		if err != nil {
			return OperationResult{}, err
		}
		uid := []byte(req.UID)
		if len(uid) == 0 {
			uid = []byte("1234567812345678")
		}
		ok := sm2.VerifyASN1WithSM2(pub, uid, payload, signature)
		return OperationResult{Verified: ok}, nil
	default:
		return OperationResult{}, fmt.Errorf("unsupported SM2 operation: %s", req.Operation)
	}
}

func (c *CryptoService) runSM9Operation(req AsymmetricRequest) (OperationResult, error) {
	mat, err := c.resolveKeyMaterial(req.KeyID, req.KeyData, req.KeyFormat)
	if err != nil {
		return OperationResult{}, err
	}
	if mat.stored == nil {
		return OperationResult{}, errors.New("SM9 operations require a stored key with type context")
	}
	op := strings.ToLower(req.Operation)
	payload, err := decodeBlob(req.Payload, req.PayloadFormat)
	if err != nil {
		return OperationResult{}, err
	}
	uid := []byte(req.UID)
	if len(uid) == 0 {
		uid = []byte("default-user")
	}
	outputFormat := normalizeOutputFormat(req.OutputFormat)

	switch op {
	case "sign", "verify":
		return c.handleSM9Signature(mat, op, uid, payload, req, outputFormat)
	case "encrypt", "decrypt":
		return c.handleSM9Encryption(mat, op, uid, payload, req, outputFormat)
	default:
		return OperationResult{}, fmt.Errorf("unsupported SM9 operation: %s", op)
	}
}

func (c *CryptoService) handleSM9Signature(mat keyMaterial, op string, uid, payload []byte, req AsymmetricRequest, outputFormat string) (OperationResult, error) {
	switch op {
	case "sign":
		if mat.stored.KeyType != "sign-user" {
			return OperationResult{}, errors.New("SM9 signing requires a user private key")
		}
		priv, err := parseSM9SignPrivate(mat.privatePEM)
		if err != nil {
			return OperationResult{}, err
		}
		digest := sha512.Sum512(payload)
		sig, err := sm9.SignASN1(rand.Reader, priv, digest[:])
		if err != nil {
			return OperationResult{}, err
		}
		return OperationResult{
			Output: encodeOutputBytes(sig, outputFormat),
			Details: map[string]string{
				"base64": encodeBase64(sig),
			},
		}, nil
	case "verify":
		pub, err := deriveSM9SignPublic(mat)
		if err != nil {
			return OperationResult{}, err
		}
		sig, err := decodeBlob(req.Signature, req.SignatureFmt)
		if err != nil {
			return OperationResult{}, err
		}
		digest := sha512.Sum512(payload)
		ok := sm9.VerifyASN1(pub, uid, 0x01, digest[:], sig)
		return OperationResult{Verified: ok}, nil
	default:
		return OperationResult{}, fmt.Errorf("unsupported SM9 signature operation: %s", op)
	}
}

func (c *CryptoService) handleSM9Encryption(mat keyMaterial, op string, uid, payload []byte, req AsymmetricRequest, outputFormat string) (OperationResult, error) {
	switch op {
	case "encrypt":
		if mat.stored.KeyType != "encrypt-master" {
			return OperationResult{}, errors.New("SM9 encryption requires a master public key")
		}
		pub, err := deriveSM9EncryptPublic(mat)
		if err != nil {
			return OperationResult{}, err
		}
		ct, err := sm9.Encrypt(rand.Reader, pub, uid, 0x02, payload, nil)
		if err != nil {
			return OperationResult{}, err
		}
		return OperationResult{
			Output: encodeOutputBytes(ct, outputFormat),
			Details: map[string]string{
				"base64": encodeBase64(ct),
			},
		}, nil
	case "decrypt":
		if mat.stored.KeyType != "encrypt-user" {
			return OperationResult{}, errors.New("SM9 decryption requires a user private key")
		}
		priv, err := parseSM9EncryptPrivate(mat.privatePEM)
		if err != nil {
			return OperationResult{}, err
		}
		plaintext, err := sm9.Decrypt(priv, uid, payload, nil)
		if err != nil {
			return OperationResult{}, err
		}
		return OperationResult{
			Output: encodeOutputBytes(plaintext, outputFormat),
			Details: map[string]string{
				"text":   string(plaintext),
				"base64": encodeBase64(plaintext),
			},
		}, nil
	default:
		return OperationResult{}, fmt.Errorf("unsupported SM9 encryption operation: %s", op)
	}
}

func (c *CryptoService) resolveKeyMaterial(id, inline, format string) (keyMaterial, error) {
	mat := keyMaterial{}
	if strings.TrimSpace(id) != "" {
		key, err := c.findKey(id)
		if err != nil {
			return mat, err
		}
		copy := *key
		mat.stored = &copy
		mat.privatePEM = key.PrivatePEM
		mat.publicPEM = key.PublicPEM
		return mat, nil
	}
	content := strings.TrimSpace(inline)
	if content == "" {
		return mat, errors.New("no key material provided")
	}
	if strings.Contains(content, "BEGIN") {
		block, _ := pem.Decode([]byte(content))
		if block == nil {
			return mat, errors.New("invalid PEM block")
		}
		pemStr := string(pem.EncodeToMemory(block))
		if strings.Contains(block.Type, "PRIVATE") {
			mat.privatePEM = pemStr
		} else {
			mat.publicPEM = pemStr
		}
		return mat, nil
	}
	raw, err := decodeData(format, content)
	if err != nil {
		return mat, err
	}
	mat.privatePEM = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: raw}))
	return mat, nil
}

func resolveECCOptions(req AsymmetricRequest) (eccOptions, error) {
	opts := eccOptions{
		mode:      strings.ToLower(req.EccMode),
		symmetric: strings.ToLower(req.SymmetricCipher),
		mac:       strings.ToLower(req.MacAlgorithm),
	}
	if opts.mode == "" {
		opts.mode = "dhaes"
	}
	if opts.symmetric == "" {
		if opts.mode == "ecies" {
			opts.symmetric = "aes-256-cbc"
		} else {
			opts.symmetric = "aes-256-gcm"
		}
	}
	if opts.symmetric == "aes-256-cbc" && opts.mac == "" {
		opts.mac = "hmac-sha256"
	}
	kdf, err := resolveHashAlgorithm(req.KDF, stdcrypto.SHA256)
	if err != nil {
		return eccOptions{}, err
	}
	opts.kdf = kdf
	if opts.mac != "" {
		macHash, err := resolveMACAlgorithm(opts.mac)
		if err != nil {
			return eccOptions{}, err
		}
		opts.macHash = macHash
	}
	return opts, nil
}

func decodeBlob(data, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "base64", "b64":
		return base64.StdEncoding.DecodeString(strings.TrimSpace(data))
	case "hex":
		return hex.DecodeString(strings.ReplaceAll(strings.TrimSpace(data), " ", ""))
	default:
		return []byte(data), nil
	}
}

func parseRSAPrivate(p string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(p))
	if block == nil {
		return nil, errors.New("invalid RSA private key")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	default:
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
		return rsaKey, nil
	}
}

func parseRSAPublic(p string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(p))
	if block == nil {
		return nil, errors.New("invalid RSA public key")
	}
	if block.Type == "RSA PUBLIC KEY" {
		return x509.ParsePKCS1PublicKey(block.Bytes)
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return rsaPub, nil
}

func parseECCPrivate(p string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(p))
	if block == nil {
		return nil, errors.New("invalid EC private key")
	}
	switch block.Type {
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)
	default:
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		ecKey, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an EC private key")
		}
		return ecKey, nil
	}
}

func ensureECCPublic(mat keyMaterial) (*ecdsa.PublicKey, error) {
	if mat.publicPEM != "" {
		block, _ := pem.Decode([]byte(mat.publicPEM))
		if block == nil {
			return nil, errors.New("invalid EC public key")
		}
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		ec, ok := pub.(*ecdsa.PublicKey)
		if !ok {
			return nil, errors.New("not an EC public key")
		}
		return ec, nil
	}
	if mat.privatePEM == "" {
		return nil, errors.New("missing EC public key")
	}
	priv, err := parseECCPrivate(mat.privatePEM)
	if err != nil {
		return nil, err
	}
	return &priv.PublicKey, nil
}

type eccCipherMeta struct {
	ephemeralLength int
	nonceLength     int
	macLength       int
	Mode            string
	Symmetric       string
	Mac             string
}

func eccEncrypt(pub *ecdsa.PublicKey, plaintext []byte, opts eccOptions) ([]byte, eccCipherMeta, error) {
	meta := eccCipherMeta{}
	ephemeral, err := ecdsa.GenerateKey(pub.Curve, rand.Reader)
	if err != nil {
		return nil, meta, err
	}
	sharedX, sharedY := pub.Curve.ScalarMult(pub.X, pub.Y, ephemeral.D.Bytes())
	if sharedX == nil || sharedY == nil {
		return nil, meta, errors.New("failed to derive ECC shared secret")
	}
	shared := append(sharedX.Bytes(), sharedY.Bytes()...)

	meta.Mode = opts.mode
	meta.Symmetric = opts.symmetric
	meta.Mac = opts.mac

	ephemeralBytes := elliptic.Marshal(pub.Curve, ephemeral.PublicKey.X, ephemeral.PublicKey.Y)
	meta.ephemeralLength = len(ephemeralBytes)

	switch opts.symmetric {
	case "aes-256-cbc":
		keyMaterial, err := deriveECCKeys(shared, opts.kdf, 64, opts.symmetric)
		if err != nil {
			return nil, meta, err
		}
		encKey := keyMaterial[:32]
		macKey := keyMaterial[32:]
		block, err := aes.NewCipher(encKey)
		if err != nil {
			return nil, meta, err
		}
		iv := make([]byte, block.BlockSize())
		if _, err := rand.Read(iv); err != nil {
			return nil, meta, err
		}
		data := applyPadding(plaintext, block.BlockSize(), "pkcs7")
		ciphertext := make([]byte, len(data))
		mode := cipher.NewCBCEncrypter(block, iv)
		mode.CryptBlocks(ciphertext, data)
		tag, err := computeHMAC(opts.macHash, macKey, append(iv, ciphertext...))
		if err != nil {
			return nil, meta, err
		}
		meta.nonceLength = len(iv)
		meta.macLength = len(tag)
		payload := append(ephemeralBytes, iv...)
		payload = append(payload, ciphertext...)
		payload = append(payload, tag...)
		return payload, meta, nil
	default:
		keyMaterial, err := deriveECCKeys(shared, opts.kdf, 32, opts.symmetric)
		if err != nil {
			return nil, meta, err
		}
		block, err := aes.NewCipher(keyMaterial)
		if err != nil {
			return nil, meta, err
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, meta, err
		}
		nonce := make([]byte, gcm.NonceSize())
		if _, err := rand.Read(nonce); err != nil {
			return nil, meta, err
		}
		ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
		meta.nonceLength = len(nonce)
		payload := append(ephemeralBytes, nonce...)
		payload = append(payload, ciphertext...)
		return payload, meta, nil
	}
}

func eccDecrypt(priv *ecdsa.PrivateKey, payload []byte, opts eccOptions) ([]byte, error) {
	byteLen := (priv.Curve.Params().BitSize + 7) / 8
	ephemeralLen := 1 + 2*byteLen
	if len(payload) <= ephemeralLen {
		return nil, errors.New("invalid ECC ciphertext")
	}
	ephemeral := payload[:ephemeralLen]
	x, y := elliptic.Unmarshal(priv.Curve, ephemeral)
	if x == nil || y == nil {
		return nil, errors.New("invalid ECC ephemeral key")
	}
	sharedX, sharedY := priv.Curve.ScalarMult(x, y, priv.D.Bytes())
	if sharedX == nil || sharedY == nil {
		return nil, errors.New("failed to derive ECC shared secret")
	}
	shared := append(sharedX.Bytes(), sharedY.Bytes()...)
	switch opts.symmetric {
	case "aes-256-cbc":
		if opts.macHash == 0 {
			return nil, errors.New("hmac algorithm required for CBC mode")
		}
		keyMaterial, err := deriveECCKeys(shared, opts.kdf, 64, opts.symmetric)
		if err != nil {
			return nil, err
		}
		encKey := keyMaterial[:32]
		macKey := keyMaterial[32:]
		block, err := aes.NewCipher(encKey)
		if err != nil {
			return nil, err
		}
		blockSize := block.BlockSize()
		macSize := opts.macHash.Size()
		if len(payload) <= ephemeralLen+blockSize+macSize {
			return nil, errors.New("invalid ECC ciphertext")
		}
		iv := payload[ephemeralLen : ephemeralLen+blockSize]
		tail := payload[ephemeralLen+blockSize:]
		if len(tail) <= macSize {
			return nil, errors.New("invalid ECC ciphertext body")
		}
		macStart := len(tail) - macSize
		ciphertext := tail[:macStart]
		tag := tail[macStart:]
		expected, err := computeHMAC(opts.macHash, macKey, append(iv, ciphertext...))
		if err != nil {
			return nil, err
		}
		if !hmac.Equal(expected, tag) {
			return nil, errors.New("invalid MAC for ECC message")
		}
		if len(ciphertext)%blockSize != 0 {
			return nil, errors.New("ciphertext is not block aligned")
		}
		plain := make([]byte, len(ciphertext))
		mode := cipher.NewCBCDecrypter(block, iv)
		mode.CryptBlocks(plain, ciphertext)
		plain, err = removePadding(plain, blockSize, "pkcs7")
		if err != nil {
			return nil, err
		}
		return plain, nil
	default:
		keyMaterial, err := deriveECCKeys(shared, opts.kdf, 32, opts.symmetric)
		if err != nil {
			return nil, err
		}
		block, err := aes.NewCipher(keyMaterial)
		if err != nil {
			return nil, err
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}
		nonceLen := gcm.NonceSize()
		if len(payload) <= ephemeralLen+nonceLen {
			return nil, errors.New("invalid ECC ciphertext")
		}
		nonce := payload[ephemeralLen : ephemeralLen+nonceLen]
		ciphertext := payload[ephemeralLen+nonceLen:]
		return gcm.Open(nil, nonce, ciphertext, nil)
	}
}

func deriveECCKeys(shared []byte, hashAlg stdcrypto.Hash, length int, info string) ([]byte, error) {
	if hashAlg == 0 {
		hashAlg = stdcrypto.SHA256
	}
	fn, err := newHashFunc(hashAlg)
	if err != nil {
		return nil, err
	}
	reader := hkdf.New(fn, shared, nil, []byte(info))
	buf := make([]byte, length)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func computeHMAC(hashType stdcrypto.Hash, key, message []byte) ([]byte, error) {
	if hashType == 0 {
		return nil, errors.New("missing mac hash")
	}
	mac := hmac.New(hashType.New, key)
	_, err := mac.Write(message)
	if err != nil {
		return nil, err
	}
	return mac.Sum(nil), nil
}

func parseSM2Private(p string) (*sm2.PrivateKey, error) {
	block, _ := pem.Decode([]byte(p))
	if block == nil {
		return nil, errors.New("invalid SM2 private key")
	}
	switch block.Type {
	case "EC PRIVATE KEY":
		return smx509.ParseSM2PrivateKey(block.Bytes)
	default:
		key, err := smx509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		sm2Key, ok := key.(*sm2.PrivateKey)
		if !ok {
			return nil, errors.New("not an SM2 private key")
		}
		return sm2Key, nil
	}
}

func ensureSM2Public(mat keyMaterial) (*ecdsa.PublicKey, error) {
	if mat.publicPEM != "" {
		block, _ := pem.Decode([]byte(mat.publicPEM))
		if block == nil {
			return nil, errors.New("invalid SM2 public key")
		}
		pub, err := smx509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		if key, ok := pub.(*ecdsa.PublicKey); ok {
			return key, nil
		}
		return nil, errors.New("not an SM2 public key")
	}
	if mat.privatePEM == "" {
		return nil, errors.New("missing SM2 public key")
	}
	priv, err := parseSM2Private(mat.privatePEM)
	if err != nil {
		return nil, err
	}
	return &priv.PublicKey, nil
}

func parseSM9SignPrivate(p string) (*sm9.SignPrivateKey, error) {
	block, _ := pem.Decode([]byte(p))
	if block == nil {
		return nil, errors.New("invalid SM9 private key")
	}
	return sm9.UnmarshalSignPrivateKeyASN1(block.Bytes)
}

func parseSM9EncryptPrivate(p string) (*sm9.EncryptPrivateKey, error) {
	block, _ := pem.Decode([]byte(p))
	if block == nil {
		return nil, errors.New("invalid SM9 encrypt private key")
	}
	return sm9.UnmarshalEncryptPrivateKeyASN1(block.Bytes)
}

func deriveSM9SignPublic(mat keyMaterial) (*sm9.SignMasterPublicKey, error) {
	if mat.publicPEM != "" {
		block, _ := pem.Decode([]byte(mat.publicPEM))
		if block == nil {
			return nil, errors.New("invalid SM9 public key")
		}
		return sm9.UnmarshalSignMasterPublicKeyASN1(block.Bytes)
	}
	if mat.privatePEM != "" {
		block, _ := pem.Decode([]byte(mat.privatePEM))
		if block == nil {
			return nil, errors.New("invalid SM9 private key")
		}
		switch {
		case strings.Contains(block.Type, "SM9 SIGN MASTER"):
			priv, err := sm9.UnmarshalSignMasterPrivateKeyASN1(block.Bytes)
			if err != nil {
				return nil, err
			}
			return priv.PublicKey(), nil
		case strings.Contains(block.Type, "SM9 SIGN PRIVATE"):
			priv, err := sm9.UnmarshalSignPrivateKeyASN1(block.Bytes)
			if err != nil {
				return nil, err
			}
			return priv.MasterPublic(), nil
		}
	}
	return nil, errors.New("missing SM9 sign public key")
}

func deriveSM9EncryptPublic(mat keyMaterial) (*sm9.EncryptMasterPublicKey, error) {
	if mat.publicPEM != "" {
		block, _ := pem.Decode([]byte(mat.publicPEM))
		if block == nil {
			return nil, errors.New("invalid SM9 encrypt public key")
		}
		return sm9.UnmarshalEncryptMasterPublicKeyASN1(block.Bytes)
	}
	if mat.privatePEM != "" {
		block, _ := pem.Decode([]byte(mat.privatePEM))
		if block == nil {
			return nil, errors.New("invalid SM9 private key")
		}
		switch {
		case strings.Contains(block.Type, "SM9 ENCRYPT MASTER"):
			priv, err := sm9.UnmarshalEncryptMasterPrivateKeyASN1(block.Bytes)
			if err != nil {
				return nil, err
			}
			return priv.PublicKey(), nil
		case strings.Contains(block.Type, "SM9 ENCRYPT PRIVATE"):
			priv, err := sm9.UnmarshalEncryptPrivateKeyASN1(block.Bytes)
			if err != nil {
				return nil, err
			}
			return priv.MasterPublic(), nil
		}
	}
	return nil, errors.New("missing SM9 encrypt public key")
}

func resolveHashAlgorithm(name string, fallback stdcrypto.Hash) (stdcrypto.Hash, error) {
	if name == "" {
		if fallback != 0 {
			return fallback, nil
		}
		return stdcrypto.SHA256, nil
	}
	switch strings.ToLower(name) {
	case "sha1":
		return stdcrypto.SHA1, nil
	case "sha224":
		return stdcrypto.SHA224, nil
	case "sha256":
		return stdcrypto.SHA256, nil
	case "sha384":
		return stdcrypto.SHA384, nil
	case "sha512":
		return stdcrypto.SHA512, nil
	case "sha512/256", "sha512-256":
		return stdcrypto.SHA512_256, nil
	case "md5":
		return stdcrypto.MD5, nil
	default:
		return 0, fmt.Errorf("unsupported hash: %s", name)
	}
}

func resolveMACAlgorithm(name string) (stdcrypto.Hash, error) {
	switch strings.ToLower(name) {
	case "", "hmac-sha256":
		return stdcrypto.SHA256, nil
	case "hmac-sha384":
		return stdcrypto.SHA384, nil
	case "hmac-sha512":
		return stdcrypto.SHA512, nil
	default:
		return 0, fmt.Errorf("unsupported mac algorithm: %s", name)
	}
}

func newHashFunc(h stdcrypto.Hash) (func() hash.Hash, error) {
	if !h.Available() {
		return nil, fmt.Errorf("hash %v unavailable", h)
	}
	return func() hash.Hash {
		return h.New()
	}, nil
}

func encryptRSAOAEP(pub *rsa.PublicKey, msg, label []byte, hashAlg, mgfHash stdcrypto.Hash) ([]byte, error) {
	if hashAlg == 0 {
		hashAlg = stdcrypto.SHA256
	}
	if mgfHash == 0 {
		mgfHash = hashAlg
	}
	h, err := newHashFromType(hashAlg)
	if err != nil {
		return nil, err
	}
	mgf, err := newHashFromType(mgfHash)
	if err != nil {
		return nil, err
	}
	k := (pub.N.BitLen() + 7) / 8
	hLen := h.Size()
	if len(msg) > k-2*hLen-2 {
		return nil, errors.New("message too long for RSA OAEP")
	}
	hashLabel := h.Sum(label[:0])
	db := make([]byte, k-hLen-1)
	copy(db[:hLen], hashLabel)
	psLen := len(db) - len(msg) - 1 - hLen
	db[hLen+psLen] = 1
	copy(db[hLen+psLen+1:], msg)
	seed := make([]byte, hLen)
	if _, err := io.ReadFull(rand.Reader, seed); err != nil {
		return nil, err
	}
	dbMask := make([]byte, len(db))
	if err := mgf1XOR(dbMask, seed, mgf); err != nil {
		return nil, err
	}
	for i := range db {
		db[i] ^= dbMask[i]
	}
	seedMask := make([]byte, hLen)
	if err := mgf1XOR(seedMask, db, mgf); err != nil {
		return nil, err
	}
	for i := range seed {
		seed[i] ^= seedMask[i]
	}
	em := make([]byte, 1+hLen+len(db))
	em[0] = 0
	copy(em[1:1+hLen], seed)
	copy(em[1+hLen:], db)
	m := new(big.Int).SetBytes(em)
	if m.Cmp(pub.N) >= 0 {
		return nil, errors.New("message representative out of range")
	}
	e := big.NewInt(int64(pub.E))
	c := new(big.Int).Exp(m, e, pub.N)
	out := c.Bytes()
	if len(out) == k {
		return out, nil
	}
	result := make([]byte, k)
	copy(result[k-len(out):], out)
	return result, nil
}

func mgf1XOR(out []byte, seed []byte, h hash.Hash) error {
	counter := make([]byte, 4)
	var i int
	for i = 0; i < len(out); i += h.Size() {
		h.Reset()
		h.Write(seed)
		binary.BigEndian.PutUint32(counter, uint32(i/h.Size()))
		h.Write(counter)
		sum := h.Sum(nil)
		for j := 0; j < len(sum) && i+j < len(out); j++ {
			out[i+j] ^= sum[j]
		}
	}
	return nil
}

func newHashFromType(h stdcrypto.Hash) (hash.Hash, error) {
	if !h.Available() {
		return nil, fmt.Errorf("hash %v unavailable", h)
	}
	return h.New(), nil
}
