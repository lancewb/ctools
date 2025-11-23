package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/smx509"
	"github.com/google/uuid"
)

func (c *CryptoService) ListCertificates() []CertRecord {
	return c.readCerts()
}

func (c *CryptoService) DeleteCertificate(id string) []CertRecord {
	certs := c.readCerts()
	out := make([]CertRecord, 0, len(certs))
	for _, crt := range certs {
		if crt.ID == id {
			continue
		}
		out = append(out, crt)
	}
	c.writeCerts(out)
	return out
}

func (c *CryptoService) ParseCertificate(req CertParseRequest) (CertParseResult, error) {
	block, _ := pem.Decode([]byte(req.PEM))
	if block == nil {
		return CertParseResult{}, errors.New("invalid certificate PEM")
	}
	if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
		return buildCertResult(cert.Subject, cert.Issuer, cert.SerialNumber, cert.NotBefore, cert.NotAfter, cert.DNSNames, cert.EmailAddresses, cert.IPAddresses, cert.URIs, cert.KeyUsage, cert.ExtKeyUsage, cert.Raw, cert.SignatureAlgorithm, cert.PublicKeyAlgorithm), nil
	}
	if cert, err := smx509.ParseCertificate(block.Bytes); err == nil {
		return buildCertResult(cert.Subject, cert.Issuer, cert.SerialNumber, cert.NotBefore, cert.NotAfter, cert.DNSNames, cert.EmailAddresses, cert.IPAddresses, cert.URIs, cert.KeyUsage, cert.ExtKeyUsage, cert.Raw, cert.SignatureAlgorithm, cert.PublicKeyAlgorithm), nil
	}
	return CertParseResult{}, errors.New("unable to parse certificate contents")
}

func (c *CryptoService) IssueCertificate(req CertIssueRequest) (CertIssueResult, error) {
	switch strings.ToLower(req.Algorithm) {
	case "rsa":
		return c.issueRSACertificate(req)
	case "sm2":
		return c.issueSM2Certificate(req)
	default:
		return CertIssueResult{}, fmt.Errorf("unsupported algorithm: %s", req.Algorithm)
	}
}

func (c *CryptoService) issueRSACertificate(req CertIssueRequest) (CertIssueResult, error) {
	rootKey, rootCert, rootRecord, err := c.ensureRSARoot()
	if err != nil {
		return CertIssueResult{}, err
	}
	validDays := req.ValidDays
	if validDays == 0 {
		validDays = 365
	}
	leafKey, err := rsa.GenerateKey(rand.Reader, chooseRSABits(req.KeySize))
	if err != nil {
		return CertIssueResult{}, err
	}
	serial, _ := rand.Int(rand.Reader, big.NewInt(0).Lsh(big.NewInt(1), 128))
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   fallbackCommonName(req.CommonName, "RSA Service"),
			Organization: []string{"Wails Toolbox"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(time.Duration(validDays) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           determineExtUsage(req.Usage),
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, rootCert, &leafKey.PublicKey, rootKey)
	if err != nil {
		return CertIssueResult{}, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: must(x509.MarshalPKIXPublicKey(&leafKey.PublicKey))})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(leafKey)})

	storedKey := c.saveKey(StoredKey{
		ID:         uuidString(),
		Name:       fmt.Sprintf("%s-key", req.CommonName),
		Algorithm:  "RSA",
		KeyType:    "private",
		Format:     "generated",
		Usage:      []string{"leaf"},
		PrivatePEM: string(keyPEM),
		PublicPEM:  string(pubPEM),
		CreatedAt:  time.Now(),
	})

	record := c.appendCertificate(CertRecord{
		ID:        uuidString(),
		Name:      req.CommonName,
		Algorithm: "RSA",
		Usage:     strings.ToLower(req.Usage),
		CertPEM:   string(certPEM),
		KeyID:     storedKey.ID,
		Serial:    template.SerialNumber.String(),
		NotBefore: template.NotBefore.Format(time.RFC3339),
		NotAfter:  template.NotAfter.Format(time.RFC3339),
		Subject:   map[string]string{"CN": template.Subject.CommonName},
		Issuer:    map[string]string{"CN": rootCert.Subject.CommonName},
		CreatedAt: time.Now(),
	})

	result := CertIssueResult{
		Keys:         []*StoredKey{&storedKey},
		Certificates: []CertRecord{record},
	}
	if rootRecord != nil {
		result.RootCA = rootRecord
	}
	return result, nil
}

