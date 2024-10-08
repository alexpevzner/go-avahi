// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Service resolver
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
// void serviceResolverCallback (
//	AvahiServiceResolver *r,
//	AvahiIfIndex interface,
//	AvahiProtocol proto,
//	AvahiResolverEvent event,
//	char *name,
//	char *type,
//	char *domain,
//	char *host_name,
//	AvahiAddress *a,
//	uint16_t port,
//	AvahiStringList *txt,
//	AvahiLookupResultFlags flags,
//	void *userdata);
import "C"

// ServiceResolver resolves hostname, IP address and TXT record of
// the discovered services.
type ServiceResolver struct {
	clnt          *Client                           // Owning Client
	handle        cgo.Handle                        // Handle to self
	avahiResolver *C.AvahiServiceResolver           // Underlying object
	queue         eventqueue[*ServiceResolverEvent] // Event queue
	closed        atomic.Bool                       // Resolver is closed
}

// ServiceResolverEvent represents events, generated by the
// [ServiceResolver].
//
// Notes:
//   - Addr is not available, if [NewServiceResolver] is called with
//     the [LookupNoAddress] flag
//   - Txt is not available, if [NewServiceResolver] is called with
//     the [LookupNoTXT] flag
//   - Port is always available, but may be 0, which indicates, that
//     service doesn't actually process responses and exists as a
//     service instance name placeholder only.
type ServiceResolverEvent struct {
	Event        ResolverEvent     // Event code
	IfIdx        IfIndex           // Network interface index
	Proto        Protocol          // Network protocol
	Err          ErrCode           // In a case of ResolverFailure
	Flags        LookupResultFlags // Lookup flags
	InstanceName string            // Service instance name (mirrored)
	SvcType      string            // Service type (mirrored)
	Domain       string            // Service domain (mirrored)
	Hostname     string            // Service hostname (resolved)
	Port         uint16            // Service IP port (resolved)
	Addr         netip.Addr        // Service IP address (resolved)
	Txt          []string          // TXT record ("key=value"...) (resolved)
}

// FQDN returns a Fully Qualified Domain Name by joining
// Hostname and Domain.
func (evnt *ServiceResolverEvent) FQDN() string {
	fqdn := evnt.Hostname
	if evnt.Domain != "" {
		fqdn += "." + evnt.Domain
	}
	return fqdn
}

