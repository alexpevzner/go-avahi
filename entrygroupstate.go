// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Avahi Entry Group State
//
//go:build linux || freebsd

package avahi

import "fmt"

// #include <avahi-client/publish.h>
import "C"

// EntryGroupState represents an [EntryGroup] state.
type EntryGroupState int

// EntryGroupState values:
const (
	// The group has not yet been commited
	EntryGroupStateUncommited EntryGroupState = C.AVAHI_ENTRY_GROUP_UNCOMMITED

	// The group is currently being registered
	EntryGroupStateRegistering EntryGroupState = C.AVAHI_ENTRY_GROUP_REGISTERING

	// The group has been successfully established
	EntryGroupStateEstablished EntryGroupState = C.AVAHI_ENTRY_GROUP_ESTABLISHED

	// A name collision for one of entries in the group has been detected.
	// The entries has been withdrawn.
	EntryGroupStateCollision EntryGroupState = C.AVAHI_ENTRY_GROUP_COLLISION

	// Some kind of failure has been detected, the entries has been withdrawn.
	EntryGroupStateFailure EntryGroupState = C.AVAHI_ENTRY_GROUP_FAILURE
)

// clientStateNames contains names for known client states.
var entryGroupStateNames = map[EntryGroupState]string{
	EntryGroupStateUncommited:  "uncommited",
	EntryGroupStateRegistering: "registering",
	EntryGroupStateEstablished: "established",
	EntryGroupStateCollision:   "collision",
	EntryGroupStateFailure:     "failure",
}

// String returns a name of the EntryGroupState.
func (state EntryGroupState) String() string {
	n := entryGroupStateNames[state]
	if n == "" {
		n = fmt.Sprintf("UNKNOWN 0x%4.4x", int(state))
	}
	return n
}
