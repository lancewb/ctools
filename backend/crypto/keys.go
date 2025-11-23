package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/sm9"
	"github.com/emmansun/gmsm/smx509"
	"github.com/google/uuid"
)

// ParseKey parses key material and returns details about the key.
//
// req: The KeyParseRequest containing key data and format.
// Returns a KeyParseResult or an error.
func (c *CryptoService) ParseKey(req KeyParseRequest) (KeyParseResult, error) {
	result := KeyParseResult{
		Summary: map[string]string{},
	}
	if strings.TrimSpace(req.Data) == "" {
		return result, errors.New("key material is empty")
	}
	switch strings.ToLower(req.Algorithm) {
	case "rsa":
		return c.parseRSAKey(req)
	case "ecc":
		return c.parseECCKey(req)
	case "sm2":
		return c.parseSM2Key(req)
	case "sm9":
		return c.parseSM9Key(req)
	default:
		return result, fmt.Errorf("unsupported algorithm: %s", req.Algorithm)
	}
}

func (c *CryptoService) parseRSAKey(req KeyParseRequest) (KeyParseResult, error) {
	result := KeyParseResult{Summary: map[string]string{}}
	block, der, err := extractPEMOrDER(req.Data, req.Format)
	if err != nil {
		return result, err
	}
	var keyPEM string
	var pubPEM string
	var summary = map[string]string{}

	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
		pubBytes, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
		pubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
		summary["type"] = "private"
		summary["bits"] = fmt.Sprintf("%d", key.N.BitLen())
		summary["publicExponent"] = fmt.Sprintf("%d", key.E)
	} else if parsed, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		if key, ok := parsed.(*rsa.PrivateKey); ok {
			keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
			pubBytes, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
			pubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
			summary["type"] = "private"
			summary["bits"] = fmt.Sprintf("%d", key.N.BitLen())
			summary["publicExponent"] = fmt.Sprintf("%d", key.E)
		}
	}

	if keyPEM == "" {
		if parsed, err := x509.ParsePKIXPublicKey(der); err == nil {
			if key, ok := parsed.(*rsa.PublicKey); ok {
				pubBytes, _ := x509.MarshalPKIXPublicKey(key)
				pubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
				summary["type"] = "public"
				summary["bits"] = fmt.Sprintf("%d", key.N.BitLen())
				summary["publicExponent"] = fmt.Sprintf("%d", key.E)
			}
		}
	}

	if keyPEM == "" && pubPEM == "" {
		if block != nil {
			keyPEM = string(pem.EncodeToMemory(block))
		}
		return result, errors.New("failed to parse RSA material")
	}

	result.PrivatePEM = keyPEM
	result.PublicPEM = pubPEM
	result.Summary = summary
	if req.Save {
		stored := c.saveKey(StoredKey{
			ID:         uuid.New().String(),
			Name:       fallbackName(req.Name, "RSA"),
			Algorithm:  "RSA",
			KeyType:    summary["type"],
			Format:     formatLabel(req.Format),
			Usage:      req.Usage,
			PrivatePEM: keyPEM,
			PublicPEM:  pubPEM,
			Extra: map[string]string{
				"variant": req.Variant,
			},
			CreatedAt: time.Now(),
		})
		result.Stored = true
		result.Key = &stored
	}
	return result, nil
}

