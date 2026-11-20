package integrity

import (
	"context"
	"encoding/json"
	"fmt"
)

// GoogleIntegrityDecoder decodes/verifies a Play Integrity token (Standard or Classic).
// Production uses Google Play Integrity API; tests inject StubDecoder (no network).
type GoogleIntegrityDecoder interface {
	Decode(ctx context.Context, integrityToken string) (IntegrityTokenPayload, error)
}

// StubDecoder returns a fixed payload or error — never calls the network.
type StubDecoder struct {
	Payload IntegrityTokenPayload
	Err     error
}

func (s StubDecoder) Decode(ctx context.Context, integrityToken string) (IntegrityTokenPayload, error) {
	if s.Err != nil {
		return IntegrityTokenPayload{}, s.Err
	}
	if integrityToken == "" {
		return IntegrityTokenPayload{}, fmt.Errorf("empty integrity token")
	}
	return s.Payload, nil
}

// HTTPGoogleIntegrityDecoder is the production-shaped client.
// It does not embed secrets; credentials come from env / ADC at call time.
type HTTPGoogleIntegrityDecoder struct {
	PackageName string
	// DecodeFn allows wiring a real HTTP call without hardcoding credentials in-repo.
	DecodeFn func(ctx context.Context, integrityToken string) (IntegrityTokenPayload, error)
}

func (h HTTPGoogleIntegrityDecoder) Decode(ctx context.Context, integrityToken string) (IntegrityTokenPayload, error) {
	if h.DecodeFn == nil {
		return IntegrityTokenPayload{}, fmt.Errorf("HTTPGoogleIntegrityDecoder.DecodeFn not configured")
	}
	payload, err := h.DecodeFn(ctx, integrityToken)
	if err != nil {
		return IntegrityTokenPayload{}, err
	}
	// Defensive: ensure JSON round-trip shape is valid for persistence tests.
	if _, err := json.Marshal(payload); err != nil {
		return IntegrityTokenPayload{}, fmt.Errorf("invalid decoded payload: %w", err)
	}
	return payload, nil
}
