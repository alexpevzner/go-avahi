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

// #include <avahi-client/client.h>
import "C"

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
	if LookupResultLocal&LookupResultCached != 0 {
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