func (c *CryptoService) parseECCKey(req KeyParseRequest) (KeyParseResult, error) {
	result := KeyParseResult{Summary: map[string]string{}}
	block, der, err := extractPEMOrDER(req.Data, req.Format)
	if err != nil {
		return result, err
	}

	var priv *ecdsa.PrivateKey
	var pub *ecdsa.PublicKey

	if key, err := x509.ParseECPrivateKey(der); err == nil {
		priv = key
		pub = key.Public().(*ecdsa.PublicKey)
	} else if parsed, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		if key, ok := parsed.(*ecdsa.PrivateKey); ok {
			priv = key
			pub = key.Public().(*ecdsa.PublicKey)
		}
	}

	if priv == nil {
		if parsed, err := x509.ParsePKIXPublicKey(der); err == nil {
			if key, ok := parsed.(*ecdsa.PublicKey); ok {
				pub = key
			}
		}
	}

	if priv == nil && pub == nil {
		if block != nil {
			return result, fmt.Errorf("failed to parse ECC key from PEM block %s", block.Type)
		}
		return result, errors.New("failed to parse ECC key material")
	}

	var privPEM, pubPEM string
	if priv != nil {
		derPriv, err := x509.MarshalECPrivateKey(priv)
		if err == nil {
			privPEM = string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: derPriv}))
		}
		pubBytes, _ := x509.MarshalPKIXPublicKey(priv.Public())
		pubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
		result.Summary["type"] = "private"
	} else if pub != nil {
		pubBytes, _ := x509.MarshalPKIXPublicKey(pub)
		pubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
		result.Summary["type"] = "public"
	}

	if pub != nil {
		result.Summary["curve"] = pub.Curve.Params().Name
	}

	result.PrivatePEM = privPEM
	result.PublicPEM = pubPEM

	if req.Save {
		stored := c.saveKey(StoredKey{
			ID:         uuid.New().String(),
			Name:       fallbackName(req.Name, "ECC"),
			Algorithm:  "ECC",
			KeyType:    result.Summary["type"],
			Format:     formatLabel(req.Format),
			Usage:      req.Usage,
			PrivatePEM: privPEM,
			PublicPEM:  pubPEM,
			Extra: map[string]string{
				"curve":   result.Summary["curve"],
				"variant": req.Variant,
			},
			CreatedAt: time.Now(),
		})
		result.Stored = true
		result.Key = &stored
	}
	return result, nil
}

func (c *CryptoService) parseSM2Key(req KeyParseRequest) (KeyParseResult, error) {
	result := KeyParseResult{Summary: map[string]string{}}
	format := strings.ToLower(req.Format)

	var priv *sm2.PrivateKey
	var pub *ecdsa.PublicKey
	var err error
	var block *pem.Block
	var der []byte
	var derr error

	switch format {
	case "raw32", "raw":
		var data []byte
		data, err = decodeFlexible(req.Data)
		if err == nil {
			priv, err = sm2.ParseRawPrivateKey(data)
		}
	case "sdf":
		var data []byte
		data, err = decodeFlexible(req.Data)
		if err == nil {
			priv, err = parseSDFPrivateKey(data)
		}
	default:
		block, der, derr = extractPEMOrDER(req.Data, req.Format)
		if derr != nil {
			err = derr
		} else {
			if key, e := smx509.ParseSM2PrivateKey(der); e == nil {
				priv = key
			}
			if priv == nil {
				if parsed, e := smx509.ParsePKCS8PrivateKey(der); e == nil {
					switch k := parsed.(type) {
					case *sm2.PrivateKey:
						priv = k
					}
				}
			}
			if priv == nil {
				if parsed, e := smx509.ParsePKIXPublicKey(der); e == nil {
					if p, ok := parsed.(*ecdsa.PublicKey); ok {
						pub = p
					}
				}
			}
		}

		if err == nil && priv == nil && pub == nil && block != nil {
			err = fmt.Errorf("unsupported SM2 PEM block: %s", block.Type)
		}
	}

	if err != nil {
		return result, err
	}

	if priv == nil && pub == nil {
		return result, errors.New("failed to parse SM2 key material")
	}

	if priv != nil {
		pub = priv.Public().(*ecdsa.PublicKey)
	}

	var privPEM, pubPEM string
	if priv != nil {
		if der, err := smx509.MarshalSM2PrivateKey(priv); err == nil {
			privPEM = string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}))
		}
		pubBytes, _ := smx509.MarshalPKIXPublicKey(priv.Public())
		pubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
		result.Summary["type"] = "private"
	} else if pub != nil {
		pubBytes, _ := smx509.MarshalPKIXPublicKey(pub)
		pubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
		result.Summary["type"] = "public"
	}

	if pub != nil {
		result.Summary["curve"] = pub.Curve.Params().Name
	}

	result.PrivatePEM = privPEM
	result.PublicPEM = pubPEM
	if req.Save {
		stored := c.saveKey(StoredKey{
			ID:         uuid.New().String(),
			Name:       fallbackName(req.Name, "SM2"),
			Algorithm:  "SM2",
			KeyType:    result.Summary["type"],
			Format:     formatLabel(req.Format),
			Usage:      req.Usage,
			PrivatePEM: privPEM,
			PublicPEM:  pubPEM,
			Extra: map[string]string{
				"variant": req.Variant,
			},
			CreatedAt: time.Now(),
		})
		result.Stored = true
		result.Key = &stored
	}
	return result, nil
}

