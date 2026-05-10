package db

import (
	"database/sql/driver"
	"net/netip"
)

// Addr wraps netip.Addr to provide database/sql Valuer and Scanner
// implementations so it can be stored in the database as a textual IP address.
type Addr netip.Addr

func (a Addr) Unwrap() netip.Addr {
	return netip.Addr(a)
}

// Value implements driver.Valuer. It returns SQL NULL if the address is invalid.
func (a Addr) Value() (driver.Value, error) {
	return a.Unwrap().MarshalText()
}

// Scan implements sql.Scanner. It accepts string or []byte input with the
// textual representation of an IP address.
func (a *Addr) Scan(src any) error {
	return (*netip.Addr)(a).UnmarshalText(src.([]byte))
}
