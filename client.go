// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Avahi Client
//
//go:build linux || freebsd

package avahi

import (
	"context"
	"fmt"
	"runtime/cgo"
	"sync/atomic"
	"unsafe"
)

// #include <avahi-client/client.h>
// #include <avahi-common/thread-watch.h>
//
// void clientCallback (AvahiClient*, AvahiClientState, void*);
import "C"

// Client represents a client connection to the Avahi daemon.
//
// Client may change its state dynamically. [ClientState] changes
// reported as a series of [ClientEvent] via the channel returned
// by the [Client.Chan] call.
//
// When Client is not in use anymore, it must be closed using the
// Client.Close call to free associated resources. Closing the client
// closes its event notifications channel, effectively unblocking
// pending readers.
type Client struct {
	flags        ClientFlags              // Client creation flags
	handle       cgo.Handle               // Handle to self
	avahiClient  *C.AvahiClient           // Underlying AvahiClient
	threadedPoll *C.AvahiThreadedPoll     // Avahi event loop
	queue        eventqueue[*ClientEvent] // Event queue
	children     closers                  // Children objects
	closed       atomic.Bool              // Client is closed
}

// ClientFlags modify certain aspects of the Client behavior.
type ClientFlags int

// ClientFlags bits:
const (
	// Loopback handling in Avahi is broken. In particular:
	//   - AvahiServiceResolver and AvahiAddressResolver
	//     return real host name and domain for loopback addresses
	//   - AvahiHostNameResolver doesn't resolve neither
	//     "localhost" nor "localhost.localdomain".
	//
	// Among other things, it breaks IPP over USB support. This
	// protocol uses Avahi for local service discovery and has
	// a strong requirement to use "localhost" as a hostname
	// when working with local services.
	//
	// If Client is created with this flag, the following
	// workarounds are enabled:
	//   - Host name and domain returned by ServiceResolver and
	//     AddressResolver for the loopback addresses (127.0.0.1
	//     and ::1) are forced to be localhost.localdomain
	//   - HostNameResolver resolves "localhost" and
	//     "localhost.localdomain" as 127.0.0.1.
	ClientLoopbackWorkarounds ClientFlags = 1 << iota
)

// ClientEvent represents events, generated by the [Client].
type ClientEvent struct {
	State ClientState // New client state
	Err   ErrCode     // Only for ClientStateFailure
}

// NewClient creates a new [Client].
func NewClient(flags ClientFlags) (*Client, error) {
	// Create Avahi event loop. We use individual event loop for
	// each client to simplify things.
	threadedPoll := C.avahi_threaded_poll_new()
	if threadedPoll == nil {
		return nil, ErrNoMemory
	}

	// Create Avahi client
	clnt := &Client{flags: flags, threadedPoll: threadedPoll}

	clnt.handle = cgo.NewHandle(clnt)
	clnt.queue.init()
	clnt.children.init()

	var rc C.int
	clnt.avahiClient = C.avahi_client_new(
		C.avahi_threaded_poll_get(threadedPoll),
		C.AVAHI_CLIENT_NO_FAIL,
		C.AvahiClientCallback(C.clientCallback),
		unsafe.Pointer(&clnt.handle),
		&rc)

	if clnt.avahiClient == nil {
		C.avahi_threaded_poll_free(threadedPoll)
		clnt.queue.Close()
		clnt.handle.Delete()
		return nil, fmt.Errorf("avahi: error %d", rc)
	}

	// And now we finally ready to let AvahiClient run.
	C.avahi_threaded_poll_start(threadedPoll)

	return clnt, nil
}

// Close closes a [Client].
//
// Note, double close is safe.
func (clnt *Client) Close() {
	if !clnt.closed.Swap(true) {
		C.avahi_threaded_poll_stop(clnt.threadedPoll)

		clnt.children.close()

		C.avahi_client_free(clnt.avahiClient)
		clnt.avahiClient = nil

		C.avahi_threaded_poll_free(clnt.threadedPoll)
		clnt.threadedPoll = nil

		clnt.queue.Close()
		clnt.handle.Delete()
	}
}

// addCloser adds a child object that will be closed when client is closed
func (clnt *Client) addCloser(obj closer) {
	clnt.children.add(obj)
}

// delCloser deletes a child object
func (clnt *Client) delCloser(obj closer) {
	clnt.children.del(obj)
}

// Chan returns a channel where [ClientState] change events
// are delivered.
//
// Client.Close closes the sending side of this channel, effectively
// unblocking pending receivers.
func (clnt *Client) Chan() <-chan *ClientEvent {
	return clnt.queue.Chan()
}

// Get waits for the next [ClientEvent].
//
// It returns:
//   - event, nil - on success
//   - nil, error - if context is canceled
//   - nil, nil   - if Client was closed
func (clnt *Client) Get(ctx context.Context) (*ClientEvent, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case state := <-clnt.Chan():
		return state, nil
	}
}

// GetVersionString returns avahi-daemon version string
func (clnt *Client) GetVersionString() string {
	clnt.begin()
	defer clnt.end()

	s := C.avahi_client_get_version_string(clnt.avahiClient)
	return C.GoString(s)
}

// GetHostName returns host name (e.g., "name")
func (clnt *Client) GetHostName() string {
	clnt.begin()
	defer clnt.end()

	s := C.avahi_client_get_host_name(clnt.avahiClient)
	return C.GoString(s)
}

// GetDomainName returns domain name (e.g., "local")
func (clnt *Client) GetDomainName() string {
	clnt.begin()
	defer clnt.end()

	s := C.avahi_client_get_domain_name(clnt.avahiClient)
	return C.GoString(s)
}

// GetHostFQDN returns FQDN host name (e.g., "name.local")
func (clnt *Client) GetHostFQDN() string {
	clnt.begin()
	defer clnt.end()

	s := C.avahi_client_get_host_name_fqdn(clnt.avahiClient)
	return C.GoString(s)
}

// begin locks the Client event loop and returns *C.AvahiClient.
//
// All operations that affects underlying AvahiClient must begin
// with this call.
//
// Caller MUST call Client.end after end of operation.
func (clnt *Client) begin() *C.AvahiClient {
	C.avahi_threaded_poll_lock(clnt.threadedPoll)
	return clnt.avahiClient
}

// end must be called after completion of any operation, started
// with Client.begin.
func (clnt *Client) end() {
	C.avahi_threaded_poll_unlock(clnt.threadedPoll)
}

// hasFlags checks if some of the specified flags were used during
// the Client creation.
func (clnt *Client) hasFlags(flags ClientFlags) bool {
	return clnt.flags&flags != 0
}

// errno returns an error code of latest failed operation.
func (clnt *Client) errno() ErrCode {
	return ErrCode(C.avahi_client_errno(clnt.avahiClient))
}

// clientCallback called by AvahiClient to report client state change
//
//export clientCallback
func clientCallback(avahiClient *C.AvahiClient,
	s C.AvahiClientState, p unsafe.Pointer) {

	clntHandle := *(*cgo.Handle)(p)
	clnt := clntHandle.Value().(*Client)

	state := ClientState(s)
	evnt := &ClientEvent{State: state}

	if state == ClientStateFailure {
		// The very first callback may come too early, even
		// before C.avahi_client_new returns, so Client.avahiClient
		// may be not yet initialized at that time...
		evnt.Err = ErrFailure
		if clnt.avahiClient != nil {
			evnt.Err = clnt.errno()
		}
	}

	clnt.queue.Push(evnt)
}