func (c *CryptoService) parseSM9Key(req KeyParseRequest) (KeyParseResult, error) {
	result := KeyParseResult{Summary: map[string]string{}}
	var privPEM string
	var summary = map[string]string{}
	data := strings.TrimSpace(req.Data)
	var priv any
	var err error

	if strings.Contains(data, "BEGIN") || strings.ToLower(req.Format) == "pem" {
		block, _, derr := extractPEMOrDER(data, "pem")
		if derr != nil {
			return result, derr
		}
		if block != nil {
			switch {
			case strings.Contains(block.Type, "SM9 SIGN MASTER"):
				priv, err = sm9.UnmarshalSignMasterPrivateKeyASN1(block.Bytes)
				summary["type"] = "sign-master"
			case strings.Contains(block.Type, "SM9 ENCRYPT MASTER"):
				priv, err = sm9.UnmarshalEncryptMasterPrivateKeyASN1(block.Bytes)
				summary["type"] = "encrypt-master"
			case strings.Contains(block.Type, "SM9 SIGN PRIVATE"):
				priv, err = sm9.UnmarshalSignPrivateKeyASN1(block.Bytes)
				summary["type"] = "sign-user"
			case strings.Contains(block.Type, "SM9 ENCRYPT PRIVATE"):
				priv, err = sm9.UnmarshalEncryptPrivateKeyASN1(block.Bytes)
				summary["type"] = "encrypt-user"
			}
			if err != nil {
				return result, err
			}
			privPEM = string(pem.EncodeToMemory(block))
		}
	}

	if priv == nil {
		raw, err := decodeFlexible(data)
		if err != nil {
			return result, err
		}
		switch strings.ToLower(req.Variant) {
		case "sign-master":
			priv, err = sm9.UnmarshalSignMasterPrivateKeyASN1(raw)
			summary["type"] = "sign-master"
		case "encrypt-master":
			priv, err = sm9.UnmarshalEncryptMasterPrivateKeyASN1(raw)
			summary["type"] = "encrypt-master"
		case "sign-user":
			priv, err = sm9.UnmarshalSignPrivateKeyASN1(raw)
			summary["type"] = "sign-user"
		case "encrypt-user":
			priv, err = sm9.UnmarshalEncryptPrivateKeyASN1(raw)
			summary["type"] = "encrypt-user"
		default:
			priv, err = sm9.UnmarshalSignPrivateKeyASN1(raw)
			summary["type"] = "sign-user"
		}
		if err != nil {
			return result, err
		}
		privPEM = encodeSM9PEM(priv)
	}

	result.PrivatePEM = privPEM
	result.Summary = summary
	if req.Save {
		stored := c.saveKey(StoredKey{
			ID:         uuid.New().String(),
			Name:       fallbackName(req.Name, "SM9"),
			Algorithm:  "SM9",
			KeyType:    summary["type"],
			Format:     formatLabel(req.Format),
			Usage:      req.Usage,
			PrivatePEM: privPEM,
			Extra: map[string]string{
				"variant": req.Variant,
			},
			CreatedAt: time.Now(),
		})
		result.Stored = true
		result.Key = &stored
	}
	return result, nil
}

// GenerateKeyPair generates a new key pair for the specified algorithm.
//
// req: The KeyGenRequest containing parameters like algorithm, size, and curve.
// Returns a KeyParseResult containing the generated keys or an error.
func (c *CryptoService) GenerateKeyPair(req KeyGenRequest) (KeyParseResult, error) {
	switch strings.ToLower(req.Algorithm) {
	case "rsa":
		return c.generateRSA(req)
	case "ecc":
		return c.generateECC(req)
	case "sm2":
		return c.generateSM2(req)
	case "sm9":
		return c.generateSM9(req)
	default:
		return KeyParseResult{}, fmt.Errorf("unsupported algorithm: %s", req.Algorithm)
	}
}

