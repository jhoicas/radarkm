package integrity

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"
)

// Ed25519Signer signs canonical JSON payloads and verifies signatures (cross-checks in tests).
type Ed25519Signer struct {
	private ed25519.PrivateKey
	public  ed25519.PublicKey
	keyID   string
	clock   Clock
}

func NewEd25519SignerFromEnv(clock Clock) (*Ed25519Signer, error) {
	privPEM := os.Getenv("RADARKM_ED25519_PRIVATE_KEY_PEM")
	privB64 := os.Getenv("RADARKM_ED25519_PRIVATE_KEY_B64")
	keyID := os.Getenv("RADARKM_ED25519_KEY_ID")
	if keyID == "" {
		keyID = "default"
	}
	if privPEM == "" && privB64 == "" {
		return nil, fmt.Errorf("RADARKM_ED25519_PRIVATE_KEY_PEM or RADARKM_ED25519_PRIVATE_KEY_B64 required")
	}
	var seed []byte
	var err error
	if privPEM != "" {
		seed, err = parseEd25519PrivatePEM(privPEM)
	} else {
		seed, err = base64.StdEncoding.DecodeString(privB64)
	}
	if err != nil {
		return nil, err
	}
	return NewEd25519Signer(seed, keyID, clock)
}

func NewEd25519Signer(privateKeySeed []byte, keyID string, clock Clock) (*Ed25519Signer, error) {
	if clock == nil {
		clock = realClock{}
	}
	if keyID == "" {
		keyID = "default"
	}
	var priv ed25519.PrivateKey
	switch len(privateKeySeed) {
	case ed25519.SeedSize:
		priv = ed25519.NewKeyFromSeed(privateKeySeed)
	case ed25519.PrivateKeySize:
		priv = ed25519.PrivateKey(privateKeySeed)
	default:
		return nil, fmt.Errorf("ed25519 private key must be seed (%d) or private key (%d) bytes", ed25519.SeedSize, ed25519.PrivateKeySize)
	}
	pub := priv.Public().(ed25519.PublicKey)
	return &Ed25519Signer{private: priv, public: pub, keyID: keyID, clock: clock}, nil
}

func (s *Ed25519Signer) KeyID() string { return s.keyID }

func (s *Ed25519Signer) PublicKey() ed25519.PublicKey { return s.public }

func (s *Ed25519Signer) SignCanonical(body any, ttl time.Duration) (SignedEnvelope, []byte, error) {
	now := s.clock.Now().UTC()
	payload := CanonicalPayload{
		IssuedAt:  now,
		ExpiresAt: now.Add(ttl),
		KeyID:     s.keyID,
		Body:      body,
	}
	canonical, err := json.Marshal(payload)
	if err != nil {
		return SignedEnvelope{}, nil, err
	}
	sig := ed25519.Sign(s.private, canonical)
	return SignedEnvelope{
		PayloadB64:   base64.StdEncoding.EncodeToString(canonical),
		SignatureB64: base64.StdEncoding.EncodeToString(sig),
		KeyID:        s.keyID,
	}, canonical, nil
}

func (s *Ed25519Signer) Verify(envelope SignedEnvelope, now time.Time) error {
	canonical, err := base64.StdEncoding.DecodeString(envelope.PayloadB64)
	if err != nil {
		return fmt.Errorf("payload_b64: %w", err)
	}
	sig, err := base64.StdEncoding.DecodeString(envelope.SignatureB64)
	if err != nil {
		return fmt.Errorf("signature_b64: %w", err)
	}
	if !ed25519.Verify(s.public, canonical, sig) {
		return fmt.Errorf("invalid ed25519 signature")
	}
	var payload CanonicalPayload
	if err := json.Unmarshal(canonical, &payload); err != nil {
		return fmt.Errorf("canonical json: %w", err)
	}
	if payload.KeyID != "" && envelope.KeyID != "" && payload.KeyID != envelope.KeyID {
		return fmt.Errorf("key_id mismatch")
	}
	if now.IsZero() {
		now = s.clock.Now().UTC()
	}
	if !payload.ExpiresAt.After(now.UTC()) {
		return fmt.Errorf("payload expired")
	}
	return nil
}

func parseEd25519PrivatePEM(pemText string) ([]byte, error) {
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		// Allow raw base64 inside "PEM" env for convenience.
		trimmed := strings.TrimSpace(pemText)
		if b, err := base64.StdEncoding.DecodeString(trimmed); err == nil {
			return b, nil
		}
		return nil, fmt.Errorf("invalid PEM for ed25519 private key")
	}
	// PKCS8 parsing would go here for production PEMs; tests use seed/raw.
	return block.Bytes, nil
}
