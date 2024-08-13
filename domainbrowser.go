// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Service browser
//
//go:build linux || freebsd

package avahi

import (
	"context"
	"runtime/cgo"
	"sync/atomic"
	"unsafe"
)

// #include <stdlib.h>
// #include <avahi-client/lookup.h>
//
// void domainBrowserCallback (
//	AvahiDomainBrowser *b,
//	AvahiIfIndex interface,
//	AvahiProtocol proto,
//	AvahiBrowserEvent event,
//	char *name,
//	char *type,
//	char *domain,
//	AvahiLookupResultFlags flags,
//	void *userdata);
import "C"

// DomainBrowser performs discovery of browsing and registration
// domains. See [NewDomainBrowser] and [RFC6763, 11] for details.
//
// [RFC6763, 11]: https://datatracker.ietf.org/doc/html/rfc6763#section-11
type DomainBrowser struct {
	clnt         *Client                         // Owning Client
	handle       cgo.Handle                      // Handle to self
	avahiBrowser *C.AvahiDomainBrowser           // Underlying object
	queue        eventqueue[*DomainBrowserEvent] // Event queue
	closed       atomic.Bool                     // Browser is closed
}

// DomainBrowserType specifies a type of domain to browse for.
type DomainBrowserType int

// DomainBrowserType values:
const (
	// Request list of available browsing domains.
	DomainBrowserBrowse DomainBrowserType = C.AVAHI_DOMAIN_BROWSER_BROWSE

	// Request the default browsing domain.
	DomainBrowserBrowseDefault DomainBrowserType = C.AVAHI_DOMAIN_BROWSER_BROWSE_DEFAULT

	// Request list of available registering domains.
	DomainBrowserRegister DomainBrowserType = C.AVAHI_DOMAIN_BROWSER_REGISTER

	// Request the default registering domains.
	DomainBrowserRegisterDefault DomainBrowserType = C.AVAHI_DOMAIN_BROWSER_REGISTER_DEFAULT

	// Request for "legacy browsing" domains. See RFC6763, 11 for details.
	DomainBrowserLegacy DomainBrowserType = C.AVAHI_DOMAIN_BROWSER_BROWSE_LEGACY
)

// DomainBrowserEvent represents events, generated by the
// [DomainBrowser].
type DomainBrowserEvent struct {
	Event    BrowserEvent      // Event code
	IfIndex  IfIndex           // Network interface index
	Protocol Protocol          // Network protocol
	Err      ErrCode           // In a case of BrowserFailure
	Flags    LookupResultFlags // Lookup flags
	Domain   string            // Domain name
}

