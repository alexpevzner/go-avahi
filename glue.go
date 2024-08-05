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

import "sync"

// #cgo pkg-config: avahi-client
//
// #include <avahi-client/client.h>
import "C"

// CGo doesn't allow us to save Go pointers at the C side (for callbacks),
// so we need some magic to access Go structures from C callbacks
var (
	clientMap = newCgomap[*C.AvahiClient, *Client]()
)

// cgomap is a generic goroutine-safe map
//
// we use it here to map C-side object into corresponding Go-side objects.
// We have to do it, because CGo doesn't allow us to save Go pointers at
// the C side.
type cgomap[C comparable, GO any] struct {
	m map[C]GO
	l sync.Mutex
}

// newCgomap creates a new cgomap
func newCgomap[C comparable, GO any]() *cgomap[C, GO] {
	return &cgomap[C, GO]{
		m: make(map[C]GO),
	}
}

// Put creates a mapping from the C-side object c to the Go-side value g
func (cgm *cgomap[C, GO]) Put(c C, g GO) {
	cgm.l.Lock()
	defer cgm.l.Unlock()

	cgm.m[c] = g
}

// Get gets Go object that corresponds to the specified C object
func (cgm *cgomap[C, GO]) Get(c C) GO {
	cgm.l.Lock()
	defer cgm.l.Unlock()

	return cgm.m[c]
}

// Del deletes the mapping previously created by cgomap.Put
func (cgm *cgomap[C, GO]) Del(c C) {
	cgm.l.Lock()
	defer cgm.l.Unlock()

	delete(cgm.m, c)
}
