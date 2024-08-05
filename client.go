// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Avahi Client
//
//go:build linux || freebsd

package avahi

import (
	"errors"
	"fmt"
	"unsafe"
)

// #include <avahi-client/client.h>
// #include <avahi-common/thread-watch.h>
//
// void clientCallback(AvahiClient*, AvahiClientState, void*);
import "C"

// Client is the CGo binding for [AvahiClient].
// It represents a client connection to the Avahi daemon.
//
// Client may change its state dynamically. [ClientState] reported via
// channel, not via callbacks as in C version of Avahi API.
//
// When Client is not in use anymore, it must be closed using the
// Client.Close call to free associated resources. Closing the client
// closes its event notifications channel, effectively unblocking
// pending readers.
//
// [AvahiClient]: https://avahi.org/doxygen/html/client_8h.html#a3d65e9ea7182c44fa8df04a72f1a56bb
type Client struct {
	avahiClient  *C.AvahiClient       // Underlying AvahiClient
	threadedPoll *C.AvahiThreadedPoll // Avahi event loop
	state        ClientState          // Last known state
	evq          *queue[ClientState]  // Event queue
}

// NewClient creates a new [Client].
func NewClient() (*Client, error) {
	// Create Avahi event loop. We use individual event loop for
	// each client to simplify things.
	threadedPoll := C.avahi_threaded_poll_new()
	if threadedPoll == nil {
		return nil, errors.New("avahi: not enough memory")
	}

	// Create Avahi client
	var rc C.int
	avahiClient := C.avahi_client_new(
		C.avahi_threaded_poll_get(threadedPoll),
		C.AVAHI_CLIENT_NO_FAIL,
		C.AvahiClientCallback(C.clientCallback),
		nil,
		&rc)

	if avahiClient == nil {
		C.avahi_threaded_poll_free(threadedPoll)
		return nil, fmt.Errorf("avahi: error %d", rc)
	}

	clnt := &Client{
		avahiClient:  avahiClient,
		threadedPoll: threadedPoll,
		evq:          newQueue[ClientState](),
	}

	// Avahi calls callback for the first time very early, even
	// before avahi_client_new() returns. We can't handle it at
	// that time, between the mapping between AvahiClient and Client
	// is not established yet.
	//
	// So some additional effort is required if we don't want to
	// miss the very first state change notification.
	state := ClientState(C.avahi_client_get_state(avahiClient))
	clnt.state = state
	clnt.evq.Push(state)

	clientMap.Put(avahiClient, clnt)

	// And now we finally ready to let AvahiClient run.
	C.avahi_threaded_poll_start(threadedPoll)

	return clnt, nil
}

// Close closes a [Client].
func (clnt *Client) Close() {
	C.avahi_threaded_poll_stop(clnt.threadedPoll)
	clientMap.Del(clnt.avahiClient)
	C.avahi_client_free(clnt.avahiClient)
	C.avahi_threaded_poll_free(clnt.threadedPoll)
	clnt.evq.Close()
}

// Chan returns a channel where [ClientState] change events
// are delivered.
func (clnt *Client) Chan() <-chan ClientState {
	return clnt.evq.Chan()
}

// clientCallback called by AvahiClient to report client state change
//
//export clientCallback
func clientCallback(avahiClient *C.AvahiClient,
	s C.AvahiClientState, _ unsafe.Pointer) {

	clnt := clientMap.Get(avahiClient)
	if clnt == nil {
		// First callback comes even before avahi_client_new()
		// returns, so Client may be not in the map yet. We must
		// handle this case here.
		return
	}

	// As we may loose the very first callback invocation, the very
	// first event is posted "manually". To avoid duplication, here
	// we drop events that doesn't change last known ClientState.
	state := ClientState(s)
	if state != clnt.state {
		clnt.state = state
		clnt.evq.Push(state)
	}
}
