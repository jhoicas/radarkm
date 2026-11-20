package integrity_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"sync"
	"testing"
	"time"

	"radarkm/api/internal/db"
	"radarkm/api/internal/integrity"
	"radarkm/api/internal/migrate"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = t
}

type integrityHarness struct {
	Service   *integrity.Service
	Devices   *db.DeviceRepository
	Events    *db.IntegrityEventRepository
	Clock     *fakeClock
	Signer    *integrity.Ed25519Signer
	LeafDER   []byte
	Challenge []byte
}

func setupIntegrity(t *testing.T) integrityHarness {
	t.Helper()
	ctx := context.Background()
	pool, err := db.NewPool(ctx, db.DatabaseURL())
	if err != nil {
		t.Fatalf("postgres: %v", err)
	}
	t.Cleanup(pool.Close)
	lock, err := migrate.AcquireTestLock(ctx, pool)
	if err != nil {
		t.Fatalf("schema lock: %v", err)
	}
	t.Cleanup(func() { lock.Release(context.Background()) })
	if err := migrate.ResetLocked(ctx, pool, lock.Conn()); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	clock := &fakeClock{now: time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)}
	seed := sha256.Sum256([]byte("radarkm-test-ed25519-seed"))
	signer, err := integrity.NewEd25519Signer(seed[:], "test-key", clock)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}

	rootPEM, leafDER, challenge := mustAttestationChain(t, true, integrity.VerifiedBootVerified, []byte("challenge-ok"))
	attestation, err := integrity.NewKeyAttestationVerifier(rootPEM)
	if err != nil {
		t.Fatalf("attestation verifier: %v", err)
	}

	devices := db.NewDeviceRepository(pool)
	events := db.NewIntegrityEventRepository(pool)

	validPayload := validToken(clock.Now(), "hash-ok", "")
	svc := integrity.NewService(integrity.Deps{
		Decoder:     integrity.StubDecoder{Payload: validPayload},
		Devices:     devices,
		Events:      events,
		Signer:      signer,
		Attestation: attestation,
		Clock:       clock,
		PackageName: integrity.ExpectedPackageName,
	})
	return integrityHarness{
		Service: svc, Devices: devices, Events: events, Clock: clock,
		Signer: signer, LeafDER: leafDER, Challenge: challenge,
	}
}

func validToken(now time.Time, hash, nonce string) integrity.IntegrityTokenPayload {
	var p integrity.IntegrityTokenPayload
	p.RequestDetails.RequestPackageName = integrity.ExpectedPackageName
	p.RequestDetails.RequestHash = hash
	p.RequestDetails.Nonce = nonce
	p.RequestDetails.TimestampMillis = now.UnixMilli()
	p.AppIntegrity.AppRecognitionVerdict = "PLAY_RECOGNIZED"
	p.DeviceIntegrity.DeviceRecognitionVerdict = []string{"MEETS_DEVICE_INTEGRITY"}
	p.AccountDetails.AppLicensingVerdict = "LICENSED"
	return p
}

func seedDevice(t *testing.T, devices *db.DeviceRepository, hash string) db.Device {
	t.Helper()
	d, err := devices.Insert(context.Background(), db.Device{WidevineHash: hash, Status: "active"})
	if err != nil {
		t.Fatalf("device: %v", err)
	}
	return d
}

