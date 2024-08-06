// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// CGo glue
//
//go:build linux || freebsd

package avahi

// #cgo pkg-config: avahi-client
//
// #include <avahi-client/client.h>
import "C"
