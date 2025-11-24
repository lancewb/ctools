package crypto

import (
	"crypto/elliptic"
	"fmt"
	"strings"

	"github.com/ProtonMail/go-crypto/bitcurves"
	"github.com/ProtonMail/go-crypto/brainpool"
)

type eccCurveInfo struct {
	Identifier string
	Display    string
	Family     string
	Curve      elliptic.Curve
}

var eccCurveRegistry = map[string]eccCurveInfo{}

func init() {
	registerECCCurve(eccCurveInfo{
		Identifier: "nist-p224",
		Display:    "NIST P-224",
		Family:     "NIST P",
		Curve:      elliptic.P224(),
	}, "p-224", "p224", "secp224r1", "nistp224")

	registerECCCurve(eccCurveInfo{
		Identifier: "nist-p256",
		Display:    "NIST P-256",
		Family:     "NIST P",
		Curve:      elliptic.P256(),
	}, "p-256", "p256", "prime256v1", "secp256r1", "nistp256")

	registerECCCurve(eccCurveInfo{
		Identifier: "nist-p384",
		Display:    "NIST P-384",
		Family:     "NIST P",
		Curve:      elliptic.P384(),
	}, "p-384", "p384", "secp384r1", "nistp384")

	registerECCCurve(eccCurveInfo{
		Identifier: "nist-p521",
		Display:    "NIST P-521",
		Family:     "NIST P",
		Curve:      elliptic.P521(),
	}, "p-521", "p521", "secp521r1", "nistp521")

	registerECCCurve(eccCurveInfo{
		Identifier: "secp256k1",
		Display:    "SECG secp256k1",
		Family:     "SECG",
		Curve:      bitcurves.S256(),
	}, "k-256", "nistk256")

	registerECCCurve(eccCurveInfo{
		Identifier: "secp224k1",
		Display:    "SECG secp224k1",
		Family:     "SECG",
		Curve:      bitcurves.S224(),
	}, "k-224", "nistk224")

	registerECCCurve(eccCurveInfo{
		Identifier: "secp192k1",
		Display:    "SECG secp192k1",
		Family:     "SECG",
		Curve:      bitcurves.S192(),
	}, "k-192", "nistk192")

	registerECCCurve(eccCurveInfo{
		Identifier: "brainpool-p256r1",
		Display:    "Brainpool P256r1",
		Family:     "Brainpool",
		Curve:      brainpool.P256r1(),
	}, "brainpoolp256r1")

	registerECCCurve(eccCurveInfo{
		Identifier: "brainpool-p384r1",
		Display:    "Brainpool P384r1",
		Family:     "Brainpool",
		Curve:      brainpool.P384r1(),
	}, "brainpoolp384r1")

	registerECCCurve(eccCurveInfo{
		Identifier: "brainpool-p512r1",
		Display:    "Brainpool P512r1",
		Family:     "Brainpool",
		Curve:      brainpool.P512r1(),
	}, "brainpoolp512r1")

	registerECCCurve(eccCurveInfo{
		Identifier: "brainpool-p256t1",
		Display:    "Brainpool P256t1",
		Family:     "Brainpool",
		Curve:      brainpool.P256t1(),
	}, "brainpoolp256t1")

	registerECCCurve(eccCurveInfo{
		Identifier: "brainpool-p384t1",
		Display:    "Brainpool P384t1",
		Family:     "Brainpool",
		Curve:      brainpool.P384t1(),
	}, "brainpoolp384t1")

	registerECCCurve(eccCurveInfo{
		Identifier: "brainpool-p512t1",
		Display:    "Brainpool P512t1",
		Family:     "Brainpool",
		Curve:      brainpool.P512t1(),
	}, "brainpoolp512t1")

	// X9.62 aliases that map to NIST curves.
	registerECCAlias("prime256v1", "nist-p256")
	registerECCAlias("ansi-x9.62-prime256v1", "nist-p256")
}

func registerECCCurve(info eccCurveInfo, aliases ...string) {
	key := normalizeCurveName(info.Identifier)
	info.Identifier = key
	if info.Curve == nil {
		panic(fmt.Sprintf("missing curve instance for %s", key))
	}
	names := append([]string{info.Identifier}, aliases...)
	if paramsName := strings.TrimSpace(info.Curve.Params().Name); paramsName != "" {
		names = append(names, paramsName)
	}
	for _, name := range names {
		norm := normalizeCurveName(name)
		if norm == "" {
			continue
		}
		eccCurveRegistry[norm] = info
	}
}

func registerECCAlias(alias, target string) {
	info, ok := eccCurveRegistry[normalizeCurveName(target)]
	if !ok {
		return
	}
	eccCurveRegistry[normalizeCurveName(alias)] = info
}

func normalizeCurveName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, "-", "")
	return name
}

func resolveECCurve(name string) (eccCurveInfo, error) {
	norm := normalizeCurveName(name)
	if norm == "" {
		norm = "nistp256"
	}
	if info, ok := eccCurveRegistry[norm]; ok {
		return info, nil
	}
	if info, ok := eccCurveRegistry["nistp256"]; ok {
		return info, nil
	}
	return eccCurveInfo{}, fmt.Errorf("unsupported curve: %s", name)
}

func describeCurveByParamsName(paramsName string) (eccCurveInfo, bool) {
	if paramsName == "" {
		return eccCurveInfo{}, false
	}
	info, ok := eccCurveRegistry[normalizeCurveName(paramsName)]
	return info, ok
}
