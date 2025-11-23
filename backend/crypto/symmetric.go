package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"errors"
	"fmt"
	"strings"

	"github.com/emmansun/gmsm/sm4"
	"golang.org/x/crypto/chacha20poly1305"
)

func (c *CryptoService) RunSymmetric(req SymmetricRequest) (OperationResult, error) {
	algo := strings.ToLower(req.Algorithm)
	op := strings.ToLower(req.Operation)
	if op != "encrypt" && op != "decrypt" {
		return OperationResult{}, errors.New("operation must be encrypt or decrypt")
	}

	input, err := decodeBlob(req.Input, req.InputFormat)
	if err != nil {
		return OperationResult{}, fmt.Errorf("invalid input: %w", err)
	}
	key, err := decodeBlob(req.Key, req.KeyFormat)
	if err != nil {
		return OperationResult{}, fmt.Errorf("invalid key: %w", err)
	}
	mode := strings.ToLower(req.Mode)
	padding := strings.ToLower(req.Padding)

	switch algo {
	case "aes":
		return runAESCipher(op, mode, padding, key, input, req)
	case "sm4":
		return runSM4Cipher(op, mode, padding, key, input, req)
	case "3des", "des3", "triple-des":
		return runTripleDESCipher(op, mode, padding, key, input, req)
	case "chacha20", "cha20", "chacha20-poly1305":
		return runChaChaCipher(op, key, input, req)
	default:
		return OperationResult{}, fmt.Errorf("unsupported symmetric algorithm: %s", req.Algorithm)
	}
}

func runAESCipher(op, mode, padding string, key, input []byte, req SymmetricRequest) (OperationResult, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return OperationResult{}, errors.New("AES key must be 16/24/32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return OperationResult{}, err
	}
	switch mode {
	case "gcm":
		return runBlockGCM(op, block, input, req)
	case "ctr":
		return runStreamBlock(op, cipher.NewCTR, block, input, req)
	case "ecb":
		return runECB(op, block, input, padding, req)
	default: // default to CBC
		return runCBC(op, block, input, padding, req)
	}
}

func runSM4Cipher(op, mode, padding string, key, input []byte, req SymmetricRequest) (OperationResult, error) {
	if len(key) != 16 {
		return OperationResult{}, errors.New("SM4 key must be 16 bytes")
	}
	block, err := sm4.NewCipher(key)
	if err != nil {
		return OperationResult{}, err
	}
	switch mode {
	case "ctr":
		return runStreamBlock(op, cipher.NewCTR, block, input, req)
	default:
		return runCBC(op, block, input, padding, req)
	}
}

func runTripleDESCipher(op, mode, padding string, key, input []byte, req SymmetricRequest) (OperationResult, error) {
	if len(key) != 24 {
		return OperationResult{}, errors.New("3DES key must be 24 bytes")
	}
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return OperationResult{}, err
	}
	if mode == "ctr" {
		return runStreamBlock(op, cipher.NewCTR, block, input, req)
	}
	return runCBC(op, block, input, padding, req)
}

func runChaChaCipher(op string, key, input []byte, req SymmetricRequest) (OperationResult, error) {
	if len(key) != chacha20poly1305.KeySize {
		return OperationResult{}, errors.New("ChaCha20-Poly1305 key must be 32 bytes")
	}
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return OperationResult{}, err
	}
	nonce, err := decodeBlob(req.Nonce, req.NonceFormat)
	if err != nil {
		return OperationResult{}, fmt.Errorf("invalid nonce: %w", err)
	}
	if len(nonce) != chacha20poly1305.NonceSize {
		return OperationResult{}, fmt.Errorf("nonce must be %d bytes", chacha20poly1305.NonceSize)
	}
	ad, err := decodeBlob(req.Additional, req.AdditionalFormat)
	if err != nil {
		return OperationResult{}, fmt.Errorf("invalid additional data: %w", err)
	}
	switch op {
	case "encrypt":
		out := aead.Seal(nil, nonce, input, ad)
		return OperationResult{
			Output: encodeOutputBytes(out, req.OutputFormat),
			Details: map[string]string{
				"base64": encodeBase64(out),
			},
		}, nil
	default:
		plain, err := aead.Open(nil, nonce, input, ad)
		if err != nil {
			return OperationResult{}, err
		}
		return OperationResult{
			Output: encodeOutputBytes(plain, req.OutputFormat),
			Details: map[string]string{
				"text":   string(plain),
				"base64": encodeBase64(plain),
			},
		}, nil
	}
}

func runBlockGCM(op string, block cipher.Block, input []byte, req SymmetricRequest) (OperationResult, error) {
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return OperationResult{}, err
	}
	nonce, err := decodeBlob(req.Nonce, req.NonceFormat)
	if err != nil {
		return OperationResult{}, fmt.Errorf("invalid nonce: %w", err)
	}
	if len(nonce) != aead.NonceSize() {
		return OperationResult{}, fmt.Errorf("nonce must be %d bytes", aead.NonceSize())
	}
	ad, err := decodeBlob(req.Additional, req.AdditionalFormat)
	if err != nil {
		return OperationResult{}, fmt.Errorf("invalid additional data: %w", err)
	}
	if op == "encrypt" {
		out := aead.Seal(nil, nonce, input, ad)
		return OperationResult{
			Output: encodeOutputBytes(out, req.OutputFormat),
			Details: map[string]string{
				"base64": encodeBase64(out),
			},
		}, nil
	}
	plain, err := aead.Open(nil, nonce, input, ad)
	if err != nil {
		return OperationResult{}, err
	}
	return OperationResult{
		Output: encodeOutputBytes(plain, req.OutputFormat),
		Details: map[string]string{
			"text":   string(plain),
			"base64": encodeBase64(plain),
		},
	}, nil
}

