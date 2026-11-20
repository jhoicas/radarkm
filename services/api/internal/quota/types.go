package quota

import (
	"time"

	"radarkm/api/internal/db"
)

type Decision string

const (
	DecisionAllowed        Decision = "ALLOWED"
	DecisionDedup          Decision = "DEDUP"
	DecisionQuotaExceeded  Decision = "QUOTA_EXCEEDED"
	DecisionDeviceBlocked  Decision = "DEVICE_BLOCKED"
)

type ConsumeRequest struct {
	DeviceHash     string
	Platform       string
	FareCop        int64
	DistanceMeters int64
}

type ConsumeResult struct {
	Decision          Decision
	Used              int64
	Remaining         int64
	Limit             int64
	DayBogota         string
	ResetAt           time.Time
	ActivePlans       []db.Plan
	EntitlementBypass bool
	OfferHash         string
}
