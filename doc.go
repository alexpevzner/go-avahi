// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Package documentation
//
//go:build linux || freebsd

/*
Package avahi provides a fairly complete CGo binding for [Avahi] client.

Avahi is the standard implementation of Multicast DNS and DNS-SD for Linux, and
likely for some BSD systems as well. This technology is essential for automatic
network configuration, service discovery on local networks, and driverless
printing and scanning.

Please notice, there is an alternative Avahi binding for Go:

  - GitHub project: https://github.com/holoplot/go-avahi
  - The documentation: https://pkg.go.dev/github.com/holoplot/go-avahi

This package has the following key differences:

  - This is CGo binding, not pure Go
  - It uses native/stdlib types, where appropriate. For example,
    IP addresses returned as [netip.AddrPort]
  - It uses a single channel for all events reported by an object,
    so add/remove events cannot be reordered
  - It survives Avahi restart
  - Integer values, like various flags, DNS class and type and
    so own, have their own type, not a generic int16/int32
  - And the last but not least, it attempts to fill the gaps
    in Avahi documentation, which is not very detailed

There is also a pure Go DNS library:

  - GitHub project: https://github.com/miekg/dns
  - The documentation: https://pkg.go.dev/github.com/miekg/dns

This library is comprehensive, high-quality, and quite popular. It is possible
(and not very difficult) to implement MDNS/DNS-SD directly on top of it,
allowing the entire protocol to run within the user process without relying on
a system daemon like Avahi.

There are several existing implementations; however, I don't have experience
with them, so I can't provide a review.

One inherent disadvantage of all these implementations is that they do not work
with local services operating via the loopback network interface. MDNS is a
multicast-based protocol, and the loopback interface does not support
multicasting. System daemons like Avahi do not actually use multicasting for
loopback services; instead, they emulate the publishing and discovery
functionality for those services. An in-process implementation cannot achieve
this.

# Key objects

The key objects exposed by this package are:

  - [Client] represents a client connection to the avahi-daemon
  - Assortment of browsers: [DomainBrowser], [RecordBrowser],
    [ServiceBrowser], [ServiceTypeBrowser]
  - Assortment of resolvers: [AddressResolver], [HostNameResolver],
    [ServiceResolver]

These objects have 1:1 relations to the corresponding avahi objects
(i.e., Client represents AvahiClient, DomainBrowser represennts
AvahiDomainBrowser and so on).

These objects are explicitly created with appropriate constructor
functions (e.g., [NewClient], [NewDomainBrowser], [NewServiceResolver]
and so on).

All these objects report their state change and discovered information
using provided channel (use Chan() method to obtain the channel). There
is also a [context.Context]-aware Get() methods which can be used to
wait for the next event.

As these objects own some resources, such as DBus connection to the
avahi-daemon, which is not automatically released when objects are
garbage-collected, this is important to call appropriate Close
method, when object is not longer in use.

Once object is closed, the sending side of its event channel is closed
too, which effectively unblocks all users waiting for events.

# Client

The [Client] represents a client connection to the avahi-daemon.
Client is the required parameter for creation of Browsers and Resolvers
and "owns" these objects.

Client has a state and this state can change dynamically. Changes in
the Client state reported as a series of [ClientState] events, reported
via the [Client.Chan] channel or [Client.Get] convenience wrapper.

The Client itself can survive avahi-daemon (and DBus server) failure
and restart. If it happens, [ClientStateFailure] event will be reported,
followed by [ClientStateConnecting] and filanny [ClientStateRunning] events
when client connection will be recovered. However, all Browsers and Resolvers
owned by the Client will fail (with [BrowserFailure]/[ResolverFailure]
events) and will not be restarted automatically.

The Client manages underlying AvahiPoll object (Avahi event loop) automatically
and doesn't expose it via its interface.

# Browsers

Browser constantly monitors the network for newly discovered or removed
objects of the specified type and report discovered information as a
series of events, delivered via provided channel.

More technically, browser monitors the network for reception of the
MDNS messages of the browser-specific type and reports these messages
as browser events.

There are 5 types of browser events, represented as values of the
[BrowserEvent] integer type:
  - [BrowserNew] - new object was discovered on a network
  - [BrowserRemove] - the object was removed from the network
  - [BrowserCacheExhausted] - one-time hint event, that notifies the user
    that all entries from the avahi-daemon cache have been sent
  - [BrowserAllForNow] - one-time hint event, that notifies the user that
    more events are are unlikely to be shown in the near feature
  - [BrowserFailure] - browsing failed and needs to be restarted

Avahi documentation doesn't explain in detail, when [BrowserAllForNow]
is generated, but generally, it is generated after an one-second interval
from the reception of MDNS message of related type has been expired.

Each browser has a constructor function (e.g., [NewDomainBrowser]) and
three methods:
  - Chan, which returns the event channel
  - Get, the convenience wrapper which waits for the next event
    and can be canceled using [context.Context] parameter
  - Close, which closes the browser.

This is important to call Close method when browser is not longer in use.

# Resolvers

Resolver performs a series of appropriate MDNS queries to resolve
supplied parameters into the requested information, depending on Resolver
type (e.g,, ServiceResolver will resolve service name into hostname,
IP address:port and TXT record).

Like Browsers, Resolvers return discovered information as a series of
resolver events.

There are 2 types of resolver events, represented by integer value
of the [ResolverEvent] type:
  - [ResolverFound] - new portion of required information received
    from the network
  - [ResolverFailure] - resolving failed and needs to be restarted

Please notice a single query may return multiple [ResolverFound] events.
For example, if target has multiple IP addresses, each address will be
reported via separate event.

Unlike the Browser, the Resolver does not provide any indication of
which event is considered "last" in the sequence. Technically, there is
no definitive "last" event, as a continuously running Resolver will
generate a [ResolverFound] event each time the service data changes.
However, if we simply need to connect to a discovered service, we must
eventually stop waiting. A reasonable approach would be to wait for a
meaningful duration (for example, 1 second) after the last event in the
sequence arrives.

# IP4 vs IP6

When new Browser or Resolver is created, the 3rd parameter of constructor
function specified a transport protocol, used for queries.

Some Resolver constructors have a second parameter of the [Protocol]
type, the "addrproto" parameter. This parameter specifies which kind
of addresses, IP4 or IP6, we are interested in output (technically,
which kind of address records, A or AAAA, are queried).

If you create a Browser, using [ProtocolUnspec] transport protocol, it will
report both IP4 and IP6 RRs and report them as separate events.

A new Resolver, created with [ProtocolUnspec] transport protocol will
use IP6 as its transport protocol, as if [ProtocolIP6] was specified.

If "addrproto" is specified as [ProtocolUnspec], Resolver will always
query for addresses that match the transport protocol.

It can be summarized by the following table:

	proto		addrproto	transport	query for

	ProtocolIP4	ProtocolIP4	IP4		IP4
	ProtocolIP4	ProtocolIP6	IP4		IP6
	ProtocolIP4	ProtocolUnspec	IP4		IP4

	ProtocolIP6	ProtocolIP4	IP6		IP4
	ProtocolIP6	ProtocolIP6	IP6		IP6
	ProtocolIP6	ProtocolUnspec	IP6		IP6

	ProtocolUnspec	ProtocolIP4	IP6		IP4
	ProtocolUnspec	ProtocolIP6	IP6		IP6
	ProtocolUnspec	ProtocolUnspec	IP6		IP6

By default the Avahi daemon publishes both IP4 and IP6 addresses when
queried over IP4, but only IP6 addresses, when queried over IP6. This
default can be changed using 'publish-aaaa-on-ipv4' and
'publish-a-on-ipv6' in 'avahi-daemon.conf').

Other servers (especially DNS-SD servers found on devices, like printers
or scanners) may have a different, sometimes surprising, behavior.

So it makes sense to perform queries of all four transport/address
combinations and merge results.

[Avahi]: https://avahi.org/
*/
package avahi
