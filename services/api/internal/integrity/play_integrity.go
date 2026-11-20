package integrity

import (
	"fmt"
	"slices"
	"time"
)

func validatePlayIntegrity(
	payload IntegrityTokenPayload,
	packageName string,
	expectedHash string,
	expectedNonce string,
	now time.Time,
) (rejectReason string) {
	if payload.RequestDetails.RequestPackageName != packageName {
		return RejectPackageMismatch
	}

	ts := time.UnixMilli(payload.RequestDetails.TimestampMillis).UTC()
	age := now.UTC().Sub(ts)
	if age < 0 {
		age = -age
	}
	if age > MaxTokenAge {
		return RejectTokenExpired
	}

	switch {
	case expectedHash != "":
		if payload.RequestDetails.RequestHash != expectedHash {
			return RejectHashMismatch
		}
	case expectedNonce != "":
		if payload.RequestDetails.Nonce != expectedNonce {
			return RejectHashMismatch
		}
	default:
		return fmt.Sprintf("%s: missing expected requestHash/nonce", RejectHashMismatch)
	}

	if payload.AppIntegrity.AppRecognitionVerdict != "PLAY_RECOGNIZED" {
		return RejectUnrecognizedApp
	}

	verdicts := payload.DeviceIntegrity.DeviceRecognitionVerdict
	if !slices.Contains(verdicts, "MEETS_DEVICE_INTEGRITY") &&
		!slices.Contains(verdicts, "MEETS_STRONG_INTEGRITY") {
		return RejectDeviceIntegrity
	}

	if payload.AccountDetails.AppLicensingVerdict != "LICENSED" {
		return RejectUnlicensed
	}
	return ""
}
