// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Network interface indices
//
//go:build linux || freebsd

package avahi

// #include <avahi-common/address.h>
import "C"

// IfIndex specifies network interface index
type IfIndex int

// IfIndex values:
const (
	IfIndexUnspec = IfIndex(C.AVAHI_PROTO_UNSPEC)
)
