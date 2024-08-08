// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Avahi publishing flags
//
//go:build linux || freebsd

package avahi

import (
	"strings"
)

// #include <avahi-common/defs.h>
import "C"

// PublishFlags represents flags for publishing functions
type PublishFlags int

// PublishFlags for raw records:
const (
	// RRset is intended to be unique
	PublishUnique PublishFlags = C.AVAHI_PUBLISH_UNIQUE
	// hough the RRset is intended to be unique no probes shall be sent
	PublishNoProbe PublishFlags = C.AVAHI_PUBLISH_NO_PROBE
	// Do not announce this RR to other hosts
	PublishNoAnnounce PublishFlags = C.AVAHI_PUBLISH_NO_ANNOUNCE
	// Allow multiple local records of this type
	PublishAllowMultiple PublishFlags = C.AVAHI_PUBLISH_ALLOW_MULTIPLE
)

// PublishFlags for address records:
const (
	// Don't create a reverse (PTR) entry
	PublishNoReverse PublishFlags = C.AVAHI_PUBLISH_NO_REVERSE
	// Do not implicitly add the local service cookie to TXT data
	PublishNoCookie PublishFlags = C.AVAHI_PUBLISH_NO_COOKIE
)

// Other PublishFlags:
const (
	// Update existing records instead of adding new ones
	PublishUpdate PublishFlags = C.AVAHI_PUBLISH_UPDATE
	// Register the record using wide area DNS (i.e. unicast DNS update)
	PublishUseWideArea PublishFlags = C.AVAHI_PUBLISH_USE_WIDE_AREA
	// Register the record using multicast DNS
	PublishUseMulticast PublishFlags = C.AVAHI_PUBLISH_USE_MULTICAST
)

// String returns PublishFlags as string, for debugging
func (flags PublishFlags) String() string {
	s := []string{}

	if flags&PublishUnique != 0 {
		s = append(s, "unique")
	}
	if flags&PublishNoProbe != 0 {
		s = append(s, "no-probe")
	}
	if flags&PublishNoAnnounce != 0 {
		s = append(s, "no-announce")
	}
	if flags&PublishAllowMultiple != 0 {
		s = append(s, "allow-multiple")
	}

	if flags&PublishNoReverse != 0 {
		s = append(s, "no-reverse")
	}
	if flags&PublishNoCookie != 0 {
		s = append(s, "no-cookie")
	}

	if flags&PublishUpdate != 0 {
		s = append(s, "update")
	}
	if flags&PublishUseWideArea != 0 {
		s = append(s, "use-wan")
	}
	if flags&PublishUseMulticast != 0 {
		s = append(s, "use-mdns")
	}

	return strings.Join(s, ",")
}
