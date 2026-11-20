package integrity

const (
	RecallBitFirst  int64 = 1 << 0
	RecallBitSecond int64 = 1 << 1
	RecallBitThird  int64 = 1 << 2
)

// EncodeDeviceRecallBits packs the three Device Recall bits into devices.device_recall_bits.
func EncodeDeviceRecallBits(payload *DeviceRecallPayload) int64 {
	if payload == nil {
		return 0
	}
	var bits int64
	if payload.BitFirst != nil && payload.BitFirst.BitValue {
		bits |= RecallBitFirst
	}
	if payload.BitSecond != nil && payload.BitSecond.BitValue {
		bits |= RecallBitSecond
	}
	if payload.BitThird != nil && payload.BitThird.BitValue {
		bits |= RecallBitThird
	}
	return bits
}

// AnyAbuseBitActive reports whether free-quota suppression should apply.
func AnyAbuseBitActive(bits int64) bool {
	return bits&(RecallBitFirst|RecallBitSecond|RecallBitThird) != 0
}
