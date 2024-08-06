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
	"fmt"
	"runtime/cgo"
	"unsafe"
)

// #include <avahi-client/client.h>
// #include <avahi-common/thread-watch.h>
//
// void clientCallback (AvahiClient*, AvahiClientState, void*);
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
	handle       cgo.Handle              // Handle to self
	avahiClient  *C.AvahiClient          // Underlying AvahiClient
	threadedPoll *C.AvahiThreadedPoll    // Avahi event loop
	queue        eventqueue[ClientState] // Event queue
}

// NewClient creates a new [Client].
func NewClient() (*Client, error) {
	// Create Avahi event loop. We use individual event loop for
	// each client to simplify things.
	threadedPoll := C.avahi_threaded_poll_new()
	if threadedPoll == nil {
		return nil, ErrNoMemory
	}

	// Create Avahi client
	clnt := &Client{threadedPoll: threadedPoll}

	clnt.handle = cgo.NewHandle(clnt)
	clnt.queue.init()

	var rc C.int
	clnt.avahiClient = C.avahi_client_new(
		C.avahi_threaded_poll_get(threadedPoll),
		C.AVAHI_CLIENT_NO_FAIL,
		C.AvahiClientCallback(C.clientCallback),
		unsafe.Pointer(&clnt.handle),
		&rc)

	if clnt.avahiClient == nil {
		C.avahi_threaded_poll_free(threadedPoll)
		clnt.handle.Delete()
		return nil, fmt.Errorf("avahi: error %d", rc)
	}

	// And now we finally ready to let AvahiClient run.
	C.avahi_threaded_poll_start(threadedPoll)

	return clnt, nil
}

// Close closes a [Client].
func (clnt *Client) Close() {
	C.avahi_threaded_poll_stop(clnt.threadedPoll)
	C.avahi_client_free(clnt.avahiClient)
	C.avahi_threaded_poll_free(clnt.threadedPoll)
	clnt.queue.Close()
	clnt.handle.Delete()
}

// Chan returns a channel where [ClientState] change events
// are delivered.
//
// Client.Close closes the sending side of this channel, effectively
// unblocking pending receivers. Once Client is closed, any attempt
// to read from this channel will return [ClientStateClosed] value.
func (clnt *Client) Chan() <-chan ClientState {
	return clnt.queue.Chan()
}

// clientCallback called by AvahiClient to report client state change
//
//export clientCallback
func clientCallback(avahiClient *C.AvahiClient,
	s C.AvahiClientState, p unsafe.Pointer) {

	clntHandle := *(*cgo.Handle)(p)
	clnt := clntHandle.Value().(*Client)

	clnt.queue.Push(ClientState(s))
}
