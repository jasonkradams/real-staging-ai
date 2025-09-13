package http

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

// helper to build Stripe-Signature header string with given timestamp and one or more v1 signatures.
func makeSigHeader(ts int64, sigs ...[]byte) string {
	header := fmt.Sprintf("t=%d", ts)
	for _, s := range sigs {
		header += fmt.Sprintf(",v1=%s", hex.EncodeToString(s))
	}
	return header
}

func TestParseStripeSignatureHeader_Success(t *testing.T) {
	ts := int64(1700000000)
	// two valid hex signatures: 0x00 and 0xaa
	header := fmt.Sprintf("t=%d,v1=00,v1=aa", ts)

	gotTS, v1s, err := parseStripeSignatureHeader(header)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gotTS != ts {
		t.Fatalf("expected ts=%d, got %d", ts, gotTS)
	}
	if len(v1s) != 2 {
		t.Fatalf("expected 2 signatures, got %d", len(v1s))
	}
	if !bytes.Equal(v1s[0], []byte{0x00}) {
		t.Fatalf("expected first sig = 00, got %x", v1s[0])
	}
	if !bytes.Equal(v1s[1], []byte{0xaa}) {
		t.Fatalf("expected second sig = aa, got %x", v1s[1])
	}
}

func TestParseStripeSignatureHeader_MissingTimestamp(t *testing.T) {
	header := "v1=00"

	_, _, err := parseStripeSignatureHeader(header)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if want := "missing timestamp"; err == nil || !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func TestParseStripeSignatureHeader_InvalidTimestamp(t *testing.T) {
	header := "t=abc,v1=00"

	_, _, err := parseStripeSignatureHeader(header)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if want := "invalid timestamp"; !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func TestParseStripeSignatureHeader_InvalidV1Hex(t *testing.T) {
	header := "t=1700000000,v1=zz"

	_, _, err := parseStripeSignatureHeader(header)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if want := "invalid v1 signature"; !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func TestVerifyStripeSignature_Valid(t *testing.T) {
	body := []byte(`{"id":"evt_123"}`)
	secret := "whsec_test"
	ts := int64(1700000000)

	expected := computeStripeSignature(body, ts, secret)
	header := makeSigHeader(ts, expected)

	now := func() time.Time { return time.Unix(ts+60, 0) } // within 5 minutes
	if err := verifyStripeSignature(body, header, secret, 5*time.Minute, now); err != nil {
		t.Fatalf("expected valid signature, got error: %v", err)
	}
}

func TestVerifyStripeSignature_NoHeader(t *testing.T) {
	body := []byte(`{}`)
	secret := "whsec_test"
	now := func() time.Time { return time.Unix(1700000000, 0) }

	err := verifyStripeSignature(body, "", secret, 5*time.Minute, now)
	if err == nil {
		t.Fatalf("expected error for missing header, got nil")
	}
	if want := "missing Stripe-Signature header"; !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func TestVerifyStripeSignature_NoV1Provided(t *testing.T) {
	body := []byte(`{}`)
	secret := "whsec_test"
	ts := int64(1700000000)
	header := fmt.Sprintf("t=%d", ts)
	now := func() time.Time { return time.Unix(ts, 0) }

	err := verifyStripeSignature(body, header, secret, 5*time.Minute, now)
	if err == nil {
		t.Fatalf("expected error when no v1 signatures, got nil")
	}
	if want := "no v1 signatures found"; !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func TestVerifyStripeSignature_MismatchedSignature(t *testing.T) {
	body := []byte(`{"id":"evt_123"}`)
	secret := "whsec_test"
	badSecret := "whsec_other"
	ts := int64(1700000000)

	// signature computed with wrong secret
	wrong := computeStripeSignature(body, ts, badSecret)
	header := makeSigHeader(ts, wrong)

	now := func() time.Time { return time.Unix(ts, 0) }
	err := verifyStripeSignature(body, header, secret, 5*time.Minute, now)
	if err == nil {
		t.Fatalf("expected signature mismatch error, got nil")
	}
	if want := "no matching signature"; !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func TestVerifyStripeSignature_MultipleV1_OneMatches(t *testing.T) {
	body := []byte(`{"ok":true}`)
	secret := "whsec_test"
	ts := int64(1700000000)

	// first sig is bogus but valid hex, second is correct
	bogus := []byte("bogus-signature-bytes") // arbitrary
	good := computeStripeSignature(body, ts, secret)
	header := makeSigHeader(ts, bogus, good)

	now := func() time.Time { return time.Unix(ts, 0) }
	if err := verifyStripeSignature(body, header, secret, 5*time.Minute, now); err != nil {
		t.Fatalf("expected success when one of multiple signatures matches, got %v", err)
	}
}

func TestVerifyStripeSignature_TimestampOutsideTolerance(t *testing.T) {
	body := []byte(`{"id":"evt_999"}`)
	secret := "whsec_test"
	ts := int64(1700000000)

	sig := computeStripeSignature(body, ts, secret)
	header := makeSigHeader(ts, sig)

	// now is 10 minutes ahead, tolerance 5 minutes -> should fail
	now := func() time.Time { return time.Unix(ts, 0).Add(10 * time.Minute) }
	err := verifyStripeSignature(body, header, secret, 5*time.Minute, now)
	if err == nil {
		t.Fatalf("expected timestamp tolerance error, got nil")
	}
	if want := "timestamp outside tolerance window"; !contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

// contains is a tiny helper to avoid importing testify.
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