func (c *CryptoService) generateRSA(req KeyGenRequest) (KeyParseResult, error) {
	bits := req.KeySize
	if bits == 0 {
		bits = 2048
	}
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return KeyParseResult{}, err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pubBytes, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	result := KeyParseResult{
		PrivatePEM: string(privPEM),
		PublicPEM:  string(pubPEM),
		Summary: map[string]string{
			"type":           "private",
			"bits":           fmt.Sprintf("%d", bits),
			"publicExponent": fmt.Sprintf("%d", key.E),
		},
	}
	if req.Save {
		stored := c.saveKey(StoredKey{
			ID:         uuid.New().String(),
			Name:       fallbackName(req.Name, "RSA"),
			Algorithm:  "RSA",
			KeyType:    "private",
			Format:     "generated",
			Usage:      req.Usage,
			PrivatePEM: string(privPEM),
			PublicPEM:  string(pubPEM),
			CreatedAt:  time.Now(),
		})
		result.Stored = true
		result.Key = &stored
	}
	return result, nil
}

func (c *CryptoService) generateECC(req KeyGenRequest) (KeyParseResult, error) {
	curve := strings.ToUpper(req.Curve)
	if curve == "" {
		curve = "P256"
	}
	curveVal := elliptic.P256()
	switch curve {
	case "P256", "SECP256R1":
		curveVal = elliptic.P256()
	case "P384":
		curveVal = elliptic.P384()
	case "P521":
		curveVal = elliptic.P521()
	}
	key, err := ecdsa.GenerateKey(curveVal, rand.Reader)
	if err != nil {
		return KeyParseResult{}, err
	}
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return KeyParseResult{}, err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der})
	pubBytes, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	result := KeyParseResult{
		PrivatePEM: string(privPEM),
		PublicPEM:  string(pubPEM),
		Summary: map[string]string{
			"type":  "private",
			"curve": curveVal.Params().Name,
		},
	}
	if req.Save {
		stored := c.saveKey(StoredKey{
			ID:         uuid.New().String(),
			Name:       fallbackName(req.Name, "ECC"),
			Algorithm:  "ECC",
			KeyType:    "private",
			Format:     "generated",
			Usage:      req.Usage,
			PrivatePEM: string(privPEM),
			PublicPEM:  string(pubPEM),
			Extra: map[string]string{
				"curve": curveVal.Params().Name,
			},
			CreatedAt: time.Now(),
		})
		result.Stored = true
		result.Key = &stored
	}
	return result, nil
}

func (c *CryptoService) generateSM2(req KeyGenRequest) (KeyParseResult, error) {
	key, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return KeyParseResult{}, err
	}
	der, err := smx509.MarshalSM2PrivateKey(key)
	if err != nil {
		return KeyParseResult{}, err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der})
	pubBytes, _ := smx509.MarshalPKIXPublicKey(key.Public())
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	result := KeyParseResult{
		PrivatePEM: string(privPEM),
		PublicPEM:  string(pubPEM),
		Summary: map[string]string{
			"type":  "private",
			"curve": "SM2P256",
		},
	}
	if req.Save {
		stored := c.saveKey(StoredKey{
			ID:         uuid.New().String(),
			Name:       fallbackName(req.Name, "SM2"),
			Algorithm:  "SM2",
			KeyType:    "private",
			Format:     "generated",
			Usage:      req.Usage,
			PrivatePEM: string(privPEM),
			PublicPEM:  string(pubPEM),
			CreatedAt:  time.Now(),
		})
		result.Stored = true
		result.Key = &stored
	}
	return result, nil
}

