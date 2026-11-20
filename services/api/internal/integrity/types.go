package integrity

import (
	"time"

	"radarkm/api/internal/db"
)

const (
	ExpectedPackageName = "com.radarkm.app"
	MaxTokenAge         = 5 * time.Minute

	VerdictApproved            = "APPROVED"
	VerdictRejected            = "REJECTED"
	VerdictFreeQuotaSuppressed = "FREE_QUOTA_SUPPRESSED"

	RejectPackageMismatch   = "PACKAGE_MISMATCH"
	RejectUnrecognizedApp   = "UNRECOGNIZED_VERSION"
	RejectDeviceIntegrity   = "DEVICE_INTEGRITY_FAILED"
	RejectUnlicensed        = "UNLICENSED"
	RejectTokenExpired      = "TOKEN_EXPIRED"
	RejectHashMismatch      = "REQUEST_HASH_MISMATCH"
	RejectAttestationFailed = "KEY_ATTESTATION_FAILED"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

// IntegrityTokenPayload is the decoded Play Integrity token body (Standard or Classic).
type IntegrityTokenPayload struct {
	RequestDetails struct {
		RequestPackageName string `json:"requestPackageName"`
		RequestHash        string `json:"requestHash"`
		Nonce              string `json:"nonce"`
		TimestampMillis    int64  `json:"timestampMillis"`
	} `json:"requestDetails"`
	AppIntegrity struct {
		AppRecognitionVerdict string `json:"appRecognitionVerdict"`
	} `json:"appIntegrity"`
	DeviceIntegrity struct {
		DeviceRecognitionVerdict []string `json:"deviceRecognitionVerdict"`
	} `json:"deviceIntegrity"`
	AccountDetails struct {
		AppLicensingVerdict string `json:"appLicensingVerdict"`
	} `json:"accountDetails"`
	DeviceRecall *DeviceRecallPayload `json:"deviceRecall,omitempty"`
}

type DeviceRecallPayload struct {
	BitFirst   *DeviceRecallBit `json:"bitFirst,omitempty"`
	BitSecond  *DeviceRecallBit `json:"bitSecond,omitempty"`
	BitThird   *DeviceRecallBit `json:"bitThird,omitempty"`
	WriteDates *struct {
		BitFirst  *string `json:"bitFirst,omitempty"`
		BitSecond *string `json:"bitSecond,omitempty"`
		BitThird  *string `json:"bitThird,omitempty"`
	} `json:"writeDates,omitempty"`
}

type DeviceRecallBit struct {
	BitValue bool `json:"bitValue"`
}

type VerifyRequest struct {
	DeviceID              string
	WidevineHash          string
	IntegrityToken        string
	ExpectedRequestHash   string // Standard API
	ExpectedNonce         string // Classic API
	AttestationCertDERs   [][]byte
	AttestationChallenge  []byte
	HasPaidEntitlement    bool
}

type VerifyResult struct {
	Verdict              string
	RejectReason         string
	FreeQuotaSuppressed  bool
	DeviceRecallBits     int64
	PlayIntegrityOK      bool
	KeyAttestationOK     bool
	Device               db.Device
	Event                db.IntegrityEvent
}

type SignedEnvelope struct {
	PayloadB64   string `json:"payload_b64"`
	SignatureB64 string `json:"signature_b64"`
	KeyID        string `json:"key_id"`
}

type CanonicalPayload struct {
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	KeyID     string    `json:"key_id"`
	Body      any       `json:"body"`
}
