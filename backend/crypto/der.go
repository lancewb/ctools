package crypto

import (
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode/utf8"
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
			Label:       describeTag(raw),
			Constructed: raw.IsCompound,
			Length:      len(raw.Bytes),
			Hex:         strings.ToUpper(hex.EncodeToString(raw.Bytes)),
		}
		if !raw.IsCompound {
			if value := describePrimitiveValue(raw); value != "" {
				node.Value = value
			}
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

func describeTag(raw asn1.RawValue) string {
	if raw.Class == 0 {
		if name, ok := universalTagNames[int(raw.Tag)]; ok {
			return name
		}
	}
	if raw.Class == 2 {
		return fmt.Sprintf("CONTEXT [%d]", raw.Tag)
	}
	if raw.Class == 1 {
		return fmt.Sprintf("APPLICATION [%d]", raw.Tag)
	}
	if raw.Class == 3 {
		return fmt.Sprintf("PRIVATE [%d]", raw.Tag)
	}
	return fmt.Sprintf("Tag %d (%s)", raw.Tag, classLabel(raw.Class))
}

func describePrimitiveValue(raw asn1.RawValue) string {
	if raw.IsCompound {
		return ""
	}
	switch raw.Class {
	case 0:
		// Universal
	default:
		if len(raw.Bytes) == 0 {
			return ""
		}
		return fmt.Sprintf("0x%s", strings.ToUpper(hex.EncodeToString(raw.Bytes)))
	}
	switch raw.Tag {
	case 1:
		if len(raw.Bytes) == 0 {
			return ""
		}
		if raw.Bytes[0] == 0x00 {
			return "false"
		}
		return "true"
	case 2, 10:
		var bi big.Int
		if _, err := asn1.Unmarshal(raw.FullBytes, &bi); err == nil {
			return bi.String()
		}
	case 3:
		var bs asn1.BitString
		if _, err := asn1.Unmarshal(raw.FullBytes, &bs); err == nil {
			return fmt.Sprintf("bits=%d hex=%s", bs.BitLength, strings.ToUpper(hex.EncodeToString(bs.Bytes)))
		}
	case 4:
		return formatOctets(raw.Bytes)
	case 5:
		return "NULL"
	case 6:
		var oid asn1.ObjectIdentifier
		if _, err := asn1.Unmarshal(raw.FullBytes, &oid); err == nil {
			return oid.String()
		}
	case 12, 19, 20, 22, 26:
		if utf8.Valid(raw.Bytes) {
			return string(raw.Bytes)
		}
		return formatOctets(raw.Bytes)
	case 23:
		var t time.Time
		if _, err := asn1.Unmarshal(raw.FullBytes, &t); err == nil {
			return t.UTC().Format(time.RFC3339)
		}
	case 24:
		var t time.Time
		if _, err := asn1.UnmarshalWithParams(raw.FullBytes, &t, "generalized"); err == nil {
			return t.UTC().Format(time.RFC3339Nano)
		}
	case 30:
		if str, ok := decodeBMPString(raw.Bytes); ok {
			return str
		}
	}
	return ""
}

func formatOctets(data []byte) string {
	if len(data) == 0 {
		return "(empty)"
	}
	if isPrintableASCII(data) {
		return fmt.Sprintf("text=\"%s\"", string(data))
	}
	return fmt.Sprintf("hex=%s", strings.ToUpper(hex.EncodeToString(data)))
}

func isPrintableASCII(data []byte) bool {
	for _, b := range data {
		if b == '\n' || b == '\r' || b == '\t' {
			continue
		}
		if b < 0x20 || b > 0x7E {
			return false
		}
	}
	return true
}

func decodeBMPString(b []byte) (string, bool) {
	if len(b)%2 != 0 {
		return "", false
	}
	runes := make([]rune, len(b)/2)
	for i := 0; i < len(b); i += 2 {
		runes[i/2] = rune(b[i])<<8 | rune(b[i+1])
	}
	if !utf8.ValidString(string(runes)) {
		return "", false
	}
	return string(runes), true
}

var universalTagNames = map[int]string{
	1:  "BOOLEAN",
	2:  "INTEGER",
	3:  "BIT STRING",
	4:  "OCTET STRING",
	5:  "NULL",
	6:  "OBJECT IDENTIFIER",
	7:  "ObjectDescriptor",
	8:  "EXTERNAL",
	9:  "REAL",
	10: "ENUMERATED",
	11: "EMBEDDED PDV",
	12: "UTF8String",
	13: "RELATIVE-OID",
	14: "TIME",
	16: "SEQUENCE",
	17: "SET",
	18: "NumericString",
	19: "PrintableString",
	20: "TeletexString",
	21: "VideotexString",
	22: "IA5String",
	23: "UTCTime",
	24: "GeneralizedTime",
	25: "GraphicString",
	26: "VisibleString",
	27: "GeneralString",
	28: "UniversalString",
	30: "BMPString",
}