// NewServiceResolver creates a new [ServiceResolver].
//
// ServiceResolver resolves hostname, IP address and TXT record of
// the services, previously discovered by the [ServiceBrowser] by
// service instance name ([ServiceBrowserEvent.InstanceName]).
// Resolved information is reported via channel returned by the
// [ServiceResolver.Chan].
//
// If IP address and/or TXT record is not needed, resolving of these
// parameters may be suppressed, using LookupNoAddress/LookupNoTXT
// [LookupFlags].
//
// Please notice, it is a common practice to register a service
// with a zero port value as a "placeholder" for missed service.
// For example, printers always register the "_printer._tcp" service
// to reserve the service name, but if LPD protocol is actually not
// supported, it will be registered with zero port.
//
// This is important to understand the proper usage of the "proto"
// and "addrproto" parameters and difference between them.  Please
// read the "IP4 vs IP6" section of the package Overview for technical
// details.
//
// Function parameters:
//   - clnt is the pointer to [Client]
//   - ifidx is the network interface index. Use [IfIndexUnspec]
//     to specify all interfaces.
//   - proto is the IP4/IP6 protocol, used as transport for queries. If
//     set to [ProtocolUnspec], both protocols will be used.
//   - instname is the service instance name, as reported by
//     [ServiceBrowserEvent.InstanceName]
//   - svctype is the service type we are looking for (e.g., "_http._tcp")
//   - domain is domain where service is looked. If set to "", the
//     default domain is used, which depends on a avahi-daemon configuration
//     and usually is ".local"
//   - addrproto specifies a protocol family of IP addresses we are
//     interested in. See explanation above for details.
//   - flags provide some lookup options. See [LookupFlags] for details.
//
// ServiceResolver must be closed after use with the [ServiceResolver.Close]
// function call.
func NewServiceResolver(
	clnt *Client,
	ifidx IfIndex,
	proto Protocol,
	instname, svctype, domain string,
	addrproto Protocol,
	flags LookupFlags) (*ServiceResolver, error) {

	// Initialize ServiceResolver structure
	resolver := &ServiceResolver{clnt: clnt}
	resolver.handle = cgo.NewHandle(resolver)
	resolver.queue.init()

	// Convert strings from Go to C
	cinstname := C.CString(instname)
	defer C.free(unsafe.Pointer(cinstname))

	csvctype := C.CString(svctype)
	defer C.free(unsafe.Pointer(csvctype))

	cdomain := C.CString(domain)
	defer C.free(unsafe.Pointer(cdomain))

	// Create AvahiServiceResolver
	avahiClient := clnt.begin()
	defer clnt.end()

	resolver.avahiResolver = C.avahi_service_resolver_new(
		avahiClient,
		C.AvahiIfIndex(ifidx),
		C.AvahiProtocol(proto),
		cinstname, csvctype, cdomain,
		C.AvahiProtocol(addrproto),
		C.AvahiLookupFlags(flags),
		C.AvahiServiceResolverCallback(C.serviceResolverCallback),
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

// Chan returns channel where [ServiceResolverEvent]s are sent.
func (resolver *ServiceResolver) Chan() <-chan *ServiceResolverEvent {
	return resolver.queue.Chan()
}

// Get waits for the next [ServiceResolverEvent].
//
// It returns:
//   - event, nil - if event available
//   - nil, error - if context is canceled
//   - nil, nil   - if ServiceResolver was closed
func (resolver *ServiceResolver) Get(ctx context.Context) (
	*ServiceResolverEvent, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case evnt := <-resolver.Chan():
		return evnt, nil
	}
}

// Close closes the [ServiceResolver] and releases allocated resources.
// It closes the event channel, effectively unblocking pending readers.
func (resolver *ServiceResolver) Close() {
	if !resolver.closed.Swap(true) {
		resolver.clnt.begin()
		resolver.clnt.delCloser(resolver)
		C.avahi_service_resolver_free(resolver.avahiResolver)
		resolver.avahiResolver = nil
		resolver.clnt.end()

		resolver.queue.Close()
		resolver.handle.Delete()
	}
}

// serviceResolverCallback called by AvahiServiceResolver to
// report discovered services
//
//export serviceResolverCallback
func serviceResolverCallback(
	r *C.AvahiServiceResolver,
	ifidx C.AvahiIfIndex,
	proto C.AvahiProtocol,
	event C.AvahiResolverEvent,
	name, svctype, domain, hostname *C.char,
	caddr *C.AvahiAddress,
	cport C.uint16_t,
	ctxt *C.AvahiStringList,
	flags C.AvahiLookupResultFlags,
	p unsafe.Pointer) {

	resolver := (*cgo.Handle)(p).Value().(*ServiceResolver)
	clnt := resolver.clnt

	// Decode IP address:port
	ip := decodeAvahiAddress(IfIndex(ifidx), caddr)

	// Decode TXT record
	txt := decodeAvahiStringList(ctxt)

	// Generate an event
	evnt := &ServiceResolverEvent{
		Event:        ResolverEvent(event),
		IfIdx:        IfIndex(ifidx),
		Proto:        Protocol(proto),
		Flags:        LookupResultFlags(flags),
		InstanceName: C.GoString(name),
		SvcType:      C.GoString(svctype),
		Domain:       C.GoString(domain),
		Hostname:     C.GoString(hostname),
		Addr:         ip,
		Port:         uint16(cport),
		Txt:          txt,
	}

	// If host is connected to the internet, Avahi erroneously
	// uses a real host name and domain instead of localhost.localdomain.
	//
	// Fix it here.
	if clnt.hasFlags(ClientLoopbackWorkarounds) && ip.IsLoopback() {
		evnt.Hostname = "localhost"
		evnt.Domain = "localdomain"
	}

	if evnt.Event == ResolverFailure {
		evnt.Err = clnt.errno()
	}

	resolver.queue.Push(evnt)
}
