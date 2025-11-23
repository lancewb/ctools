package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"strings"

	"github.com/emmansun/gmsm/sm3"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
)

// RunHash performs a hash or HMAC operation.
//
// req: The HashRequest specifying the algorithm, mode (hash/hmac), input, and optional key.
// Returns an OperationResult with the hash digest or an error.
func (c *CryptoService) RunHash(req HashRequest) (OperationResult, error) {
	algo := strings.ToLower(req.Algorithm)
	mode := strings.ToLower(req.Mode)
	if mode == "" {
		mode = "hash"
	}
	input, err := decodeBlob(req.Input, req.InputFormat)
	if err != nil {
		return OperationResult{}, fmt.Errorf("invalid input: %w", err)
	}
	hashFunc, err := selectHashFunc(algo)
	if err != nil {
		return OperationResult{}, err
	}
	var digest []byte
	switch mode {
	case "hash":
		h := hashFunc()
		h.Write(input)
		digest = h.Sum(nil)
	case "hmac":
		key, err := decodeBlob(req.Key, req.KeyFormat)
		if err != nil {
			return OperationResult{}, fmt.Errorf("invalid HMAC key: %w", err)
		}
		h := hmac.New(hashFunc, key)
		h.Write(input)
		digest = h.Sum(nil)
	default:
		return OperationResult{}, errors.New("mode must be hash or hmac")
	}
	return OperationResult{
		Output: encodeOutputBytes(digest, req.OutputFormat),
		Details: map[string]string{
			"base64": encodeBase64(digest),
			"hex":    strings.ToUpper(fmt.Sprintf("%x", digest)),
		},
	}, nil
}

func selectHashFunc(name string) (func() hash.Hash, error) {
	switch strings.ToLower(name) {
	case "sha1":
		return sha1.New, nil
	case "sha256":
		return sha256.New, nil
	case "sha512":
		return sha512.New, nil
	case "md5":
		return md5.New, nil
	case "sm3":
		return sm3.New, nil
	case "blake2b", "blake2b-256":
		return func() hash.Hash {
			h, _ := blake2b.New256(nil)
			return h
		}, nil
	case "blake2s", "blake2s-256":
		return func() hash.Hash {
			h, _ := blake2s.New256(nil)
			return h
		}, nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", name)
	}
}
