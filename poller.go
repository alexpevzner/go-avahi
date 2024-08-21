// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Event poller
//
//go:build linux || freebsd

package avahi

import (
	"context"
	"reflect"
	"sync"
)

// Poller is the convenience object, that implements a centralized
// events reception from multiple sources.
//
// Multiple Event sources ([Client], Browsers, Resolvers and [EntryGroup])
// can be added to the Poller. Poller combines their events flows together
// and makes it available via single [Poller.Poll] API call.
type Poller struct {
	sources []reflect.SelectCase
	lock    sync.Mutex
}

// NewPoller creates a new [Poller]
func NewPoller() *Poller {
	return &Poller{}
}

// Poll waits for the next event from any of registered sources.
//
// It returns:
//   - nil, error - if context is canceled
//   - event, nil - if event is available
//
// The returned event is one of the following:
//   - [*ClientEvent]
//   - [*DomainBrowserEvent]
//   - [*RecordBrowserEvent]
//   - [*ServiceBrowserEvent]
//   - [*ServiceTypeBrowserEvent]
//   - [*AddressResolverEvent]
//   - [*HostNameResolverEvent]
//   - [*ServiceResolverEvent]
//
// If source is added while Poll is active, it may or may not affect
// the pending Poll, no guarantees are provided here except for safety
// guarantees.
//
// Events, received from the same source, are never reordered between
// each other, but events from different sources may be reordered.
//
// Adding the same source to the multiple Pollers has roughly the
// same effect as reading the same channel from multiple goroutines
// and generally not recommended.
func (p *Poller) Poll(ctx context.Context) (any, error) {
	for ctx.Err() == nil {
		// Snapshot current select sources, as it may change while
		// poll is blocked.

		// Prepend Context channel.
		p.lock.Lock()

		sources := make([]reflect.SelectCase, len(p.sources)+1)
		sources[0] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		}
		copy(sources[1:], p.sources)

		p.lock.Unlock()

		// Wait for an event
		chosen, recv, ok := reflect.Select(sources)
		switch {
		case chosen == 0:
			// Recv from the Context's channel. Just do nothing,
			// the loop condition will terminate the loop

		case !ok:
			// Recv from the closed channel. Remove the source
			// and retry.
			p.delSource(sources[chosen].Chan)

		default:
			// We have a new event
			return recv.Interface(), nil
		}
	}

	return nil, ctx.Err()
}

// AddClient adds [Client] as the event source.
func (p *Poller) AddClient(clnt *Client) {
	pollerAddSource(p, clnt.Chan())
}

// AddDomainBrowser adds [DomainBrowser] as the event source.
func (p *Poller) AddDomainBrowser(browser *DomainBrowser) {
	pollerAddSource(p, browser.Chan())
}

// AddRecordBrowser adds [RecordBrowser] as the event source.
func (p *Poller) AddRecordBrowser(browser *RecordBrowser) {
	pollerAddSource(p, browser.Chan())
}

// AddServiceBrowser adds [ServiceBrowser] as the event source.
func (p *Poller) AddServiceBrowser(browser *ServiceBrowser) {
	pollerAddSource(p, browser.Chan())
}

// AddServiceTypeBrowser adds [ServiceTypeBrowser] as the event source.
func (p *Poller) AddServiceTypeBrowser(browser *ServiceTypeBrowser) {
	pollerAddSource(p, browser.Chan())
}

// AddAddressResolver adds [AddressResolver] as the event source.
func (p *Poller) AddAddressResolver(resolver *AddressResolver) {
	pollerAddSource(p, resolver.Chan())
}

// AddHostNameResolver adds [HostNameResolver] as the event source.
func (p *Poller) AddHostNameResolver(resolver *HostNameResolver) {
	pollerAddSource(p, resolver.Chan())
}

// AddServiceResolver adds [ServiceResolver] as the event source.
func (p *Poller) AddServiceResolver(resolver *ServiceResolver) {
	pollerAddSource(p, resolver.Chan())
}

// pollerAddSource adds the source channel to the Poller
func pollerAddSource[T any](p *Poller, chn <-chan T) {
	source := reflect.ValueOf(chn)

	p.lock.Lock()
	defer p.lock.Unlock()

	for i := range p.sources {
		if p.sources[i].Chan == source {
			return
		}
	}

	p.sources = append(p.sources, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: source,
	})
}

// delSource deletes the source channel, which must be passed as reflect.Value.
func (p *Poller) delSource(source reflect.Value) {
	for i := range p.sources {
		if p.sources[i].Chan == source {
			copy(p.sources[i:], p.sources[i+1:])
			p.sources = p.sources[:len(p.sources)-1]
			return
		}
	}
}