func (c *CryptoService) issueSM2Certificate(req CertIssueRequest) (CertIssueResult, error) {
	rootKey, rootCert, rootRecord, err := c.ensureSM2Root()
	if err != nil {
		return CertIssueResult{}, err
	}
	validDays := req.ValidDays
	if validDays == 0 {
		validDays = 365
	}
	// Signing cert
	signKey, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return CertIssueResult{}, err
	}
	encKey, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return CertIssueResult{}, err
	}
	signTemplate := smx509.Certificate{
		SerialNumber: randomSerial(),
		Subject: pkix.Name{
			CommonName:   fallbackCommonName(req.CommonName, "SM2 Sign"),
			Organization: []string{"Wails Toolbox"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(time.Duration(validDays) * 24 * time.Hour),
		KeyUsage:              smx509.KeyUsageDigitalSignature,
		ExtKeyUsage:           determineExtUsage(req.Usage),
		BasicConstraintsValid: true,
	}
	encTemplate := signTemplate
	encTemplate.SerialNumber = randomSerial()
	encTemplate.Subject.CommonName = fallbackCommonName(req.CommonName, "SM2 Encrypt")
	encTemplate.KeyUsage = smx509.KeyUsageKeyAgreement | smx509.KeyUsageKeyEncipherment | smx509.KeyUsageDataEncipherment

	signDER, err := smx509.CreateCertificate(rand.Reader, &signTemplate, rootCert, &signKey.PublicKey, rootKey)
	if err != nil {
		return CertIssueResult{}, err
	}
	encDER, err := smx509.CreateCertificate(rand.Reader, &encTemplate, rootCert, &encKey.PublicKey, rootKey)
	if err != nil {
		return CertIssueResult{}, err
	}

	signPrivPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: must(smx509.MarshalSM2PrivateKey(signKey))})
	signPubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: must(smx509.MarshalPKIXPublicKey(&signKey.PublicKey))})
	encPrivPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: must(smx509.MarshalSM2PrivateKey(encKey))})
	encPubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: must(smx509.MarshalPKIXPublicKey(&encKey.PublicKey))})

	signStored := c.saveKey(StoredKey{
		ID:         uuidString(),
		Name:       fmt.Sprintf("%s-sign", req.CommonName),
		Algorithm:  "SM2",
		KeyType:    "private",
		Format:     "generated",
		Usage:      []string{"leaf"},
		PrivatePEM: string(signPrivPEM),
		PublicPEM:  string(signPubPEM),
		CreatedAt:  time.Now(),
		Extra:      map[string]string{"variant": "sign"},
	})
	encStored := c.saveKey(StoredKey{
		ID:         uuidString(),
		Name:       fmt.Sprintf("%s-enc", req.CommonName),
		Algorithm:  "SM2",
		KeyType:    "private",
		Format:     "generated",
		Usage:      []string{"leaf"},
		PrivatePEM: string(encPrivPEM),
		PublicPEM:  string(encPubPEM),
		CreatedAt:  time.Now(),
		Extra:      map[string]string{"variant": "encrypt"},
	})

	signRecord := c.appendCertificate(CertRecord{
		ID:        uuidString(),
		Name:      fmt.Sprintf("%s (签名)", req.CommonName),
		Algorithm: "SM2",
		Usage:     "sign",
		CertPEM:   string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: signDER})),
		KeyID:     signStored.ID,
		Serial:    signTemplate.SerialNumber.String(),
		NotBefore: signTemplate.NotBefore.Format(time.RFC3339),
		NotAfter:  signTemplate.NotAfter.Format(time.RFC3339),
		Subject:   map[string]string{"CN": signTemplate.Subject.CommonName},
		Issuer:    map[string]string{"CN": rootCert.Subject.CommonName},
		CreatedAt: time.Now(),
	})

	encRecord := c.appendCertificate(CertRecord{
		ID:        uuidString(),
		Name:      fmt.Sprintf("%s (加密)", req.CommonName),
		Algorithm: "SM2",
		Usage:     "encrypt",
		CertPEM:   string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: encDER})),
		KeyID:     encStored.ID,
		Serial:    encTemplate.SerialNumber.String(),
		NotBefore: encTemplate.NotBefore.Format(time.RFC3339),
		NotAfter:  encTemplate.NotAfter.Format(time.RFC3339),
		Subject:   map[string]string{"CN": encTemplate.Subject.CommonName},
		Issuer:    map[string]string{"CN": rootCert.Subject.CommonName},
		CreatedAt: time.Now(),
	})

	result := CertIssueResult{
		Keys:         []*StoredKey{&signStored, &encStored},
		Certificates: []CertRecord{signRecord, encRecord},
	}
	if rootRecord != nil {
		result.RootCA = rootRecord
	}
	return result, nil
}

