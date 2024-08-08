// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Address resolver
//
//go:build linux || freebsd

package avahi

import (
	"context"
	"net/netip"
	"runtime/cgo"
	"unsafe"
)

// #include <stdlib.h>
// #include <avahi-client/lookup.h>
//
// void addressResolverCallback (
//	AvahiAddressResolver *r,
//	AvahiIfIndex interface,
//	AvahiProtocol proto,
//	AvahiResolverEvent event,
//	AvahiAddress *a,
//	char *host_name,
//	AvahiLookupResultFlags flags,
//	void *userdata);
import "C"

// AddressResolver resolves hostname by IP address.
type AddressResolver struct {
	clnt          *Client                           // Owning Client
	handle        cgo.Handle                        // Handle to self
	avahiResolver *C.AvahiAddressResolver           // Underlying object
	queue         eventqueue[*AddressResolverEvent] // Event queue
}

// AddressResolverEvent represents events, generated by the
// [AddressResolver].
type AddressResolverEvent struct {
	Event    ResolverEvent     // Event code
	IfIndex  IfIndex           // Network interface index
	Protocol Protocol          // Network protocol
	Err      ErrCode           // In a case of ResolverFailure
	Flags    LookupResultFlags // Lookup flags
	Hostname string            // Resolved hostname
}

// NewAddressResolver creates a new [AddressResolver].
//
// AddressResolver resolves hostname by provided IP address.
// Roughly speaking, it does the work similar to gethostbyaddr
// using MDNS. Resolved information is reported via channel
// returned by the [AddressResolver.Chan].
//
// Function parameters:
//   - clnt is the pointer to [Client]
//   - ifindex is the network interface index. Use [IfIndexUnspec]
//     to specify all interfaces.
//   - proto is the IP4/IP6 protocol, used as transport for queries. If
//     set to [ProtocolUnspec], both protocols will be used.
//   - addr is the IP address for which hostname discovery is performed.
//   - flags provide some lookup options. See [LookupFlags] for details.
//
// AddressResolver must be closed after use with the [AddressResolver.Close]
// function call.
func NewAddressResolver(
	clnt *Client,
	ifindex IfIndex,
	proto Protocol,
	addr netip.Addr,
	flags LookupFlags) (*AddressResolver, error) {

	// Initialize AddressResolver structure
	resolver := &AddressResolver{clnt: clnt}
	resolver.handle = cgo.NewHandle(resolver)
	resolver.queue.init()

	// Convert address to AvahiAddress
	var caddr C.AvahiAddress
	addr = addr.Unmap()

	switch {
	case addr.Is4():
		caddr.proto = C.AVAHI_PROTO_INET
		(*(*[4]byte)(unsafe.Pointer(&caddr.data))) = addr.As4()
	case addr.Is6():
		caddr.proto = C.AVAHI_PROTO_INET6
		(*(*[16]byte)(unsafe.Pointer(&caddr.data))) = addr.As16()
	default:
		return nil, ErrInvalidAddress
	}

	// Create AvahiAddressResolver
	avahiClient := clnt.begin()
	defer clnt.end()

	resolver.avahiResolver = C.avahi_address_resolver_new(
		avahiClient,
		C.AvahiIfIndex(ifindex),
		C.AvahiProtocol(proto),
		&caddr,
		C.AvahiLookupFlags(flags),
		C.AvahiAddressResolverCallback(C.addressResolverCallback),
		unsafe.Pointer(&resolver.handle),
	)

	if resolver.avahiResolver == nil {
		resolver.queue.Close()
		resolver.handle.Delete()
		return nil, clnt.errno()
	}

	return resolver, nil
}

// Chan returns channel where [AddressResolverEvent]s are sent.
func (resolver *AddressResolver) Chan() <-chan *AddressResolverEvent {
	return resolver.queue.Chan()
}

// Get waits for the next [AddressResolverEvent].
//
// It returns:
//   - event, nil - if event available
//   - nil, error - if context is canceled
//   - nil, nil   - if AddressResolver was closed
func (resolver *AddressResolver) Get(ctx context.Context) (
	*AddressResolverEvent, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case evnt := <-resolver.Chan():
		return evnt, nil
	}
}

// Close closes the [AddressResolver] and releases allocated resources.
// It closes the event channel, effectively unblocking pending readers.
func (resolver *AddressResolver) Close() {
	resolver.clnt.begin()
	C.avahi_address_resolver_free(resolver.avahiResolver)
	resolver.avahiResolver = nil
	resolver.clnt.end()

	resolver.queue.Close()
	resolver.handle.Delete()
}

// addressResolverCallback called by AvahiAddressResolver to
// report resolved hostnames.
//
//export addressResolverCallback
func addressResolverCallback(
	r *C.AvahiAddressResolver,
	ifindex C.AvahiIfIndex,
	proto C.AvahiProtocol,
	event C.AvahiResolverEvent,
	caddr *C.AvahiAddress,
	hostname *C.char,
	flags C.AvahiLookupResultFlags,
	p unsafe.Pointer) {

	resolver := (*cgo.Handle)(p).Value().(*AddressResolver)

	// Generate an event
	evnt := &AddressResolverEvent{
		Event:    ResolverEvent(event),
		IfIndex:  IfIndex(ifindex),
		Protocol: Protocol(proto),
		Err:      resolver.clnt.errno(),
		Flags:    LookupResultFlags(flags),
		Hostname: C.GoString(hostname),
	}

	resolver.queue.Push(evnt)
}