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

# Package philosophy

The Avahi API wrapper, provided by this package, attempts to be as close
to the original Avahi C API and as transparent, as possible. However,
the following differences still exist:
  - Events are reported via channels, not via callbacks, as in C
  - AvahiPoll object is not exposed and handled internally
  - Workaround for Avahi localhost handling bug is provides (for details,
    see "Loopback interface handling and localhost" section above).

# A bit of theory (Multicast DNS and DNS-SD essentials)

Avahi API is much simpler to understand when reader knows the basics
of the Multicast DNS and DNS-SD protocols.

DNS is a kind of a distributed key-value database. In classical (unicast)
DNS, records are maintained by the hierarchy of servers, while in the
MDNS each participating host maintains its own records by itself and
responds when somebody asks, usually using multicast UDP as transport.

In the case of classical DNS, clients perform database queries by
contacting DNS servers. In contrast, with multicast DNS, clients send
their queries to all other hosts in the vicinity using UDP multicast
(e.g., "Hey! I need an IP address for the 'example.local' hostname. Who
knows the answer?"). The hosts then respond by their own. To speed things
up, when a new host connects to the network, it announces its resource
records (RRs) to all interested parties and attempts to notify others of
its removal just before it disconnects. Clients can capture and cache
this information, eliminating the need for a slow network query each
time this information is requires.

Each entry in the DNS database (called the Resource Record, RR) is
identified by the search key, which consist of:
  - record name
  - record class
  - record type

The record name always looks like domain name, i.e., it is a string that
consists from the dot-separated labels. The "example.com" name consists of
two labels: "example" and "com".

This syntax is used even for names, which are not domains by themselves.
For example, "1.0.0.127.in-addr.arpa" is the IP address "127.0.0.1", written
using a DNS name syntax (please notice the reverse order of labels),
and "_http._tcp.local" is the collective name of all HTTP servers running
over TCP on a local network.

To distinguish between normal domains and pseudo-domains, a number of special
top-level domains have been reserved for this purpose, like "in-addr.arpa"
for IP addresses

DNS defines many classes, but the only class relevant to multicast DNS is IN,
which stands for "Internet." That's all there is to it.

Record type is more important, as many record types are being used.
We will not attempt to list them all, the most important for as are
the following:

	A	- these records contains one or more IPv4 addresses
	AAAA	- these records contains one or more IPv6 addresses
	PTR	- the pointer record. They point to some other domain name
	SRV	- service descriptor
	TXT	- contains a lot of additional information, represented
		  as a list of key=value textual pairs.

Once we have a record name and type, we can query a record value.
Interpretation of this value depends on a record type.

Now lets manually discover all IPP printers in our local network.
We will use the small utility, [mcdig], which allows to manually
perform the Multicast DNS queries.

First of all, lets list all services on a network around. This is
a query of the "_services._dns-sd._udp.local" records of type PTR,
and [mcdig] will return the following answer (shortened):

	$ mcdig _services._dns-sd._udp.local ptr
	;; ANSWER SECTION:
	_services._dns-sd._udp.local.	4500	IN	PTR	_http._tcp.local.
	_services._dns-sd._udp.local.	4500	IN	PTR	_https._tcp.local.
	_services._dns-sd._udp.local.	4500	IN	PTR	_ipp._tcp.local.
	_services._dns-sd._udp.local.	4500	IN	PTR	_ipps._tcp.local.
	_services._dns-sd._udp.local.	4500	IN	PTR	_printer._tcp.local.

This is the same list as avahi-browse -a returns, and programmatically
it can be obtained, using the [ServiceTypeBrowser] object.

Please notice, the "_services._dns-sd._udp.<domain>" is a reserved
name for this purpose and <domain> is usually "local"; this top-level
domain name is reserved for this purpose.

Now we see that somebody in our network provide the "_http._tcp.local."
service (IPP printing), "_http._tcp.local." service (HTTP server) and
so on. In a typical network there will be many services and they will
duplicate in the answer.

Now, we are only interested in the IPP printers, so:

	$ mcdig _ipp._tcp.local. ptr
	;; ANSWER SECTION:
	_ipp._tcp.local.	4500	IN	PTR	Kyocera\ ECOSYS\ M2040dn._ipp._tcp.local.

Now we have a so called service instance name, "Kyocera ECOSYS M2040dn".
Please notice, unlike classical DNS, MDNS labels may contain spaces (and
virtually any valid UTF-8 characters), but among these labels looks
like human-readable names, they are network-unique (which is enforced
by the protocol) and can be used to unambiguously identify the device.

The same list will be returned by the avahi-browse _ipp._tcp command
(please notice, the .local suffix is implied here) or using the
[ServiceBrowser] object.

Now we need to know a bit more about the device, so the next query is:

	$ mcdig Kyocera\ ECOSYS\ M2040dn._ipp._tcp.local. any
	Kyocera\ ECOSYS\ M2040dn._ipp._tcp.local.	120	IN	SRV	0 0 631 KM7B6A91.local.
	KM7B6A91.local.					120	IN	A	192.168.1.102
	KM7B6A91.local.					120	IN	AAAA	fe80::217:c8ff:fe7b:6a91

The response is really huge and significantly shortened here. The TXT
record is omitted at all, as it really large.

The important records are:
  - A and AAAA brings us IP addresses of the device
  - SRV record gives us a hostname (which is not the same as the
    instance name, and is not as friendly and human-readable)
    and IP port (631, the third parameter in the SRV RR)
  - TXT record, which brings a lot of additional information,
    like duplex support ("Duplex=T"), root path for the HTTP
    requests ("rp=ipp/print"; IPP is the HTTP-based protocol),
    list of supported documents formats and much more.

The same information can be obtained programmatically, using the
[ServiceResolver] object.

And finally, we can lookup IP address by hostname and hostname by IP address:

	$ mcdig KM7B6A91.local. a
	;; ANSWER SECTION:
	KM7B6A91.local.			120	IN	A	192.168.1.102

	$ mcdig 102.1.168.192.in-addr.arpa ptr
	;; ANSWER SECTION:
	102.1.168.192.in-addr.arpa.	120	IN	PTR	KM7B6A91.local.

It corresponds to avahi commands "avahi-resolve-host-name KM7B6A91.local" and
"avahi-resolve-address 192.168.1.102".

The [HostNameResolver] and [AddressResolver] objects provide the similar
functionality in a form of API.

# Key objects

The key objects exposed by this package are:

  - [Client] represents a client connection to the avahi-daemon
  - Assortment of browsers: [DomainBrowser], [RecordBrowser],
    [ServiceBrowser], [ServiceTypeBrowser]
  - Assortment of resolvers: [AddressResolver], [HostNameResolver],
    [ServiceResolver]
  - [EntryGroup], which implements Avahi publishing API.

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
the Client state reported as a series of [ClientEVENT] events, reported
via the [Client.Chan] channel or [Client.Get] convenience wrapper.

The Client itself can survive avahi-daemon (and DBus server) failure
and restart. If it happens, [ClientStateFailure] event will be reported,
followed by [ClientStateConnecting] and finally [ClientStateRunning],
when client connection will be recovered. However, all Browsers, Resolvers
and [EntryGroup]-s owned by the Client will fail (with
[BrowserFailure]/[ResolverFailure]/[EntryGroupStateFailure] events) and
will not be restarted automatically. If it happens, application needs
to close and re-create these objects.

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

# EntryGroup

[EntryGroup] implements Avahi publishing API. This is, essentially,
a collection of resource entries which can be published "atomically",
i.e., either the whole group is published or not.

Records can be added to the EntryGroup using [EntryGroup.AddService],
[EntryGroup.AddAddress] and [EntryGroup.AddRecord] methods. Existing
services can be modified, using the [EntryGroup.AddServiceSubtype] and
[EntryGroup.UpdateServiceTxt] methods. Once group is configured,
application must call [EntryGroup.Commit] for changes to take effect.

When records are added, even before Commit, Avahi performs some basic
checking of the group consistency, and if consistency is violated or
added records contains invalid data, the appropriate call will fail
with suitable error code.

When publishing services, there is no way to set service IP address
explicitly. Instead, Avahi deduces appropriate IP address, based on
the network interface being used and available addresses assigned
to that interface.

Like other objects, EntryGroup maintains a dynamic state and reports
its state changes using [EntryGroupEvent] which can be received either
via the channel, returned by [EntryGroup.Chan] or via the
[EntryGroup.Get] convenience wrapper.

As the protocol requires, EntryGroup implies a conflict checking,
so this process takes some time. As result of this process, the
EntryGroup will eventually come into the either EntryGroupStateEstablished
or EntryGroupStateCollision state.

Unfortunately, in a case of collision there is no detailed reporting,
which entry has caused a collision. So it is not recommended to mix
unrelated entries in the same group.

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

# Loopback interface handling and localhost

As loopback network interface doesn't support multicasting, Avahi
just emulates the appropriate functionality.

Loopback support is essentially for implementing the [IPP over USB]
protocol, and [ipp-usb] daemon actively uses it. It allows the many
modern printers and scanners to work seamlessly under the Linux OS.

Unfortunately, loopback support is broken in Avahi. This is a long
story, but in short:
  - Services, published at the loopback address (127.0.0.1 or ::1)
    are erroneously reported by AvahiServiceResolver as being
    published at the real hostname and domain, instead of
    "localhost.localdomain"
  - AvahiAddressResolver also resolves these addresses using
    real hostname and domain
  - AvahiHostNameResolver doesn't resolve neither "localhost" nor
    "localhost.localdomain".

This library provides a workaround, but it needs to be explicitly
enabled, using the [ClientLoopbackWorkarounds] flag:

	clnt, err := NewClient(ClientLoopbackWorkarounds)

If this flag is in use, the following changes will occur:
  - [ServiceResolver] and [AddressResolver] will return "localhost.localdomain"
    for the loopback addresses
  - [HostNameResolver] will resolve "localhost" and "localhost.localdomain"
    as either 127.0.0.1 or ::1, depending on a value of the
    proto parameter for the [NewHostNameResolver] call. Please notice that
    if proto is [ProtocolUnspec], NewHostNameResolver will use by
    default [ProtocolIP6], to be consistent with other Avahi API
    (see section "IP4 vs IP6" for details).

[Avahi]: https://avahi.org/
[IPP over USB]: https://www.usb.org/document-library/ipp-protocol-10
[ipp-usb]: https://github.com/OpenPrinting/ipp-usb
[mcdig]: https://github.com/alexpevzner/mcdig
*/
package avahi
