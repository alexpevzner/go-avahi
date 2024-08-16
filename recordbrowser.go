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
// void recordBrowserCallback (
//	AvahiRecordBrowser *b,
//	AvahiIfIndex interface,
//	AvahiProtocol proto,
//	AvahiBrowserEvent event,
//	char *name,
//	uint16_t dnsclass,
//	uint16_t dnstype,
//	void *rdata,
//	size_t size,
//	AvahiLookupResultFlags flags,
//	void *userdata);
import "C"

// RecordBrowser is the generic browser for resource records of
// the specified name, class and type.
type RecordBrowser struct {
	clnt         *Client                         // Owning Client
	handle       cgo.Handle                      // Handle to self
	avahiBrowser *C.AvahiRecordBrowser           // Underlying object
	queue        eventqueue[*RecordBrowserEvent] // Event queue
	closed       atomic.Bool                     // Browser is closed
}

// RecordBrowserEvent represents events, generated by the
// [RecordBrowser].
type RecordBrowserEvent struct {
	Event  BrowserEvent      // Event code
	IfIdx  IfIndex           // Network interface index
	Proto  Protocol          // Network protocol
	Err    ErrCode           // In a case of BrowserFailure
	Flags  LookupResultFlags // Lookup flags
	Name   string            // Record name
	RClass DNSClass          // Record DNS class
	RType  DNSType           // Record DNS type
	RData  []byte            // Record data
}

// NewRecordBrowser creates a new [RecordBrowser].
//
// RecordBrowser is the generic browser for RRs of the specified
// name, class and type. It uses [RecordBrowserEvent] to report
// discovered information as via channel returned by the
// [RecordBrowser.Chan]
//
// Function parameters:
//   - clnt is the pointer to [Client]
//   - ifidx is the network interface index. Use [IfIndexUnspec]
//     to monitor all interfaces.
//   - proto is the IP4/IP6 protocol, used as transport for queries. If
//     set to [ProtocolUnspec], both protocols will be used.
//   - name is the RR name to look for
//   - dnsclass is the DNS class to look for (most likely, DNSClassIN)
//   - dnstype is the DNS record type.
//   - flags provide some lookup options. See [LookupFlags] for details.
//
// RecordBrowser must be closed after use with the [RecordBrowser.Close]
// function call.
func NewRecordBrowser(
	clnt *Client,
	ifidx IfIndex,
	proto Protocol,
	name string,
	dnsclass DNSClass,
	dnstype DNSType,
	flags LookupFlags) (*RecordBrowser, error) {

	// Initialize RecordBrowser structure
	browser := &RecordBrowser{clnt: clnt}
	browser.handle = cgo.NewHandle(browser)
	browser.queue.init()

	// Convert strings from Go to C
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	// Create AvahiRecordBrowser
	avahiClient := clnt.begin()
	defer clnt.end()

	browser.avahiBrowser = C.avahi_record_browser_new(
		avahiClient,
		C.AvahiIfIndex(ifidx),
		C.AvahiProtocol(proto),
		cname,
		C.uint16_t(dnsclass),
		C.uint16_t(dnstype),
		C.AvahiLookupFlags(flags),
		C.AvahiRecordBrowserCallback(C.recordBrowserCallback),
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

// Chan returns channel where [RecordBrowserEvent]s are sent.
func (browser *RecordBrowser) Chan() <-chan *RecordBrowserEvent {
	return browser.queue.Chan()
}

// Get waits for the next [RecordBrowserEvent].
//
// It returns:
//   - event, nil - if event available
//   - nil, error - if context is canceled
//   - nil, nil   - if RecordBrowser was closed
func (browser *RecordBrowser) Get(ctx context.Context) (*RecordBrowserEvent,
	error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case evnt := <-browser.Chan():
		return evnt, nil
	}
}

// Close closes the [RecordBrowser] and releases allocated resources.
// It closes the event channel, effectively unblocking pending readers.
//
// Note, double close is safe.
func (browser *RecordBrowser) Close() {
	if !browser.closed.Swap(true) {
		browser.clnt.begin()
		browser.clnt.delCloser(browser)
		C.avahi_record_browser_free(browser.avahiBrowser)
		browser.avahiBrowser = nil
		browser.clnt.end()

		browser.queue.Close()
		browser.handle.Delete()
	}
}

// recordBrowserCallback called by AvahiRecordBrowser to
// report discovered services
//
//export recordBrowserCallback
func recordBrowserCallback(
	b *C.AvahiRecordBrowser,
	ifidx C.AvahiIfIndex,
	proto C.AvahiProtocol,
	event C.AvahiBrowserEvent,
	name *C.char,
	dnsclass, dnstype C.uint16_t,
	rdata unsafe.Pointer,
	rsize C.size_t,
	flags C.AvahiLookupResultFlags,
	p unsafe.Pointer) {

	browser := (*cgo.Handle)(p).Value().(*RecordBrowser)

	evnt := &RecordBrowserEvent{
		Event:  BrowserEvent(event),
		IfIdx:  IfIndex(ifidx),
		Proto:  Protocol(proto),
		Flags:  LookupResultFlags(flags),
		Name:   C.GoString(name),
		RClass: DNSClass(dnsclass),
		RType:  DNSType(dnstype),
	}

	if rdata != nil {
		evnt.RData = C.GoBytes(rdata, C.int(rsize))
	}

	if evnt.Event == BrowserFailure {
		evnt.Err = browser.clnt.errno()
	}

	browser.queue.Push(evnt)
}
