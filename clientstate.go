// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Avahi Client State
//
//go:build linux || freebsd

package avahi

// #include <avahi-client/client.h>
import "C"
import "fmt"

// ClientState represents a [Client] state.
type ClientState int

// ClientState values:
const (
	// Invalid (zero) state. This state is reported when
	// [Client] is closed with the [Client.Close] call.
	ClientStateClosed ClientState = 0

	// Avahi server is being registering host RRs on a network
	ClientStateRegistering ClientState = C.AVAHI_CLIENT_S_REGISTERING

	// Ahavi server is up and running
	ClientStateRunning ClientState = C.AVAHI_CLIENT_S_RUNNING

	// Avahi server was not able to register host RRs due to collision
	// with some another host.
	//
	// Administrator needs to update the host name to avoid the
	// collision.
	ClientStateCollision ClientState = C.AVAHI_CLIENT_S_COLLISION

	// Avahi server failure.
	ClientStateFailure ClientState = C.AVAHI_CLIENT_FAILURE

	// Avahi Client is trying to connect the server.
	ClientStateConnecting ClientState = C.AVAHI_CLIENT_CONNECTING
)

// clientStateNames contains names for known client states.
var clientStateNames = map[ClientState]string{
	ClientStateClosed:      "closed",
	ClientStateRegistering: "registering",
	ClientStateRunning:     "running",
	ClientStateCollision:   "collision",
	ClientStateFailure:     "failure",
	ClientStateConnecting:  "connecting",
}

// String returns a name of the ClientState.
func (state ClientState) String() string {
	n := clientStateNames[state]
	if n == "" {
		n = fmt.Sprintf("UNKNOWN 0x%4.4x", int(state))
	}
	return n
}