func runStreamBlock(op string, mode func(cipher.Block, []byte) cipher.Stream, block cipher.Block, input []byte, req SymmetricRequest) (OperationResult, error) {
	iv, err := decodeBlob(req.IV, req.IVFormat)
	if err != nil {
		return OperationResult{}, fmt.Errorf("invalid IV: %w", err)
	}
	if len(iv) != block.BlockSize() {
		return OperationResult{}, fmt.Errorf("IV must be %d bytes", block.BlockSize())
	}
	stream := mode(block, iv)
	buf := make([]byte, len(input))
	stream.XORKeyStream(buf, input)
	if op == "encrypt" {
		return OperationResult{
			Output: encodeOutputBytes(buf, req.OutputFormat),
			Details: map[string]string{
				"base64": encodeBase64(buf),
			},
		}, nil
	}
	return OperationResult{
		Output: encodeOutputBytes(buf, req.OutputFormat),
		Details: map[string]string{
			"text":   string(buf),
			"base64": encodeBase64(buf),
		},
	}, nil
}

func runCBC(op string, block cipher.Block, input []byte, padding string, req SymmetricRequest) (OperationResult, error) {
	iv, err := decodeBlob(req.IV, req.IVFormat)
	if err != nil {
		return OperationResult{}, fmt.Errorf("invalid IV: %w", err)
	}
	if len(iv) != block.BlockSize() {
		return OperationResult{}, fmt.Errorf("IV must be %d bytes", block.BlockSize())
	}
	if op == "encrypt" {
		plain := applyPadding(input, block.BlockSize(), padding)
		out := make([]byte, len(plain))
		mode := cipher.NewCBCEncrypter(block, iv)
		mode.CryptBlocks(out, plain)
		return OperationResult{
			Output: encodeOutputBytes(out, req.OutputFormat),
			Details: map[string]string{
				"base64": encodeBase64(out),
			},
		}, nil
	}
	if len(input)%block.BlockSize() != 0 {
		return OperationResult{}, errors.New("ciphertext not aligned to block size")
	}
	out := make([]byte, len(input))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(out, input)
	plain, err := removePadding(out, block.BlockSize(), padding)
	if err != nil {
		return OperationResult{}, err
	}
	return OperationResult{
		Output: encodeOutputBytes(plain, req.OutputFormat),
		Details: map[string]string{
			"text":   string(plain),
			"base64": encodeBase64(plain),
		},
	}, nil
}

func runECB(op string, block cipher.Block, input []byte, padding string, req SymmetricRequest) (OperationResult, error) {
	bs := block.BlockSize()
	if op == "encrypt" {
		plain := applyPadding(input, bs, padding)
		out := make([]byte, len(plain))
		for i := 0; i < len(plain); i += bs {
			block.Encrypt(out[i:i+bs], plain[i:i+bs])
		}
		return OperationResult{
			Output: encodeOutputBytes(out, req.OutputFormat),
			Details: map[string]string{
				"base64": encodeBase64(out),
			},
		}, nil
	}
	if len(input)%bs != 0 {
		return OperationResult{}, errors.New("ciphertext not aligned to block size")
	}
	out := make([]byte, len(input))
	for i := 0; i < len(input); i += bs {
		block.Decrypt(out[i:i+bs], input[i:i+bs])
	}
	plain, err := removePadding(out, bs, padding)
	if err != nil {
		return OperationResult{}, err
	}
	return OperationResult{
		Output: encodeOutputBytes(plain, req.OutputFormat),
		Details: map[string]string{
			"text":   string(plain),
			"base64": encodeBase64(plain),
		},
	}, nil
}

func applyPadding(data []byte, blockSize int, padding string) []byte {
	switch padding {
	case "zero":
		padLen := blockSize - (len(data) % blockSize)
		if padLen == blockSize {
			return data
		}
		return append(data, make([]byte, padLen)...)
	case "none":
		return data
	default:
		if blockSize == 0 {
			return data
		}
		padLen := blockSize - (len(data) % blockSize)
		pad := bytesRepeat(byte(padLen), padLen)
		return append(data, pad...)
	}
}

func removePadding(data []byte, blockSize int, padding string) ([]byte, error) {
	switch padding {
	case "zero":
		i := len(data)
		for i > 0 && data[i-1] == 0x00 {
			i--
		}
		return data[:i], nil
	case "none":
		return data, nil
	default:
		if len(data) == 0 {
			return nil, errors.New("empty data")
		}
		padLen := int(data[len(data)-1])
		if padLen == 0 || padLen > blockSize || padLen > len(data) {
			return nil, errors.New("invalid padding")
		}
		for _, b := range data[len(data)-padLen:] {
			if int(b) != padLen {
				return nil, errors.New("invalid padding content")
			}
		}
		return data[:len(data)-padLen], nil
	}
}

func bytesRepeat(b byte, count int) []byte {
	out := make([]byte, count)
	for i := range out {
		out[i] = b
	}
	return out
}
