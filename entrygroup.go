// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Avahi Entry Group (the publishing API)
//
//go:build linux || freebsd

package avahi

import (
	"context"
	"errors"
	"net/netip"
	"runtime/cgo"
)

// #include <avahi-client/client.h>
// #include <avahi-client/publish.h>
//
// void clientCallback (AvahiClient*, AvahiClientState, void*);
import "C"

// EntryGroup represents a group of RR records published via avahi-daemon.
//
// All entries in the group are published or updated atomically.
type EntryGroup struct {
	clnt            *Client                      // Owning Client
	handle          cgo.Handle                   // Handle to self
	avahiEntryGroup *C.AvahiEntryGroup           // Avahi object
	queue           eventqueue[*EntryGroupEvent] // Event queue
}

// EntryGroupEvent represents an [EntryGroup] state change event.
type EntryGroupEvent struct {
	State EntryGroupState // Entry group state
	Err   ErrCode         // In a case of EntryGroupStateFailure
}

// EntryGroupServiceIdent contains common set of parameters
// that identify a service in EntryGroup.
//
// It is used as a part of [EntryGroupService] for service
// registration and also as an standalone parameter that identifies
// a service for modification of existent services entries with
// [EntryGroup.AddServiceSubtype] and [EntryGroup.UpdateServiceTxt]
// functions.
type EntryGroupServiceIdent struct {
	InstanceName string // Service instance name
	Type         string // Service type
	Domain       string // Service domain (use "" for default)
}

// EntryGroupService represents a service registration.
type EntryGroupService struct {
	EntryGroupServiceIdent              // Service identification
	Hostname               string       // Host name (use "" for default)
	Port                   int          // IP port
	Txt                    []string     // TXT record ("key=value"...)
	Flags                  PublishFlags // Publishing flags
}

// EntryGroupRecord represents a raw DNS record that can be added
// to the EntryGroup.
type EntryGroupRecord struct {
	Name  string   // Record name
	Class DNSClass // Record DNS class
	Type  DNSType  // Record DNS type
	TTL   int      // DNS TTL, in seconds
	Data  []byte   // Record data
}

// NewEntryGroup creates a new [EntryGroup].
func NewEntryGroup(clnt *Client) (*EntryGroup, error) {
	return nil, errors.New("not implemented")
}

// Chan returns channel where [EntryGroupEvent]s are sent.
func (egrp *EntryGroup) Chan() <-chan *EntryGroupEvent {
	return egrp.queue.Chan()
}

// Get waits for the next [EntryGroupEvent].
//
// It returns:
//   - event, nil - if event available
//   - nil, error - if context is canceled
//   - nil, nil   - if EntryGroup was closed
func (egrp *EntryGroup) Get(ctx context.Context) (*EntryGroupEvent, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case evnt := <-egrp.Chan():
		return evnt, nil
	}
}

// Close closed the [EntryGroup].
func (egrp *EntryGroup) Close() {
	egrp.clnt.begin()
	C.avahi_entry_group_free(egrp.avahiEntryGroup)
	egrp.avahiEntryGroup = nil
	egrp.clnt.end()

	egrp.queue.Close()
	egrp.handle.Delete()
}

// Commit changes to the EntryGroup.
func (egrp *EntryGroup) Commit() error {
	return errors.New("not implemented")
}

// Reset (purge) the EntryGroup. This takes effect immediately
// (without commit).
func (egrp *EntryGroup) Reset() error {
	return errors.New("not implemented")
}

// IsEmpty reports if EntryGroup is empty.
func (egrp *EntryGroup) IsEmpty() bool {
	return false
}

// AddService adds a service registration
func (egrp *EntryGroup) AddService(svc *EntryGroupService) error {
	return errors.New("not implemented")
}

// AddServiceSubtype adds subtype for the existent service.
func (egrp *EntryGroup) AddServiceSubtype(svcid *EntryGroupServiceIdent,
	subtype string) error {
	return errors.New("not implemented")
}

// UpdateServiceTxt updates TXT record for the existent service.
func (egrp *EntryGroup) UpdateServiceTxt(svcid *EntryGroupServiceIdent,
	txt []string) error {
	return errors.New("not implemented")
}

// AddAddress adds host/address pair.
func (egrp *EntryGroup) AddAddress(hostname string, addr netip.Addr) error {
	return errors.New("not implemented")
}

// AddRecord adds a raw DNS record
func (egrp *EntryGroup) AddRecord(rec *EntryGroupRecord) error {
	return errors.New("not implemented")
}
