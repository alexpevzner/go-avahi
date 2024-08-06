// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Avahi lookup flags
//
//go:build linux || freebsd

package avahi

import (
	"strings"
)

// #include <avahi-common/defs.h>
import "C"

// LookupFlags provides some options for lookup functions
type LookupFlags int

// LookupFlags values
const (
	// Force lookup via wide area DNS
	LookupUseWideArea LookupFlags = C.AVAHI_LOOKUP_USE_WIDE_AREA

	// Force lookup via multicast DNS
	LookupUseMulticast LookupFlags = C.AVAHI_LOOKUP_USE_MULTICAST

	// When doing service resolving, don't lookup TXT record
	LookupNoTXT LookupFlags = C.AVAHI_LOOKUP_NO_TXT

	// When doing service resolving, don't lookup A/AAAA records
	LookupNoAddress LookupFlags = C.AVAHI_LOOKUP_NO_ADDRESS
)

// String returns LookupFlags as string, for debugging
func (flags LookupFlags) String() string {
	s := []string{}

	if flags&LookupUseWideArea != 0 {
		s = append(s, "use-wan")
	}
	if flags&LookupUseMulticast != 0 {
		s = append(s, "use-mdns")
	}
	if flags&LookupNoTXT != 0 {
		s = append(s, "no-txt")
	}
	if flags&LookupNoAddress != 0 {
		s = append(s, "no-addr")
	}

	return strings.Join(s, ",")
}

// LookupResultFlags provides some additional information about
// lookup response.
type LookupResultFlags int

// LookupResultFlags bits:
const (
	// This response originates from the cache
	LookupResultCached LookupResultFlags = C.AVAHI_LOOKUP_RESULT_CACHED

	// This response originates from wide area DNS
	LookupResultWideArea LookupResultFlags = C.AVAHI_LOOKUP_RESULT_WIDE_AREA

	// This response originates from multicast DNS
	LookupResultMulticast LookupResultFlags = C.AVAHI_LOOKUP_RESULT_MULTICAST

	// This record/service resides on and was announced by the local host.
	// Only available in service and record browsers and only on
	// BrowserNew event.
	LookupResultLocal LookupResultFlags = C.AVAHI_LOOKUP_RESULT_LOCAL

	// This service belongs to the same local client as the browser object.
	// Only for service browsers and only on BrowserNew event.
	LookupResultOurOwn LookupResultFlags = C.AVAHI_LOOKUP_RESULT_OUR_OWN

	// The returned data was defined statically by server configuration.
	LookupResultStatic LookupResultFlags = C.AVAHI_LOOKUP_RESULT_STATIC
)

// String returns LookupResultFlags as string, for debugging
func (flags LookupResultFlags) String() string {
	s := []string{}

	if flags&LookupResultCached != 0 {
		s = append(s, "cached")
	}
	if flags&LookupResultWideArea != 0 {
		s = append(s, "wan-dns")
	}
	if flags&LookupResultMulticast != 0 {
		s = append(s, "mdns")
	}
	if flags&LookupResultCached != 0 {
		s = append(s, "local")
	}
	if flags&LookupResultOurOwn != 0 {
		s = append(s, "our-own")
	}
	if flags&LookupResultStatic != 0 {
		s = append(s, "static")
	}

	return strings.Join(s, ",")
}
