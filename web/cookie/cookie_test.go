package cookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var testSecret = []byte("test-secret-key")

type testPayload struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
}

func testRoundTrip[T any](t *testing.T, original T) {
	t.Helper()
	raw, err := SaveCookie(original, testSecret)
	if err != nil {
		t.Fatalf("SaveCookie(%+v) failed: %v", original, err)
	}

	got, err := ParseCookie[T](raw, testSecret)
	if err != nil {
		t.Fatalf("ParseCookie(%q) failed: %v", raw, err)
	}

	if diff := cmp.Diff(original, got); diff != "" {
		t.Errorf("ParseCookie returned unexpected value (-want +got):\n%s", diff)
	}
}

func TestRoundTrip(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		testRoundTrip(t, testPayload{UserID: 42, Name: "alice"})
	})
	t.Run("scalar", func(t *testing.T) {
		testRoundTrip(t, "hello")
	})
}

func TestWrongSecret(t *testing.T) {
	raw, err := SaveCookie(testPayload{UserID: 1}, testSecret)
	if err != nil {
		t.Fatalf("SaveCookie failed: %v", err)
	}

	got, err := ParseCookie[testPayload](raw, []byte("different-secret"))
	if err == nil {
		t.Fatalf("ParseCookie with wrong secret succeeded and returned %+v, want signature mismatch error", got)
	}
}

func TestTampering(t *testing.T) {
	cases := []struct {
		name   string
		tamper func(*container)
	}{
		{
			name: "payload replaced",
			tamper: func(c *container) {
				c.Payload, _ = json.Marshal(testPayload{UserID: 999, Name: "attacker"})
			},
		},
		{
			name: "signature byte flipped",
			tamper: func(c *container) {
				c.Signature[0] ^= 0xff
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := SaveCookie(testPayload{UserID: 1, Name: "alice"}, testSecret)
			if err != nil {
				t.Fatalf("SaveCookie failed: %v", err)
			}

			// Decode the base64 cookie, unmarshal the container, tamper, re-encode.
			jsonBytes, err := base64.URLEncoding.DecodeString(raw)
			if err != nil {
				t.Fatalf("base64 decode: %v", err)
			}
			c := container{}
			if err := json.Unmarshal(jsonBytes, &c); err != nil {
				t.Fatalf("unmarshal cookie container: %v", err)
			}
			tc.tamper(&c)
			tamperedJSON, _ := json.Marshal(c)
			tampered := base64.URLEncoding.EncodeToString(tamperedJSON)

			got, err := ParseCookie[testPayload](tampered, testSecret)
			if err == nil {
				t.Fatalf("ParseCookie accepted tampered cookie (%s) and returned %+v", tc.name, got)
			}
		})
	}
}

func TestParseRejectsInvalidInput(t *testing.T) {
	validPayload, _ := json.Marshal(testPayload{UserID: 1})
	// containerWith builds a valid base64-encoded container with the given signature.
	containerWith := func(sig []byte) string {
		j, _ := json.Marshal(container{Payload: validPayload, Signature: sig})
		return base64.URLEncoding.EncodeToString(j)
	}

	cases := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"not base64", "not base64!!!"},
		{"valid base64 but not json", base64.URLEncoding.EncodeToString([]byte("not json"))},
		{"null", base64.URLEncoding.EncodeToString([]byte("null"))},
		{"empty signature", containerWith([]byte{})},
		{"nil signature", containerWith(nil)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseCookie[testPayload](tc.input, testSecret)
			if err == nil {
				t.Errorf("ParseCookie returned %+v, want error", got)
			}
		})
	}
}

func TestValidContainerButInvalidPayload(t *testing.T) {
	inner, _ := json.Marshal("this is a string, not a struct")
	c := container{Payload: inner}

	h := hmac.New(sha256.New, testSecret)
	h.Write(c.Payload)
	c.Signature = h.Sum(nil)

	j, _ := json.Marshal(c)
	raw := base64.URLEncoding.EncodeToString(j)

	got, err := ParseCookie[testPayload](raw, testSecret)
	if err == nil {
		t.Fatalf("ParseCookie accepted a string payload into a struct type and returned %+v", got)
	}
}

func TestEmptySecret(t *testing.T) {
	if _, err := SaveCookie(testPayload{UserID: 1}, []byte{}); err == nil {
		t.Fatal("SaveCookie with empty secret succeeded, want error")
	}

	raw, _ := SaveCookie(testPayload{UserID: 1}, testSecret)
	if _, err := ParseCookie[testPayload](raw, []byte{}); err == nil {
		t.Fatal("ParseCookie with empty secret succeeded, want error")
	}
}
