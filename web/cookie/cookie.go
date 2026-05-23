// Package cookie provides utilities for safe cookie management.
//
// We use a server-side secret to verify cookie authenticity, using an hmac+sha256 signature, in a
// manner vaguely inspired by JWT. For our purposes, we don't need to enforce cookie lifetime, or
// bind it to the user in any particular way. Changing server side secret can be used to invalidate
// all prior cookies.
package cookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type container struct {
	Payload   []byte `json:"p"`
	Signature []byte `json:"s"`
}

func zero[T any]() T {
	var z T
	return z
}

// ParseCookie verifies cookie authenticity and decodes it payload.
func ParseCookie[T any](raw string, secret []byte) (T, error) {
	if len(secret) == 0 {
		return zero[T](), fmt.Errorf("secret must not be empty")
	}

	// Base64-decode, then unmarshal the cookie container.
	decoded, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		return zero[T](), fmt.Errorf("failed to base64-decode cookie: %w", err)
	}
	c := container{}
	if err := json.Unmarshal(decoded, &c); err != nil {
		return zero[T](), fmt.Errorf("failed to unmarshal cookie container: %w", err)
	}

	// Compute the payload signature and compare it with the provided to validate authenticity.
	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write(c.Payload); err != nil {
		return zero[T](), fmt.Errorf("failed to compute cookie signature: %w", err)
	}
	if !hmac.Equal(mac.Sum(nil), c.Signature) {
		return zero[T](), fmt.Errorf("cookie signature mismatch")
	}

	// Decode cookie payload into the concrete Go type.
	var cookie T
	if err := json.Unmarshal(c.Payload, &cookie); err != nil {
		return zero[T](), fmt.Errorf("failed to decode cookie payload: %w", err)
	}

	return cookie, nil
}

// SaveCookie generates a signed cookie value that can be passed to the client.
func SaveCookie[T any](cookie T, secret []byte) (string, error) {
	if len(secret) == 0 {
		return "", fmt.Errorf("secret must not be empty")
	}

	c := container{}

	// Encode cookie payload into a byte array.
	if b, err := json.Marshal(cookie); err != nil {
		return "", fmt.Errorf("failed to encode cookie payload: %w", err)
	} else {
		c.Payload = b
	}

	// Sign cookie payload.
	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write(c.Payload); err != nil {
		return "", fmt.Errorf("failed to compute cookie signature: %w", err)
	}
	c.Signature = mac.Sum(nil)

	// Marshal the cookie container, then base64-encode for safe cookie transport.
	j, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cookie container: %w", err)
	}
	return base64.URLEncoding.EncodeToString(j), nil
}
