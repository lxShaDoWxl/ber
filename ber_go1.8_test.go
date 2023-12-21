//go:build !go1.9
// +build !go1.9

package ber

import "encoding/asn1"

// Compatibility vars for ber_asn1_test.go
var (
	NullRawValue = asn1.RawValue{Tag: asn1.TagNull}
	NullBytes    = []byte{asn1.tagNull, 0}
)