func (c *CryptoService) ensureRSARoot() (*rsa.PrivateKey, *x509.Certificate, *CertRecord, error) {
	var rootKey *StoredKey
	for _, k := range c.readKeys() {
		if strings.EqualFold(k.Algorithm, "RSA") && containsUsage(k.Usage, "ca") {
			copy := k
			rootKey = &copy
			break
		}
	}
	var rootCert *CertRecord
	for _, cr := range c.readCerts() {
		if strings.EqualFold(cr.Algorithm, "RSA") && cr.Usage == "root-ca" {
			copy := cr
			rootCert = &copy
			break
		}
	}
	if rootKey != nil && rootCert != nil {
		priv, err := parseRSAPrivate(rootKey.PrivatePEM)
		if err != nil {
			return nil, nil, nil, err
		}
		cert, err := x509.ParseCertificate(decodePEMBytes(rootCert.CertPEM))
		if err != nil {
			return nil, nil, nil, err
		}
		return priv, cert, nil, nil
	}
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, nil, err
	}
	template := &x509.Certificate{
		SerialNumber: randomSerial(),
		Subject: pkix.Name{
			CommonName:   "Wails Toolbox RSA Root CA",
			Organization: []string{"Wails Toolbox"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, nil, err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: must(x509.MarshalPKIXPublicKey(&priv.PublicKey))})
	stored := c.saveKey(StoredKey{
		ID:         uuidString(),
		Name:       "RSA Root",
		Algorithm:  "RSA",
		KeyType:    "private",
		Format:     "generated",
		Usage:      []string{"ca"},
		PrivatePEM: string(privPEM),
		PublicPEM:  string(pubPEM),
		CreatedAt:  time.Now(),
	})
	record := CertRecord{
		ID:        uuidString(),
		Name:      "RSA Root CA",
		Algorithm: "RSA",
		Usage:     "root-ca",
		CertPEM:   string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})),
		KeyID:     stored.ID,
		Serial:    template.SerialNumber.String(),
		NotBefore: template.NotBefore.Format(time.RFC3339),
		NotAfter:  template.NotAfter.Format(time.RFC3339),
		Subject:   map[string]string{"CN": template.Subject.CommonName},
		Issuer:    map[string]string{"CN": template.Subject.CommonName},
		CreatedAt: time.Now(),
	}
	record = c.appendCertificate(record)
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, nil, err
	}
	return priv, parsed, &record, nil
}

func (c *CryptoService) ensureSM2Root() (*sm2.PrivateKey, *smx509.Certificate, *CertRecord, error) {
	var rootKey *StoredKey
	for _, k := range c.readKeys() {
		if strings.EqualFold(k.Algorithm, "SM2") && containsUsage(k.Usage, "ca") {
			copy := k
			rootKey = &copy
			break
		}
	}
	var rootCert *CertRecord
	for _, cr := range c.readCerts() {
		if strings.EqualFold(cr.Algorithm, "SM2") && cr.Usage == "root-ca" {
			copy := cr
			rootCert = &copy
			break
		}
	}
	if rootKey != nil && rootCert != nil {
		priv, err := parseSM2Private(rootKey.PrivatePEM)
		if err != nil {
			return nil, nil, nil, err
		}
		cert, err := smx509.ParseCertificate(decodePEMBytes(rootCert.CertPEM))
		if err != nil {
			return nil, nil, nil, err
		}
		return priv, cert, nil, nil
	}
	priv, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}
	template := &smx509.Certificate{
		SerialNumber: randomSerial(),
		Subject: pkix.Name{
			CommonName:   "Wails Toolbox SM2 Root CA",
			Organization: []string{"Wails Toolbox"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              smx509.KeyUsageCertSign | smx509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLen:            2,
	}
	der, err := smx509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, nil, err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: must(smx509.MarshalSM2PrivateKey(priv))})
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: must(smx509.MarshalPKIXPublicKey(&priv.PublicKey))})
	stored := c.saveKey(StoredKey{
		ID:         uuidString(),
		Name:       "SM2 Root",
		Algorithm:  "SM2",
		KeyType:    "private",
		Format:     "generated",
		Usage:      []string{"ca"},
		PrivatePEM: string(privPEM),
		PublicPEM:  string(pubPEM),
		CreatedAt:  time.Now(),
		Extra:      map[string]string{"variant": "sign"},
	})
	record := CertRecord{
		ID:        uuidString(),
		Name:      "SM2 Root CA",
		Algorithm: "SM2",
		Usage:     "root-ca",
		CertPEM:   string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})),
		KeyID:     stored.ID,
		Serial:    template.SerialNumber.String(),
		NotBefore: template.NotBefore.Format(time.RFC3339),
		NotAfter:  template.NotAfter.Format(time.RFC3339),
		Subject:   map[string]string{"CN": template.Subject.CommonName},
		Issuer:    map[string]string{"CN": template.Subject.CommonName},
		CreatedAt: time.Now(),
	}
	record = c.appendCertificate(record)
	parsed, err := smx509.ParseCertificate(der)
	if err != nil {
		return nil, nil, nil, err
	}
	return priv, parsed, &record, nil
}

