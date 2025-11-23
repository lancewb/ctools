package crypto

import (
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"strings"
)

// ParseDER parses ASN.1 DER encoded data and returns its structure.
//
// req: The DerParseRequest containing hex or base64 data.
// Returns a DerParseResult with the parsed tree of nodes, or an error.
func (c *CryptoService) ParseDER(req DerParseRequest) (DerParseResult, error) {
	var data []byte
	var err error
	if strings.TrimSpace(req.HexString) != "" {
		data, err = hex.DecodeString(strings.ReplaceAll(strings.TrimSpace(req.HexString), " ", ""))
	} else if strings.TrimSpace(req.Base64) != "" {
		data, err = decodeBlob(req.Base64, "base64")
	} else {
		return DerParseResult{}, errors.New("hexString or base64 input required")
	}
	if err != nil {
		return DerParseResult{}, err
	}
	nodes, err := parseDERTree(data)
	if err != nil {
		return DerParseResult{}, err
	}
	return DerParseResult{Nodes: nodes}, nil
}

func parseDERTree(data []byte) ([]DerNode, error) {
	nodes := []DerNode{}
	for len(data) > 0 {
		var raw asn1.RawValue
		rest, err := asn1.Unmarshal(data, &raw)
		if err != nil {
			return nil, err
		}
		node := DerNode{
			Tag:         int(raw.Tag),
			Class:       classLabel(raw.Class),
			Constructed: raw.IsCompound,
			Length:      len(raw.Bytes),
			Hex:         strings.ToUpper(hex.EncodeToString(raw.Bytes)),
		}
		if raw.IsCompound {
			children, err := parseDERTree(raw.Bytes)
			if err == nil {
				node.Children = children
			}
		}
		nodes = append(nodes, node)
		data = rest
	}
	return nodes, nil
}

func classLabel(class int) string {
	switch class {
	case 0:
		return "UNIVERSAL"
	case 1:
		return "APPLICATION"
	case 2:
		return "CONTEXT"
	case 3:
		return "PRIVATE"
	default:
		return "UNKNOWN"
	}
}