// NewDomainBrowser creates a new [DomainBrowser].
//
// DomainBrowser constantly monitors the network for the available
// browsing/registration domains reports discovered information as
// a series of [DomainBrowserEvent] events via channel returned by the
// [DomainBrowser.Chan]
//
// Avahi documentation doesn't give a lot of explanation about purpose
// of this functionality, but [RFC6763, 11] gives some technical for details.
// In short, DomainBrowser performs DNS PTR queries in the following
// special domains:
//
//	DomainBrowserBrowse:            b._dns-sd._udp.<domain>.
//	DomainBrowserBrowseDefault:    db._dns-sd._udp.<domain>.
//	DomainBrowserRegister:          r._dns-sd._udp.<domain>.
//	DomainBrowserRegisterDefault:  dr._dns-sd._udp.<domain>.
//	DomainBrowserLegacy:           lb._dns-sd._udp.<domain>.
//
// According to RFC6763, the <domain> is usually "local", (meaning
// "perform the query using link-local multicast") or it may be learned
// through some other mechanism, such as the DHCP "Domain" option
// (option code 15) [RFC2132].
//
// So network administrator can configure some MDNS responder located
// in the local network to provide this information for applications.
//
// In fact, this mechanism seems to be rarely used in practice and
// provided here just for consistency.
//
// Function parameters:
//   - clnt is the pointer to [Client]
//   - ifindex is the network interface index. Use [IfIndexUnspec]
//     to monitor all interfaces.
//   - proto is the IP4/IP6 protocol, used as transport for queries. If
//     set to [ProtocolUnspec], both protocols will be used.
//   - domain is domain where domains are looked. If set to "", the
//     default domain is used, which depends on a avahi-daemon configuration
//     and usually is ".local"
//   - btype specified a type of domains being browses. See
//     [DomainBrowserType] for details.
//   - flags provide some lookup options. See [LookupFlags] for details.
//
// DomainBrowser must be closed after use with the [DomainBrowser.Close]
// function call.
//
// [RFC6763, 11]: https://datatracker.ietf.org/doc/html/rfc6763#section-11
// [RFC2132]: https://datatracker.ietf.org/doc/html/rfc2132
func NewDomainBrowser(
	clnt *Client,
	ifindex IfIndex,
	proto Protocol,
	domain string,
	btype DomainBrowserType,
	flags LookupFlags) (*DomainBrowser, error) {

	// Initialize DomainBrowser structure
	browser := &DomainBrowser{clnt: clnt}
	browser.handle = cgo.NewHandle(browser)
	browser.queue.init()

	// Convert strings from Go to C
	var cdomain *C.char
	if domain != "" {
		cdomain = C.CString(domain)
		defer C.free(unsafe.Pointer(cdomain))
	}

	// Create AvahiDomainBrowser
	avahiClient := clnt.begin()
	defer clnt.end()

	browser.avahiBrowser = C.avahi_domain_browser_new(
		avahiClient,
		C.AvahiIfIndex(ifindex),
		C.AvahiProtocol(proto),
		cdomain,
		C.AvahiDomainBrowserType(btype),
		C.AvahiLookupFlags(flags),
		C.AvahiDomainBrowserCallback(C.domainBrowserCallback),
		unsafe.Pointer(&browser.handle),
	)

	if browser.avahiBrowser == nil {
		browser.queue.Close()
		browser.handle.Delete()
		return nil, clnt.errno()
	}

	// Register self to be closed if Client is closed
	browser.clnt.addCloser(browser)

	return browser, nil
}

// Chan returns channel where [DomainBrowserEvent]s are sent.
func (browser *DomainBrowser) Chan() <-chan *DomainBrowserEvent {
	return browser.queue.Chan()
}

// Get waits for the next [DomainBrowserEvent].
//
// It returns:
//   - event, nil - if event available
//   - nil, error - if context is canceled
//   - nil, nil   - if DomainBrowser was closed
func (browser *DomainBrowser) Get(ctx context.Context) (*DomainBrowserEvent,
	error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case evnt := <-browser.Chan():
		return evnt, nil
	}
}

// Close closes the [DomainBrowser] and releases allocated resources.
// It closes the event channel, effectively unblocking pending readers.
//
// Note, double close is safe.
func (browser *DomainBrowser) Close() {
	if !browser.closed.Swap(true) {
		browser.clnt.begin()
		browser.clnt.delCloser(browser)
		C.avahi_domain_browser_free(browser.avahiBrowser)
		browser.avahiBrowser = nil
		browser.clnt.end()

		browser.queue.Close()
		browser.handle.Delete()
	}
}

// domainBrowserCallback called by AvahiDomainBrowser to
// report discovered services
//
//export domainBrowserCallback
func domainBrowserCallback(
	b *C.AvahiDomainBrowser,
	ifindex C.AvahiIfIndex,
	proto C.AvahiProtocol,
	event C.AvahiBrowserEvent,
	name, svctype, domain *C.char,
	flags C.AvahiLookupResultFlags,
	p unsafe.Pointer) {

	browser := (*cgo.Handle)(p).Value().(*DomainBrowser)

	evnt := &DomainBrowserEvent{
		Event:    BrowserEvent(event),
		IfIndex:  IfIndex(ifindex),
		Protocol: Protocol(proto),
		Flags:    LookupResultFlags(flags),
		Domain:   C.GoString(domain),
	}

	if evnt.Event == BrowserFailure {
		evnt.Err = browser.clnt.errno()
	}

	browser.queue.Push(evnt)
}