func (c *CryptoService) appendCertificate(record CertRecord) CertRecord {
	certs := c.readCerts()
	certs = append(certs, record)
	c.writeCerts(certs)
	return record
}

func buildCertResult(subject pkix.Name, issuer pkix.Name, serial *big.Int, notBefore, notAfter time.Time, dns, emails []string, ips []net.IP, uris []*url.URL, keyUsage x509.KeyUsage, ext []x509.ExtKeyUsage, raw []byte, sigAlg x509.SignatureAlgorithm, pubAlg x509.PublicKeyAlgorithm) CertParseResult {
	result := CertParseResult{
		Subject:            nameToMap(subject),
		Issuer:             nameToMap(issuer),
		Serial:             serial.String(),
		NotBefore:          notBefore.Format(time.RFC3339),
		NotAfter:           notAfter.Format(time.RFC3339),
		DNSNames:           append([]string{}, dns...),
		KeyUsage:           keyUsageStrings(keyUsage),
		ExtKeyUsage:        extKeyUsageStrings(ext),
		SignatureAlgorithm: sigAlg.String(),
		PublicKeyAlgorithm: pubAlg.String(),
		RawHex:             strings.ToUpper(hex.EncodeToString(raw)),
	}
	for _, ip := range ips {
		result.IPAddresses = append(result.IPAddresses, ip.String())
	}
	for _, uri := range uris {
		result.SANs = append(result.SANs, uri.String())
	}
	result.SANs = append(result.SANs, emails...)
	return result
}

func determineExtUsage(usage string) []x509.ExtKeyUsage {
	switch strings.ToLower(usage) {
	case "server":
		return []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	case "client":
		return []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	default:
		return []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	}
}

func fallbackCommonName(name, fallback string) string {
	if strings.TrimSpace(name) != "" {
		return name
	}
	return fallback
}

func containsUsage(list []string, target string) bool {
	for _, item := range list {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
}

func chooseRSABits(bits int) int {
	if bits == 0 {
		return 2048
	}
	if bits < 2048 {
		return 2048
	}
	return bits
}

func decodePEMBytes(pemString string) []byte {
	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return nil
	}
	return block.Bytes
}

func randomSerial() *big.Int {
	serial, _ := rand.Int(rand.Reader, big.NewInt(0).Lsh(big.NewInt(1), 128))
	return serial
}

func must(data []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return data
}

func uuidString() string {
	return uuid.New().String()
}

func nameToMap(name pkix.Name) map[string]string {
	result := map[string]string{}
	if name.CommonName != "" {
		result["CN"] = name.CommonName
	}
	if len(name.Organization) > 0 {
		result["O"] = strings.Join(name.Organization, ",")
	}
	if len(name.Country) > 0 {
		result["C"] = strings.Join(name.Country, ",")
	}
	if len(name.Province) > 0 {
		result["ST"] = strings.Join(name.Province, ",")
	}
	if len(name.Locality) > 0 {
		result["L"] = strings.Join(name.Locality, ",")
	}
	return result
}

func keyUsageStrings(usage x509.KeyUsage) []string {
	var out []string
	if usage&x509.KeyUsageDigitalSignature != 0 {
		out = append(out, "DigitalSignature")
	}
	if usage&x509.KeyUsageContentCommitment != 0 {
		out = append(out, "ContentCommitment")
	}
	if usage&x509.KeyUsageKeyEncipherment != 0 {
		out = append(out, "KeyEncipherment")
	}
	if usage&x509.KeyUsageDataEncipherment != 0 {
		out = append(out, "DataEncipherment")
	}
	if usage&x509.KeyUsageKeyAgreement != 0 {
		out = append(out, "KeyAgreement")
	}
	if usage&x509.KeyUsageCertSign != 0 {
		out = append(out, "CertSign")
	}
	if usage&x509.KeyUsageCRLSign != 0 {
		out = append(out, "CRLSign")
	}
	return out
}

func extKeyUsageStrings(usages []x509.ExtKeyUsage) []string {
	out := make([]string, 0, len(usages))
	for _, u := range usages {
		switch u {
		case x509.ExtKeyUsageServerAuth:
			out = append(out, "ServerAuth")
		case x509.ExtKeyUsageClientAuth:
			out = append(out, "ClientAuth")
		case x509.ExtKeyUsageCodeSigning:
			out = append(out, "CodeSigning")
		case x509.ExtKeyUsageEmailProtection:
			out = append(out, "EmailProtection")
		case x509.ExtKeyUsageTimeStamping:
			out = append(out, "TimeStamping")
		default:
			out = append(out, fmt.Sprintf("Ext(%d)", u))
		}
	}
	return out
}
