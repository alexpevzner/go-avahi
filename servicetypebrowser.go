// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Service browser
//
//go:build linux || freebsd

package avahi

import (
	"runtime/cgo"
	"unsafe"
)

// #include <stdlib.h>
// #include <avahi-client/lookup.h>
//
// void serviceTypeBrowserCallback (
//	AvahiServiceTypeBrowser *b,
//	AvahiIfIndex interface,
//	AvahiProtocol proto,
//	AvahiBrowserEvent event,
//	char *type,
//	char *domain,
//	AvahiLookupResultFlags flags,
//	void *userdata);
import "C"

// ServiceTypeBrowser returns available services of the specified type.
// Service type is a string that looks like "_http._tcp", "_ipp._tcp"
// and so on.
type ServiceTypeBrowser struct {
	clnt         *Client                              // Owning Client
	handle       cgo.Handle                           // Handle to self
	avahiBrowser *C.AvahiServiceTypeBrowser           // Underlying object
	queue        eventqueue[*ServiceTypeBrowserEvent] // Event queue
}

// ServiceTypeBrowserEvent represents events, generated by the
// [ServiceTypeBrowser].
type ServiceTypeBrowserEvent struct {
	Event    BrowserEvent      // Event code
	IfIndex  IfIndex           // Network interface index
	Protocol Protocol          // Network protocol
	Err      ErrCode           // In a case of BrowserFailure
	Flags    LookupResultFlags // Lookup flags
	Type     string            // Service type
	Domain   string            // Service domain
}

// NewServiceTypeBrowser creates a new [ServiceTypeBrowser].
//
// ServiceTypeBrowser constantly monitors the network for available
// service types and reports discovered information as a series of
// [ServiceTypeBrowserEvent] events via channel returned by the
// [ServiceTypeBrowser.Chan]
//
// Function parameters:
//   - clnt is the pointer to [Client]
//   - ifindex is the network interface index. Use [IfIndexUnspec]
//     to monitor all interfaces.
//   - proto is the IP4/IP6 protocol, used as transport for queries. If
//     set to [ProtocolUnspec], both protocols will be used.
//   - domain is domain where service is looked. If set to "", the
//     default domain is used, which depends on a avahi-daemon configuration
//     and usually is ".local"
//   - flags provide some lookup options. See [LookupFlags] for details.
func NewServiceTypeBrowser(
	clnt *Client,
	ifindex IfIndex,
	proto Protocol,
	domain string,
	flags LookupFlags) (*ServiceTypeBrowser, error) {

	// Initialize ServiceTypeBrowser structure
	browser := &ServiceTypeBrowser{clnt: clnt}
	browser.handle = cgo.NewHandle(browser)
	browser.queue.init()

	// Convert strings from Go to C
	var cdomain *C.char
	if domain != "" {
		cdomain = C.CString(domain)
		defer C.free(unsafe.Pointer(cdomain))
	}

	// Create AvahiServiceBrowser
	avahiClient := clnt.begin()
	defer clnt.end()

	browser.avahiBrowser = C.avahi_service_type_browser_new(
		avahiClient,
		C.AvahiIfIndex(ifindex),
		C.AvahiProtocol(proto),
		cdomain,
		C.AvahiLookupFlags(flags),
		C.AvahiServiceBrowserCallback(C.serviceTypeBrowserCallback),
		unsafe.Pointer(&browser.handle),
	)

	if browser.avahiBrowser == nil {
		browser.queue.Close()
		browser.handle.Delete()
		return nil, clnt.errno()
	}

	return browser, nil
}

// Chan returns channel where [ServiceBrowserEvent]s are sent.
func (browser *ServiceTypeBrowser) Chan() <-chan *ServiceTypeBrowserEvent {
	return browser.queue.Chan()
}

// Close closes the [ServiceTypeBrowser] and releases allocated resources.
// It closes the event channel, effectively unblocking pending readers.
func (browser *ServiceTypeBrowser) Close() {
	browser.clnt.begin()
	C.avahi_service_type_browser_free(browser.avahiBrowser)
	browser.avahiBrowser = nil
	browser.clnt.end()

	browser.queue.Close()
	browser.handle.Delete()
}

// serviceTypeBrowserCallback called by AvahiServiceTypeBrowser to
// report discovered services
//
//export serviceTypeBrowserCallback
func serviceTypeBrowserCallback(
	b *C.AvahiServiceTypeBrowser,
	ifindex C.AvahiIfIndex,
	proto C.AvahiProtocol,
	event C.AvahiBrowserEvent,
	svctype, domain *C.char,
	flags C.AvahiLookupResultFlags,
	p unsafe.Pointer) {

	browser := (*cgo.Handle)(p).Value().(*ServiceTypeBrowser)

	evnt := &ServiceTypeBrowserEvent{
		Event:    BrowserEvent(event),
		IfIndex:  IfIndex(ifindex),
		Protocol: Protocol(proto),
		Err:      browser.clnt.errno(),
		Flags:    LookupResultFlags(flags),
		Type:     C.GoString(svctype),
		Domain:   C.GoString(domain),
	}

	browser.queue.Push(evnt)
}