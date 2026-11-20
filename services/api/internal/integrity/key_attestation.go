package integrity

import (
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
)

// Android Key Attestation extension OID: 1.3.6.1.4.1.11129.2.1.17
var androidKeyAttestationOID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 1, 17}

const (
	VerifiedBootVerified = 0
)

// keyDescriptionASN1 is a parseable subset of the Android KeyDescription
// carried in extension 1.3.6.1.4.1.11129.2.1.17 (challenge + RootOfTrust fields).
type keyDescriptionASN1 struct {
	AttestationVersion       int
	AttestationSecurityLevel asn1.Enumerated
	KeymasterVersion         int
	KeymasterSecurityLevel   asn1.Enumerated
	AttestationChallenge     []byte
	UniqueID                 []byte
	DeviceLocked             bool
	VerifiedBootState        asn1.Enumerated
}

type KeyAttestationVerifier struct {
	Roots *x509.CertPool
}

func NewKeyAttestationVerifier(rootPEMs []byte) (*KeyAttestationVerifier, error) {
	pool := x509.NewCertPool()
	if len(rootPEMs) > 0 {
		if !pool.AppendCertsFromPEM(rootPEMs) {
			return nil, fmt.Errorf("no valid root certificates in PEM")
		}
	}
	return &KeyAttestationVerifier{Roots: pool}, nil
}

func (v *KeyAttestationVerifier) Verify(certDERs [][]byte, expectedChallenge []byte) error {
	if len(certDERs) == 0 {
		return fmt.Errorf("empty attestation chain")
	}
	certs := make([]*x509.Certificate, 0, len(certDERs))
	for i, der := range certDERs {
		c, err := x509.ParseCertificate(der)
		if err != nil {
			return fmt.Errorf("parse cert[%d]: %w", i, err)
		}
		certs = append(certs, c)
	}
	leaf := certs[0]
	intermediates := x509.NewCertPool()
	for _, c := range certs[1:] {
		intermediates.AddCert(c)
	}
	opts := x509.VerifyOptions{
		Roots:         v.Roots,
		Intermediates: intermediates,
		KeyUsages:    []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	if _, err := leaf.Verify(opts); err != nil {
		return fmt.Errorf("chain verify: %w", err)
	}

	extValue, ok := findAttestationExt(leaf)
	if !ok {
		return fmt.Errorf("missing android key attestation extension")
	}
	var desc keyDescriptionASN1
	if _, err := asn1.Unmarshal(extValue, &desc); err != nil {
		return fmt.Errorf("asn1 key description: %w", err)
	}
	if !bytes.Equal(desc.AttestationChallenge, expectedChallenge) {
		return fmt.Errorf("attestationChallenge mismatch")
	}
	if !desc.DeviceLocked {
		return fmt.Errorf("deviceLocked must be true")
	}
	if int(desc.VerifiedBootState) != VerifiedBootVerified {
		return fmt.Errorf("verifiedBootState must be Verified")
	}
	return nil
}

func findAttestationExt(cert *x509.Certificate) ([]byte, bool) {
	for _, e := range cert.Extensions {
		if e.Id.Equal(androidKeyAttestationOID) {
			return e.Value, true
		}
	}
	for _, e := range cert.ExtraExtensions {
		if e.Id.Equal(androidKeyAttestationOID) {
			return e.Value, true
		}
	}
	return nil, false
}

// BuildKeyDescriptionExtension marshals a KeyDescription for tests (OID applied by caller).
func BuildKeyDescriptionExtension(challenge []byte, deviceLocked bool, verifiedBootState int) ([]byte, error) {
	desc := keyDescriptionASN1{
		AttestationVersion:       200,
		AttestationSecurityLevel: 1,
		KeymasterVersion:         41,
		KeymasterSecurityLevel:   1,
		AttestationChallenge:     challenge,
		UniqueID:                 []byte("radarkm-test"),
		DeviceLocked:             deviceLocked,
		VerifiedBootState:        asn1.Enumerated(verifiedBootState),
	}
	return asn1.Marshal(desc)
}

// Ensure pkix import used when building templates in tests (package-level helper OID accessor).
func AndroidKeyAttestationOID() pkix.Extension {
	return pkix.Extension{Id: androidKeyAttestationOID}
}
