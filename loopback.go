// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Loopback interface
//
//go:build linux || freebsd

package avahi

import (
	"fmt"
	"net"
	"sync/atomic"
)

// Cached loopback interface index
var loopback int32 = -1

// Loopback returns index of the loopback network interface.
//
// This function may fail, if [net.Interfaces] fails or there
// is no loopback interface in the response.
//
// As this function is extremely unlikely to fail, you may consider
// using [MustLoopback] instead.
func Loopback() (IfIndex, error) {
	// Lookup cache
	idx := atomic.LoadInt32(&loopback)
	if idx != -1 {
		return IfIndex(idx), nil
	}

	// Consult net.Interfaces
	ift, err := net.Interfaces()
	if err != nil {
		return 0, fmt.Errorf("avahi.Loopback: %w", err)
	}

	for _, ifi := range ift {
		if ifi.Flags&net.FlagLoopback != 0 {
			atomic.StoreInt32(&loopback, int32(ifi.Index))
			return IfIndex(ifi.Index), nil
		}
	}

	return 0, fmt.Errorf("avahi.Loopback: interface not found")
}

// MustLoopback returns index of the loopback network interface.
//
// This is convenience wrapper around the [Loopback] function. If
// Loopback function fails, MustLoopback panics instead of returning
// the error.
func MustLoopback() IfIndex {
	idx, err := Loopback()
	if err != nil {
		panic(err)
	}
	return idx
}
