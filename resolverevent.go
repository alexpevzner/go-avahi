// CGo binding for Avahi
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
import "fmt"

// ResolverEvent is the CGo representation of the [AvahiResolverEvent].
//
// [AvahiResolverEvent]: https://avahi.org/doxygen/html/defs_8h.html#ae524657615ba2ec3b17613098a3394cf
type ResolverEvent int

// ResolverEvent values:
const (
	// Successful resolving
	ResolverFound ResolverEvent = C.AVAHI_RESOLVER_FOUND

	// Resolving failed due to some reason.
	ResolverFailure ResolverEvent = C.AVAHI_RESOLVER_FAILURE
)

// resolverEventNames contains names for known resolver events.
var resolverEventNames = map[ResolverEvent]string{
	ResolverFound:   "ResolverFound",
	ResolverFailure: "ResolverFailure",
}

// String returns a name of ResolverEvent
func (e ResolverEvent) String() string {
	n := resolverEventNames[e]
	if n == "" {
		n = fmt.Sprintf("UNKNOWN %d", int(e))
	}
	return n
}
