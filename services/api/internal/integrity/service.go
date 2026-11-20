package integrity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"radarkm/api/internal/db"
)

type Deps struct {
	Decoder     GoogleIntegrityDecoder
	Devices     *db.DeviceRepository
	Events      *db.IntegrityEventRepository
	Signer      *Ed25519Signer
	Attestation *KeyAttestationVerifier
	Clock       Clock
	PackageName string
}

// Service evaluates Play Integrity + Device Recall + Key Attestation and persists outcomes.
type Service struct {
	decoder     GoogleIntegrityDecoder
	devices     *db.DeviceRepository
	events      *db.IntegrityEventRepository
	signer      *Ed25519Signer
	attestation *KeyAttestationVerifier
	clock       Clock
	packageName string
}

func NewService(deps Deps) *Service {
	clk := deps.Clock
	if clk == nil {
		clk = realClock{}
	}
	pkg := deps.PackageName
	if pkg == "" {
		pkg = ExpectedPackageName
	}
	return &Service{
		decoder:     deps.Decoder,
		devices:     deps.Devices,
		events:      deps.Events,
		signer:      deps.Signer,
		attestation: deps.Attestation,
		clock:       clk,
		packageName: pkg,
	}
}

func (s *Service) Verify(ctx context.Context, req VerifyRequest) (VerifyResult, error) {
	if s.decoder == nil || s.devices == nil || s.events == nil || s.attestation == nil {
		return VerifyResult{}, fmt.Errorf("integrity service missing dependencies")
	}
	if req.DeviceID == "" || req.WidevineHash == "" || req.IntegrityToken == "" {
		return VerifyResult{}, fmt.Errorf("device_id, widevine_hash and integrity token are required")
	}

	now := s.clock.Now().UTC()
	payload, err := s.decoder.Decode(ctx, req.IntegrityToken)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("decode integrity token: %w", err)
	}

	if reason := validatePlayIntegrity(payload, s.packageName, req.ExpectedRequestHash, req.ExpectedNonce, now); reason != "" {
		return s.persist(ctx, req, payload, VerifyResult{
			Verdict:      VerdictRejected,
			RejectReason: reason,
		}, now)
	}

	if err := s.attestation.Verify(req.AttestationCertDERs, req.AttestationChallenge); err != nil {
		return s.persist(ctx, req, payload, VerifyResult{
			Verdict:          VerdictRejected,
			RejectReason:     RejectAttestationFailed,
			PlayIntegrityOK:  true,
			KeyAttestationOK: false,
		}, now)
	}

	bits := EncodeDeviceRecallBits(payload.DeviceRecall)
	result := VerifyResult{
		PlayIntegrityOK:     true,
		KeyAttestationOK:    true,
		DeviceRecallBits:    bits,
		FreeQuotaSuppressed: AnyAbuseBitActive(bits),
	}
	switch {
	case AnyAbuseBitActive(bits) && !req.HasPaidEntitlement:
		result.Verdict = VerdictFreeQuotaSuppressed
	default:
		result.Verdict = VerdictApproved
	}
	return s.persist(ctx, req, payload, result, now)
}

func (s *Service) ResetDeviceRecallBits(ctx context.Context, deviceID string) (db.Device, error) {
	if s.devices == nil {
		return db.Device{}, fmt.Errorf("devices repository required")
	}
	return s.devices.ResetDeviceRecallBits(ctx, deviceID)
}

func (s *Service) persist(
	ctx context.Context,
	req VerifyRequest,
	payload IntegrityTokenPayload,
	result VerifyResult,
	now time.Time,
) (VerifyResult, error) {
	claims, _ := json.Marshal(map[string]any{
		"reject_reason":         result.RejectReason,
		"free_quota_suppressed": result.FreeQuotaSuppressed,
		"device_recall_bits":    result.DeviceRecallBits,
		"play_integrity_ok":     result.PlayIntegrityOK,
		"key_attestation_ok":    result.KeyAttestationOK,
		"token":                 payload,
	})
	deviceID := req.DeviceID
	event, err := s.events.Insert(ctx, db.IntegrityEvent{
		DeviceID:  &deviceID,
		Verdict:   result.Verdict,
		RawClaims: claims,
	})
	if err != nil {
		return VerifyResult{}, fmt.Errorf("insert integrity_events: %w", err)
	}
	device, err := s.devices.UpdateIntegrityState(
		ctx,
		req.DeviceID,
		req.WidevineHash,
		result.Verdict,
		result.DeviceRecallBits,
		now,
	)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("update devices: %w", err)
	}
	result.Device = device
	result.Event = event
	return result, nil
}