func TestVerify_validPlayIntegrity_persistsDeviceAndEvent(t *testing.T) {
	h := setupIntegrity(t)
	dev := seedDevice(t, h.Devices, "wv-integrity-ok")

	res, err := h.Service.Verify(context.Background(), integrity.VerifyRequest{
		DeviceID:             dev.ID,
		WidevineHash:         dev.WidevineHash,
		IntegrityToken:       "token-ok",
		ExpectedRequestHash:  "hash-ok",
		AttestationCertDERs:  [][]byte{h.LeafDER},
		AttestationChallenge: h.Challenge,
	})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if res.Verdict != integrity.VerdictApproved || !res.PlayIntegrityOK || !res.KeyAttestationOK {
		t.Fatalf("result=%+v", res)
	}
	got, err := h.Devices.GetByID(context.Background(), dev.ID)
	if err != nil {
		t.Fatalf("get device: %v", err)
	}
	if got.PlayIntegrityVerdict == nil || *got.PlayIntegrityVerdict != integrity.VerdictApproved {
		t.Fatalf("device verdict=%v", got.PlayIntegrityVerdict)
	}
	if got.LastSeenAt == nil || !got.LastSeenAt.UTC().Equal(h.Clock.Now().UTC()) {
		t.Fatalf("last_seen_at=%v want %v", got.LastSeenAt, h.Clock.Now().UTC())
	}
	list, err := h.Events.ListByDevice(context.Background(), dev.ID, 5)
	if err != nil || len(list) == 0 || list[0].Verdict != integrity.VerdictApproved {
		t.Fatalf("events=%v err=%v", list, err)
	}
}

func TestVerify_rejectsSpecificPlayIntegrityFailures(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*integrity.IntegrityTokenPayload)
		hash   string
		nonce  string
		want   string
		age    time.Duration
	}{
		{
			name: "package",
			mutate: func(p *integrity.IntegrityTokenPayload) {
				p.RequestDetails.RequestPackageName = "com.other.app"
			},
			hash: "hash-ok",
			want: integrity.RejectPackageMismatch,
		},
		{
			name: "unrecognized",
			mutate: func(p *integrity.IntegrityTokenPayload) {
				p.AppIntegrity.AppRecognitionVerdict = "UNRECOGNIZED_VERSION"
			},
			hash: "hash-ok",
			want: integrity.RejectUnrecognizedApp,
		},
		{
			name: "device",
			mutate: func(p *integrity.IntegrityTokenPayload) {
				p.DeviceIntegrity.DeviceRecognitionVerdict = []string{"MEETS_BASIC_INTEGRITY"}
			},
			hash: "hash-ok",
			want: integrity.RejectDeviceIntegrity,
		},
		{
			name: "unlicensed",
			mutate: func(p *integrity.IntegrityTokenPayload) {
				p.AccountDetails.AppLicensingVerdict = "UNLICENSED"
			},
			hash: "hash-ok",
			want: integrity.RejectUnlicensed,
		},
		{
			name: "expired",
			mutate: func(p *integrity.IntegrityTokenPayload) {
				p.RequestDetails.TimestampMillis = time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC).Add(-6 * time.Minute).UnixMilli()
			},
			hash: "hash-ok",
			want: integrity.RejectTokenExpired,
		},
		{
			name:   "hash",
			mutate: func(p *integrity.IntegrityTokenPayload) {},
			hash:   "hash-tampered",
			want:   integrity.RejectHashMismatch,
		},
		{
			name: "nonce",
			mutate: func(p *integrity.IntegrityTokenPayload) {
				p.RequestDetails.RequestHash = ""
				p.RequestDetails.Nonce = "nonce-ok"
			},
			nonce: "nonce-tampered",
			want:  integrity.RejectHashMismatch,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			pool, err := db.NewPool(ctx, db.DatabaseURL())
			if err != nil {
				t.Fatalf("postgres: %v", err)
			}
			t.Cleanup(pool.Close)
			lock, err := migrate.AcquireTestLock(ctx, pool)
			if err != nil {
				t.Fatalf("schema lock: %v", err)
			}
			t.Cleanup(func() { lock.Release(context.Background()) })
			if err := migrate.ResetLocked(ctx, pool, lock.Conn()); err != nil {
				t.Fatalf("migrate: %v", err)
			}

			clock := &fakeClock{now: time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)}
			payload := validToken(clock.Now(), "hash-ok", "")
			tc.mutate(&payload)
			seed := sha256.Sum256([]byte("seed-" + tc.name))
			signer, _ := integrity.NewEd25519Signer(seed[:], "k", clock)
			rootPEM, leafDER, challenge := mustAttestationChain(t, true, integrity.VerifiedBootVerified, []byte("challenge-ok"))
			att, _ := integrity.NewKeyAttestationVerifier(rootPEM)
			devices := db.NewDeviceRepository(pool)
			events := db.NewIntegrityEventRepository(pool)
			dev := seedDevice(t, devices, "wv-"+tc.name)
			svc := integrity.NewService(integrity.Deps{
				Decoder: integrity.StubDecoder{Payload: payload},
				Devices: devices, Events: events, Signer: signer, Attestation: att, Clock: clock,
			})
			req := integrity.VerifyRequest{
				DeviceID:             dev.ID,
				WidevineHash:         dev.WidevineHash,
				IntegrityToken:       "tok",
				ExpectedRequestHash:  tc.hash,
				ExpectedNonce:        tc.nonce,
				AttestationCertDERs:  [][]byte{leafDER},
				AttestationChallenge: challenge,
			}
			res, err := svc.Verify(ctx, req)
			if err != nil {
				t.Fatalf("err=%v", err)
			}
			if res.Verdict != integrity.VerdictRejected || res.RejectReason != tc.want {
				t.Fatalf("got verdict=%s reason=%s want %s", res.Verdict, res.RejectReason, tc.want)
			}
		})
	}
}

