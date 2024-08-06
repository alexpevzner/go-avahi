// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Avahi IP4/IP6 protocol
//
//go:build linux || freebsd

package avahi

// #include <avahi-common/address.h>
import "C"

// Protocol specifies IP4/IP6 protocol
type Protocol int

// Protocol values:
const (
	ProtocolIP4    = Protocol(C.AVAHI_PROTO_INET)
	ProtocolIP6    = Protocol(C.AVAHI_PROTO_INET6)
	ProtocolUnspec = Protocol(C.AVAHI_PROTO_UNSPEC)
)
