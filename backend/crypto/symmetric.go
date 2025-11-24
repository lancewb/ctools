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

// RunSymmetric performs a symmetric cryptographic operation.
//
// req: The SymmetricRequest containing algorithm, mode, key, IV, and input data.
// Returns an OperationResult with the output or an error.
func (c *CryptoService) RunSymmetric(req SymmetricRequest) (OperationResult, error) {
	algo := strings.ToLower(req.Algorithm)
	op := strings.ToLower(req.Operation)
	input, err := decodeBlob(req.Input, req.InputFormat)
	if err != nil {
		return OperationResult{}, fmt.Errorf("invalid input: %w", err)
	}
	key, err := decodeBlob(req.Key, req.KeyFormat)
	if err != nil {
		return OperationResult{}, fmt.Errorf("invalid key: %w", err)
	}
	switch op {
	case "cmac":
		return runCMACOperation(algo, key, input, req)
	case "diversify8", "div8":
		return runDiversifyOperation(algo, 8, key, input, req)
	case "diversify16", "div16":
		return runDiversifyOperation(algo, 16, key, input, req)
	}
	if op != "encrypt" && op != "decrypt" {
		return OperationResult{}, errors.New("operation must be encrypt or decrypt")
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

func runCMACOperation(algo string, key, input []byte, req SymmetricRequest) (OperationResult, error) {
	block, err := resolveBlockCipher(algo, key)
	if err != nil {
		return OperationResult{}, err
	}
	tag, err := computeCMAC(block, input)
	if err != nil {
		return OperationResult{}, err
	}
	return OperationResult{
		Output: encodeOutputBytes(tag, req.OutputFormat),
		Details: map[string]string{
			"base64": encodeBase64(tag),
		},
	}, nil
}

func runDiversifyOperation(algo string, blockSize int, key, diversifier []byte, req SymmetricRequest) (OperationResult, error) {
	block, err := resolveBlockCipher(algo, key)
	if err != nil {
		return OperationResult{}, err
	}
	if block.BlockSize() != blockSize {
		return OperationResult{}, fmt.Errorf("%s diversification requires %d-byte block cipher", strings.ToUpper(algo), blockSize)
	}
	if len(diversifier) != blockSize {
		return OperationResult{}, fmt.Errorf("diversifier must be %d bytes", blockSize)
	}
	out := make([]byte, blockSize)
	block.Encrypt(out, diversifier)
	return OperationResult{
		Output: encodeOutputBytes(out, req.OutputFormat),
		Details: map[string]string{
			"base64": encodeBase64(out),
		},
	}, nil
}

func resolveBlockCipher(algo string, key []byte) (cipher.Block, error) {
	switch strings.ToLower(algo) {
	case "aes":
		if len(key) != 16 && len(key) != 24 && len(key) != 32 {
			return nil, errors.New("AES key must be 16/24/32 bytes")
		}
		return aes.NewCipher(key)
	case "sm4":
		if len(key) != 16 {
			return nil, errors.New("SM4 key must be 16 bytes")
		}
		return sm4.NewCipher(key)
	case "3des", "des3", "triple-des":
		if len(key) != 24 {
			return nil, errors.New("3DES key must be 24 bytes")
		}
		return des.NewTripleDESCipher(key)
	default:
		return nil, fmt.Errorf("unsupported block cipher for %s", algo)
	}
}

func computeCMAC(block cipher.Block, msg []byte) ([]byte, error) {
	k1, k2, err := cmacSubkeys(block)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	var blockCount int
	complete := false
	if len(msg) == 0 {
		blockCount = 1
	} else {
		blockCount = len(msg) / bs
		complete = len(msg)%bs == 0
		if !complete {
			blockCount++
		}
	}
	last := make([]byte, bs)
	if len(msg) > 0 {
		copy(last, msg[(blockCount-1)*bs:])
	}
	if complete && len(msg) > 0 {
		xorBytes(last, last, k1)
	} else {
		padIndex := len(msg) % bs
		last[padIndex] = 0x80
		for i := padIndex + 1; i < bs; i++ {
			last[i] = 0
		}
		xorBytes(last, last, k2)
	}
	x := make([]byte, bs)
	buf := make([]byte, bs)
	for i := 0; i < blockCount-1; i++ {
		blockData := msg[i*bs : (i+1)*bs]
		xorBytes(buf, x, blockData)
		block.Encrypt(x, buf)
	}
	xorBytes(buf, x, last)
	block.Encrypt(x, buf)
	return x, nil
}

func cmacSubkeys(block cipher.Block) ([]byte, []byte, error) {
	bs := block.BlockSize()
	rb, err := cmacRbConstant(bs)
	if err != nil {
		return nil, nil, err
	}
	L := make([]byte, bs)
	block.Encrypt(L, make([]byte, bs))
	k1 := cmacDouble(L, rb)
	k2 := cmacDouble(k1, rb)
	return k1, k2, nil
}

func cmacDouble(input []byte, rb byte) []byte {
	out := make([]byte, len(input))
	var carry byte
	for i := len(input) - 1; i >= 0; i-- {
		b := input[i]
		out[i] = (b << 1) | carry
		carry = (b >> 7) & 0x01
	}
	if carry != 0 {
		out[len(out)-1] ^= rb
	}
	return out
}

func cmacRbConstant(blockSize int) (byte, error) {
	switch blockSize {
	case 16:
		return 0x87, nil
	case 8:
		return 0x1B, nil
	default:
		return 0, fmt.Errorf("unsupported CMAC block size: %d", blockSize)
	}
}

func xorBytes(dst, a, b []byte) {
	for i := 0; i < len(dst); i++ {
		dst[i] = a[i] ^ b[i]
	}
}
