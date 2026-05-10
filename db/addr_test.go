package db

import (
	"net/netip"
	"testing"
)

func TestAddr_ValueScanRoundtrip(t *testing.T) {
	testCases := []struct {
		name string
		addr Addr
	}{
		{
			name: "ipv4",
			addr: Addr(netip.MustParseAddr("127.0.0.1")),
		},
		{
			name: "ipv6",
			addr: Addr(netip.MustParseAddr("2001:db8::1")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := tc.addr.Value()
			if err != nil {
				t.Fatalf("Addr.Value() returned error: %v", err)
			}

			encoded, ok := value.([]byte)
			if !ok {
				t.Fatalf("Addr.Value() returned %T, want []byte", value)
			}
			if got, want := string(encoded), tc.addr.Unwrap().String(); got != want {
				t.Fatalf("Addr.Value() = %q, want %q", got, want)
			}

			var got Addr
			if err := got.Scan(encoded); err != nil {
				t.Fatalf("Addr.Scan() returned error: %v", err)
			}

			if got != tc.addr {
				t.Errorf("Addr roundtrip mismatch: got %v, want %v", got.Unwrap(), tc.addr.Unwrap())
			}
		})
	}
}
