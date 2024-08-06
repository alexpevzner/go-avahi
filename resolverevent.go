// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Avahi resolver events
//
//go:build linux || freebsd

package avahi

// #include <avahi-client/client.h>
import "C"

// ResolverEvent is the CGo representation of the [AvahiResolverEvent].
//
// [AvahiResolverEvent]: https://avahi.org/doxygen/html/defs_8h.html#ae524657615ba2ec3b17613098a3394cf
type ResolverEvent int

// ResolverEvent values:
const (
	// Successful resolving
	ResolverFould ResolverEvent = C.AVAHI_RESOLVER_FOUND

	// Resolving failed due to some reason.
	ResolverFailure ResolverEvent = C.AVAHI_RESOLVER_FAILURE
)
