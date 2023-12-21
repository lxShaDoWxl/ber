// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ber

import (
	"encoding/asn1"
	"reflect"
	"strconv"
	"strings"
)

type tagAndLength struct {
	class, tag, length int
	isCompound         bool
	isIndefinite       bool
}

// ASN.1 has IMPLICIT and EXPLICIT tags, which can be translated as "instead
// of" and "in addition to". When not specified, every primitive type has a
// default tag in the UNIVERSAL class.
//
// For example: a BIT STRING is tagged [UNIVERSAL 3] by default (although ASN.1
// doesn't actually have a UNIVERSAL keyword). However, by saying [IMPLICIT
// CONTEXT-SPECIFIC 42], that means that the tag is replaced by another.
//
// On the other hand, if it said [EXPLICIT CONTEXT-SPECIFIC 10], then an
// /additional/ tag would wrap the default tag. This explicit tag will have the
// compound flag set.
//
// (This is used in order to remove ambiguity with optional elements.)
//
// You can layer EXPLICIT and IMPLICIT tags to an arbitrary depth, however we
// don't support that here. We support a single layer of EXPLICIT or IMPLICIT
// tagging with tag strings on the fields of a structure.

// fieldParameters is the parsed representation of tag string from a structure field.
type fieldParameters struct {
	optional     bool   // true iff the field is OPTIONAL
	explicit     bool   // true iff an EXPLICIT tag is in use.
	application  bool   // true iff an APPLICATION tag is in use.
	private      bool   // true iff a PRIVATE tag is in use.
	defaultValue *int64 // a default value for INTEGER typed fields (maybe nil).
	tag          *int   // the EXPLICIT or IMPLICIT tag (maybe nil).
	stringType   int    // the string tag to use when marshaling.
	timeType     int    // the time tag to use when marshaling.
	set          bool   // true iff this should be encoded as a SET
	omitEmpty    bool   // true iff this should be omitted if empty when marshaling.

	// Invariants:
	//   if explicit is set, tag is non-nil.
}

// Given a tag string with the format specified in the package comment,
// parseFieldParameters will parse it into a fieldParameters structure,
// ignoring unknown parts of the string.
func parseFieldParameters(str string) (ret fieldParameters) {
	for _, part := range strings.Split(str, ",") {
		switch {
		case part == "optional":
			ret.optional = true
		case part == "explicit":
			ret.explicit = true
			if ret.tag == nil {
				ret.tag = new(int)
			}
		case part == "generalized":
			ret.timeType = asn1.TagGeneralizedTime
		case part == "utc":
			ret.timeType = asn1.TagUTCTime
		case part == "ia5":
			ret.stringType = asn1.TagIA5String
		case part == "printable":
			ret.stringType = asn1.TagPrintableString
		case part == "numeric":
			ret.stringType = asn1.TagNumericString
		case part == "utf8":
			ret.stringType = asn1.TagUTF8String
		case strings.HasPrefix(part, "default:"):
			i, err := strconv.ParseInt(part[8:], 10, 64)
			if err == nil {
				ret.defaultValue = new(int64)
				*ret.defaultValue = i
			}
		case strings.HasPrefix(part, "tag:"):
			i, err := strconv.Atoi(part[4:])
			if err == nil {
				ret.tag = new(int)
				*ret.tag = i
			}
		case part == "set":
			ret.set = true
		case part == "application":
			ret.application = true
			if ret.tag == nil {
				ret.tag = new(int)
			}
		case part == "private":
			ret.private = true
			if ret.tag == nil {
				ret.tag = new(int)
			}
		case part == "omitempty":
			ret.omitEmpty = true
		}
	}
	return
}

// Given a reflected Go type, getUniversalType returns the default tag number
// and expected compound flag.
func getUniversalType(t reflect.Type) (matchAny bool, tagNumber int, isCompound, ok bool) {
	switch t {
	case rawValueType:
		return true, -1, false, true
	case objectIdentifierType:
		return false, asn1.TagOID, false, true
	case bitStringType:
		return false, asn1.TagBitString, false, true
	case timeType:
		return false, asn1.TagUTCTime, false, true
	case enumeratedType:
		return false, asn1.TagEnum, false, true
	case bigIntType:
		return false, asn1.TagInteger, false, true
	}
	switch t.Kind() {
	case reflect.Bool:
		return false, asn1.TagBoolean, false, true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return false, asn1.TagInteger, false, true
	case reflect.Struct:
		return false, asn1.TagSequence, true, true
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return false, asn1.TagOctetString, false, true
		}
		if strings.HasSuffix(t.Name(), "SET") {
			return false, asn1.TagSet, true, true
		}
		return false, asn1.TagSequence, true, true
	case reflect.String:
		return false, asn1.TagPrintableString, false, true
	}
	return false, 0, false, false
}