func (c *CryptoService) generateSM9(req KeyGenRequest) (KeyParseResult, error) {
	var privPEM string
	var summary = map[string]string{}
	var err error

	variant := strings.ToLower(req.Variant)
	if variant == "" {
		variant = "sign-master"
	}

	switch variant {
	case "sign-master":
		var priv *sm9.SignMasterPrivateKey
		priv, err = sm9.GenerateSignMasterKey(rand.Reader)
		if err == nil {
			der, _ := priv.MarshalASN1()
			privPEM = string(pem.EncodeToMemory(&pem.Block{Type: "SM9 SIGN MASTER PRIVATE KEY", Bytes: der}))
			summary["type"] = "sign-master"
		}
	case "encrypt-master":
		var priv *sm9.EncryptMasterPrivateKey
		priv, err = sm9.GenerateEncryptMasterKey(rand.Reader)
		if err == nil {
			der, _ := priv.MarshalASN1()
			privPEM = string(pem.EncodeToMemory(&pem.Block{Type: "SM9 ENCRYPT MASTER PRIVATE KEY", Bytes: der}))
			summary["type"] = "encrypt-master"
		}
	default:
		err = errors.New("unsupported SM9 variant for generation")
	}

	if err != nil {
		return KeyParseResult{}, err
	}
	result := KeyParseResult{
		PrivatePEM: privPEM,
		Summary:    summary,
	}
	if req.Save {
		stored := c.saveKey(StoredKey{
			ID:         uuid.New().String(),
			Name:       fallbackName(req.Name, "SM9"),
			Algorithm:  "SM9",
			KeyType:    summary["type"],
			Format:     "generated",
			Usage:      req.Usage,
			PrivatePEM: privPEM,
			CreatedAt:  time.Now(),
		})
		result.Stored = true
		result.Key = &stored
	}
	return result, nil
}

func fallbackName(name, prefix string) string {
	if strings.TrimSpace(name) != "" {
		return name
	}
	return fmt.Sprintf("%s-%s", prefix, time.Now().Format("20060102150405"))
}

func formatLabel(fmt string) string {
	if fmt == "" {
		return "auto"
	}
	return strings.ToLower(fmt)
}

func decodeFlexible(data string) ([]byte, error) {
	trimmed := strings.TrimSpace(data)
	if trimmed == "" {
		return nil, errors.New("empty input")
	}
	if strings.HasPrefix(trimmed, "0x") {
		trimmed = trimmed[2:]
	}
	if b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(trimmed)); err == nil {
		return b, nil
	}
	trimmed = strings.ReplaceAll(trimmed, " ", "")
	return hex.DecodeString(trimmed)
}

func extractPEMOrDER(data, format string) (*pem.Block, []byte, error) {
	body := []byte(data)
	block, rest := pem.Decode(body)
	if block != nil {
		return block, block.Bytes, nil
	}
	der, err := decodeData(format, data)
	if err != nil {
		return nil, nil, err
	}
	if len(rest) > 0 {
		return nil, nil, errors.New("unexpected data after pem block")
	}
	return nil, der, nil
}

func parseSDFPrivateKey(bytes []byte) (*sm2.PrivateKey, error) {
	if len(bytes) < 40 {
		return nil, errors.New("sdf blob too short")
	}
	d := bytes[8:40]
	return sm2.ParseRawPrivateKey(d)
}

func encodeSM9PEM(key any) string {
	switch k := key.(type) {
	case *sm9.SignMasterPrivateKey:
		der, _ := k.MarshalASN1()
		return string(pem.EncodeToMemory(&pem.Block{Type: "SM9 SIGN MASTER PRIVATE KEY", Bytes: der}))
	case *sm9.EncryptMasterPrivateKey:
		der, _ := k.MarshalASN1()
		return string(pem.EncodeToMemory(&pem.Block{Type: "SM9 ENCRYPT MASTER PRIVATE KEY", Bytes: der}))
	case *sm9.SignPrivateKey:
		der, _ := k.MarshalASN1()
		return string(pem.EncodeToMemory(&pem.Block{Type: "SM9 SIGN PRIVATE KEY", Bytes: der}))
	case *sm9.EncryptPrivateKey:
		der, _ := k.MarshalASN1()
		return string(pem.EncodeToMemory(&pem.Block{Type: "SM9 ENCRYPT PRIVATE KEY", Bytes: der}))
	default:
		return ""
	}
}
