// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Avahi browser events
//
//go:build linux || freebsd

package avahi

// #include <avahi-client/client.h>
import "C"
import "fmt"

// BrowserEvent is the CGo representation of [AvahiBrowserEvent].
//
// [AvahiBrowserEvent]: https://avahi.org/doxygen/html/defs_8h.html#af7ff3b95259b3441a282b87d82eebd87
type BrowserEvent int

// BrowserEvent values:
const (
	// New object discovered on the network.
	BrowserNew BrowserEvent = C.AVAHI_BROWSER_NEW

	// The object has been removed from the network.
	BrowserRemove BrowserEvent = C.AVAHI_BROWSER_REMOVE

	// One-time event, to notify the user that all entries from
	// the cache have been sent.
	BrowserCacheExhausted BrowserEvent = C.AVAHI_BROWSER_CACHE_EXHAUSTED

	// One-time event, to hint the user that more records
	// are unlikely to be shown in the near feature.
	BrowserAllForNow BrowserEvent = C.AVAHI_BROWSER_ALL_FOR_NOW

	// Browsing failed with a error.
	BrowserFailure BrowserEvent = C.AVAHI_BROWSER_FAILURE
)

// browserEventNames contains names for known browser events.
var browserEventNames = map[BrowserEvent]string{
	BrowserNew:            "BrowserNew",
	BrowserRemove:         "BrowserRemove",
	BrowserCacheExhausted: "BrowserCacheExhausted",
	BrowserAllForNow:      "BrowserAllForNow",
	BrowserFailure:        "BrowserAllForNow",
}

// String returns a name of BrowserEvent
func (e BrowserEvent) String() string {
	n := browserEventNames[e]
	if n == "" {
		n = fmt.Sprintf("UNKNOWN %d", int(e))
	}
	return n
}
