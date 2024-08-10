// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Closers
//
//go:build linux || freebsd

package avahi

// closer is the object that can be closed
type closer interface {
	Close()
}

// closers is a set of closers
type closers map[closer]struct{}

// init initializes the set
func (set *closers) init() {
	*set = make(closers)
}

// add adds object into the set
func (set closers) add(obj closer) {
	set[obj] = struct{}{}
}

// del deletes object from the set
func (set closers) del(obj closer) {
	delete(set, obj)
}

// close closes all objects still in set
func (set closers) close() {
	for obj := range set {
		obj.Close()
	}
}
