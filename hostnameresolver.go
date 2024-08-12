// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Hostname resolver
//
//go:build linux || freebsd

package avahi

import (
	"context"
	"net/netip"
	"runtime/cgo"
	"sync/atomic"
	"unsafe"
)

// #include <stdlib.h>
// #include <avahi-client/lookup.h>
//
// void hostnameResolverCallback (
//	AvahiHostNameResolver *r,
//	AvahiIfIndex interface,
//	AvahiProtocol proto,
//	AvahiResolverEvent event,
//	char *host_name,
//	AvahiAddress *a,
//	AvahiLookupResultFlags flags,
//	void *userdata);
import "C"

// HostNameResolver resolves hostname by IP address.
type HostNameResolver struct {
	clnt          *Client                            // Owning Client
	handle        cgo.Handle                         // Handle to self
	avahiResolver *C.AvahiHostNameResolver           // Underlying object
	queue         eventqueue[*HostNameResolverEvent] // Event queue
	closed        atomic.Bool                        // Resolver is closed
}

// HostNameResolverEvent represents events, generated by the
// [HostNameResolver].
type HostNameResolverEvent struct {
	Event    ResolverEvent     // Event code
	IfIndex  IfIndex           // Network interface index
	Protocol Protocol          // Network protocol
	Err      ErrCode           // In a case of ResolverFailure
	Flags    LookupResultFlags // Lookup flags
	Hostname string            // Hostname (mirrored)
	Addr     netip.Addr        // IP address (resolved)
}

// NewHostNameResolver creates a new [HostNameResolver].
//
// HostNameResolver resolves IP addresses by provided hostname.
// Roughly speaking, it does the work similar to gethostbyname
// using MDNS. Resolved information is reported via channel
// returned by the [HostNameResolver.Chan].
//
// This is important to understand the proper usage of the "proto"
// and "addrproto" parameters and difference between them.  Please
// read the "IP4 vs IP6" section of the package Overview for technical
// details.
//
// Function parameters:
//   - clnt is the pointer to [Client]
//   - ifindex is the network interface index. Use [IfIndexUnspec]
//     to specify all interfaces.
//   - proto is the IP4/IP6 protocol, used as transport for queries. If
//     set to [ProtocolUnspec], both protocols will be used.
//   - hostname is the name of the host to lookup for
//   - flags provide some lookup options. See [LookupFlags] for details.
//
// HostNameResolver must be closed after use with the [HostNameResolver.Close]
// function call.
func NewHostNameResolver(
	clnt *Client,
	ifindex IfIndex,
	proto Protocol,
	hostname string,
	addrproto Protocol,
	flags LookupFlags) (*HostNameResolver, error) {

	// Initialize HostNameResolver structure
	resolver := &HostNameResolver{clnt: clnt}
	resolver.handle = cgo.NewHandle(resolver)
	resolver.queue.init()

	// Convert strings from Go to C
	chostname := C.CString(hostname)
	defer C.free(unsafe.Pointer(chostname))

	// Create AvahiHostNameResolver
	avahiClient := clnt.begin()
	defer clnt.end()

	resolver.avahiResolver = C.avahi_host_name_resolver_new(
		avahiClient,
		C.AvahiIfIndex(ifindex),
		C.AvahiProtocol(proto),
		chostname,
		C.AvahiProtocol(addrproto),
		C.AvahiLookupFlags(flags),
		C.AvahiHostNameResolverCallback(C.hostnameResolverCallback),
		unsafe.Pointer(&resolver.handle),
	)

	if resolver.avahiResolver == nil {
		resolver.queue.Close()
		resolver.handle.Delete()
		return nil, clnt.errno()
	}

	// Register self to be closed if Client is closed
	resolver.clnt.addCloser(resolver)

	return resolver, nil
}

// Chan returns channel where [HostNameResolverEvent]s are sent.
func (resolver *HostNameResolver) Chan() <-chan *HostNameResolverEvent {
	return resolver.queue.Chan()
}

// Get waits for the next [HostNameResolverEvent].
//
// It returns:
//   - event, nil - if event available
//   - nil, error - if context is canceled
//   - nil, nil   - if HostNameResolver was closed
func (resolver *HostNameResolver) Get(ctx context.Context) (
	*HostNameResolverEvent, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case evnt := <-resolver.Chan():
		return evnt, nil
	}
}

// Close closes the [HostNameResolver] and releases allocated resources.
// It closes the event channel, effectively unblocking pending readers.
//
// Note, double close is safe
func (resolver *HostNameResolver) Close() {
	if !resolver.closed.Swap(true) {
		resolver.clnt.begin()
		resolver.clnt.delCloser(resolver)
		C.avahi_host_name_resolver_free(resolver.avahiResolver)
		resolver.avahiResolver = nil
		resolver.clnt.end()

		resolver.queue.Close()
		resolver.handle.Delete()
	}
}

// hostnameResolverCallback called by AvahiHostNameResolver to
// report discovered services
//
//export hostnameResolverCallback
func hostnameResolverCallback(
	r *C.AvahiHostNameResolver,
	ifindex C.AvahiIfIndex,
	proto C.AvahiProtocol,
	event C.AvahiResolverEvent,
	hostname *C.char,
	caddr *C.AvahiAddress,
	flags C.AvahiLookupResultFlags,
	p unsafe.Pointer) {

	resolver := (*cgo.Handle)(p).Value().(*HostNameResolver)

	// Generate an event
	ip := decodeAvahiAddress(caddr)
	evnt := &HostNameResolverEvent{
		Event:    ResolverEvent(event),
		IfIndex:  IfIndex(ifindex),
		Protocol: Protocol(proto),
		Flags:    LookupResultFlags(flags),
		Hostname: C.GoString(hostname),
		Addr:     ip,
	}

	if evnt.Event == ResolverFailure {
		evnt.Err = resolver.clnt.errno()
	}

	resolver.queue.Push(evnt)
}