func TestVerify_deviceRecallSuppressesFreeQuotaButAllowsPaid(t *testing.T) {
	ctx := context.Background()
	pool, err := db.NewPool(ctx, db.DatabaseURL())
	if err != nil {
		t.Fatalf("postgres: %v", err)
	}
	t.Cleanup(pool.Close)
	lock, err := migrate.AcquireTestLock(ctx, pool)
	if err != nil {
		t.Fatalf("schema lock: %v", err)
	}
	t.Cleanup(func() { lock.Release(context.Background()) })
	if err := migrate.ResetLocked(ctx, pool, lock.Conn()); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	clock := &fakeClock{now: time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)}
	payload := validToken(clock.Now(), "hash-ok", "")
	payload.DeviceRecall = &integrity.DeviceRecallPayload{
		BitFirst: &integrity.DeviceRecallBit{BitValue: true},
	}
	seed := sha256.Sum256([]byte("recall-seed"))
	signer, _ := integrity.NewEd25519Signer(seed[:], "k", clock)
	rootPEM, leafDER, challenge := mustAttestationChain(t, true, integrity.VerifiedBootVerified, []byte("challenge-ok"))
	att, _ := integrity.NewKeyAttestationVerifier(rootPEM)
	devices := db.NewDeviceRepository(pool)
	events := db.NewIntegrityEventRepository(pool)
	dev := seedDevice(t, devices, "wv-recall")
	svc := integrity.NewService(integrity.Deps{
		Decoder: integrity.StubDecoder{Payload: payload},
		Devices: devices, Events: events, Signer: signer, Attestation: att, Clock: clock,
	})

	free, err := svc.Verify(ctx, integrity.VerifyRequest{
		DeviceID: dev.ID, WidevineHash: dev.WidevineHash, IntegrityToken: "tok",
		ExpectedRequestHash: "hash-ok", AttestationCertDERs: [][]byte{leafDER}, AttestationChallenge: challenge,
		HasPaidEntitlement: false,
	})
	if err != nil {
		t.Fatalf("free: %v", err)
	}
	if free.Verdict != integrity.VerdictFreeQuotaSuppressed || !free.FreeQuotaSuppressed {
		t.Fatalf("free=%+v", free)
	}
	if free.DeviceRecallBits&integrity.RecallBitFirst == 0 {
		t.Fatalf("bits=%d", free.DeviceRecallBits)
	}

	paid, err := svc.Verify(ctx, integrity.VerifyRequest{
		DeviceID: dev.ID, WidevineHash: dev.WidevineHash, IntegrityToken: "tok",
		ExpectedRequestHash: "hash-ok", AttestationCertDERs: [][]byte{leafDER}, AttestationChallenge: challenge,
		HasPaidEntitlement: true,
	})
	if err != nil {
		t.Fatalf("paid: %v", err)
	}
	if paid.Verdict != integrity.VerdictApproved || !paid.FreeQuotaSuppressed {
		t.Fatalf("paid users must operate with suppression flag only: %+v", paid)
	}

	reset, err := svc.ResetDeviceRecallBits(ctx, dev.ID)
	if err != nil || reset.DeviceRecallBits != 0 {
		t.Fatalf("reset=%+v err=%v", reset, err)
	}
}

func TestKeyAttestation_validAndRejects(t *testing.T) {
	challenge := []byte("att-challenge")
	rootPEM, leafOK, _ := mustAttestationChain(t, true, integrity.VerifiedBootVerified, challenge)
	v, err := integrity.NewKeyAttestationVerifier(rootPEM)
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}
	if err := v.Verify([][]byte{leafOK}, challenge); err != nil {
		t.Fatalf("valid chain: %v", err)
	}

	_, leafUnlocked, _ := mustAttestationChain(t, false, integrity.VerifiedBootVerified, challenge)
	if err := v.Verify([][]byte{leafUnlocked}, challenge); err == nil {
		t.Fatal("expected reject deviceLocked=false")
	}
	_, leafUnverified, _ := mustAttestationChain(t, true, 2, challenge) // Unverified
	if err := v.Verify([][]byte{leafUnverified}, challenge); err == nil {
		t.Fatal("expected reject verifiedBootState")
	}
	if err := v.Verify([][]byte{leafOK}, []byte("other-challenge")); err == nil {
		t.Fatal("expected reject challenge mismatch")
	}
}

func TestEd25519_signVerifyAndTamperExpiry(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)}
	seed := sha256.Sum256([]byte("ed25519-cross"))
	signer, err := integrity.NewEd25519Signer(seed[:], "kid-1", clock)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}
	env, _, err := signer.SignCanonical(map[string]any{"quota": int64(2)}, time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if err := signer.Verify(env, clock.Now()); err != nil {
		t.Fatalf("verify: %v", err)
	}

	tampered := env
	raw, _ := decodeB64(t, tampered.PayloadB64)
	raw[0] ^= 0xff
	tampered.PayloadB64 = encodeB64(raw)
	if err := signer.Verify(tampered, clock.Now()); err == nil {
		t.Fatal("expected tampered payload to fail")
	}

	sigTampered := env
	sig, _ := decodeB64(t, sigTampered.SignatureB64)
	sig[0] ^= 0xff
	sigTampered.SignatureB64 = encodeB64(sig)
	if err := signer.Verify(sigTampered, clock.Now()); err == nil {
		t.Fatal("expected tampered signature to fail")
	}

	clock.Set(clock.Now().Add(2 * time.Hour))
	if err := signer.Verify(env, clock.Now()); err == nil {
		t.Fatal("expected expired payload to fail")
	}
}

func mustAttestationChain(t *testing.T, locked bool, bootState int, challenge []byte) (rootPEM []byte, leafDER []byte, usedChallenge []byte) {
	t.Helper()
	rootKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("root key: %v", err)
	}
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("leaf key: %v", err)
	}
	rootTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "RadarKm Test Root"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	rootDER, err := x509.CreateCertificate(rand.Reader, rootTmpl, rootTmpl, &rootKey.PublicKey, rootKey)
	if err != nil {
		t.Fatalf("root cert: %v", err)
	}
	rootCert, err := x509.ParseCertificate(rootDER)
	if err != nil {
		t.Fatalf("parse root: %v", err)
	}
	extValue, err := integrity.BuildKeyDescriptionExtension(challenge, locked, bootState)
	if err != nil {
		t.Fatalf("ext: %v", err)
	}
	leafTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "RadarKm Attestation Leaf"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtraExtensions: []pkix.Extension{{
			Id:       []int{1, 3, 6, 1, 4, 1, 11129, 2, 1, 17},
			Critical: false,
			Value:    extValue,
		}},
	}
	leafDER, err = x509.CreateCertificate(rand.Reader, leafTmpl, rootCert, &leafKey.PublicKey, rootKey)
	if err != nil {
		t.Fatalf("leaf cert: %v", err)
	}
	rootPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootDER})
	return rootPEM, leafDER, challenge
}

func decodeB64(t *testing.T, s string) ([]byte, error) {
	t.Helper()
	return base64.StdEncoding.DecodeString(s)
}

func encodeB64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
