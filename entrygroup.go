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
	"math"
	"net/netip"
	"runtime/cgo"
	"sync/atomic"
	"time"
	"unsafe"
)

// #include <stdlib.h>
// #include <avahi-client/client.h>
// #include <avahi-client/publish.h>
//
// void entryGroupCallback (
//	AvahiEntryGroup *g,
//	AvahiClientState s,
//	void *userdata);
import "C"

// EntryGroup represents a group of RR records published via avahi-daemon.
//
// All entries in the group are published or updated atomically.
type EntryGroup struct {
	clnt            *Client                      // Owning Client
	handle          cgo.Handle                   // Handle to self
	avahiEntryGroup *C.AvahiEntryGroup           // Avahi object
	queue           eventqueue[*EntryGroupEvent] // Event queue
	empty           atomic.Bool                  // The group is empty
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
	Name  string        // Record name
	Class DNSClass      // Record DNS class
	Type  DNSType       // Record DNS type
	TTL   time.Duration // DNS TTL, rounded to seconds and must fit int32
	Data  []byte        // Record data
}

// NewEntryGroup creates a new [EntryGroup].
func NewEntryGroup(clnt *Client) (*EntryGroup, error) {
	// Initialize EntryGroup structure
	egrp := &EntryGroup{clnt: clnt}
	egrp.handle = cgo.NewHandle(egrp)
	egrp.queue.init()
	egrp.empty.Store(true)

	// Create AvahiEntryGroup
	avahiClient := clnt.begin()
	defer clnt.end()

	egrp.avahiEntryGroup = C.avahi_entry_group_new(
		avahiClient,
		C.AvahiEntryGroupCallback(C.entryGroupCallback),
		unsafe.Pointer(&egrp.handle),
	)

	if egrp.avahiEntryGroup == nil {
		egrp.queue.Close()
		egrp.handle.Delete()
		return nil, clnt.errno()
	}

	return egrp, nil
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
	egrp.clnt.begin()
	defer egrp.clnt.end()

	rc := C.avahi_entry_group_commit(egrp.avahiEntryGroup)
	if rc >= 0 {
		return ErrCode(rc)
	}

	return nil
}

// Reset (purge) the EntryGroup. This takes effect immediately
// (without commit).
func (egrp *EntryGroup) Reset() error {
	egrp.clnt.begin()
	defer egrp.clnt.end()

	rc := C.avahi_entry_group_reset(egrp.avahiEntryGroup)
	if rc >= 0 {
		return ErrCode(rc)
	}

	egrp.empty.Store(true)

	return nil
}

// IsEmpty reports if EntryGroup is empty.
func (egrp *EntryGroup) IsEmpty() bool {
	return egrp.empty.Load()
}

// AddService adds a service registration
func (egrp *EntryGroup) AddService(
	ifindex IfIndex,
	proto Protocol,
	flags PublishFlags,
	svc *EntryGroupService) error {

	// Convert strings from Go to C
	cinstancename := C.CString(svc.InstanceName)
	defer C.free(unsafe.Pointer(cinstancename))

	ctype := C.CString(svc.Type)
	defer C.free(unsafe.Pointer(ctype))

	var cdomain *C.char
	if svc.Domain != "" {
		cdomain = C.CString(svc.Domain)
		defer C.free(unsafe.Pointer(cdomain))
	}

	var chostname *C.char
	if svc.Hostname != "" {
		chostname = C.CString(svc.Hostname)
		defer C.free(unsafe.Pointer(chostname))
	}

	// Convert TXT from Go to C
	ctxt, err := makeAvahiStringList(svc.Txt)
	if err != nil {
		return err
	}
	defer C.avahi_string_list_free(ctxt)

	// Call Avahi
	egrp.clnt.begin()
	defer egrp.clnt.end()

	rc := C.avahi_entry_group_add_service_strlst(
		egrp.avahiEntryGroup,
		C.AvahiIfIndex(ifindex),
		C.AvahiProtocol(proto),
		C.AvahiPublishFlags(flags),
		cinstancename,
		ctype,
		cdomain,
		chostname,
		C.uint16_t(svc.Port),
		ctxt,
	)

	if rc < 0 {
		return ErrCode(rc)
	}

	return nil
}

// AddServiceSubtype adds subtype for the existent service.
func (egrp *EntryGroup) AddServiceSubtype(
	ifindex IfIndex,
	proto Protocol,
	flags PublishFlags,
	svcid *EntryGroupServiceIdent,
	subtype string) error {

	// Convert strings from Go to C
	cinstancename := C.CString(svcid.InstanceName)
	defer C.free(unsafe.Pointer(cinstancename))

	ctype := C.CString(svcid.Type)
	defer C.free(unsafe.Pointer(ctype))

	var cdomain *C.char
	if svcid.Domain != "" {
		cdomain = C.CString(svcid.Domain)
		defer C.free(unsafe.Pointer(cdomain))
	}

	csubtype := C.CString(subtype)
	defer C.free(unsafe.Pointer(csubtype))

	// Call Avahi
	egrp.clnt.begin()
	defer egrp.clnt.end()

	rc := C.avahi_entry_group_add_service_subtype(
		egrp.avahiEntryGroup,
		C.AvahiIfIndex(ifindex),
		C.AvahiProtocol(proto),
		C.AvahiPublishFlags(flags),
		cinstancename,
		ctype,
		cdomain,
		csubtype,
	)

	if rc < 0 {
		return ErrCode(rc)
	}

	return nil
}

// UpdateServiceTxt updates TXT record for the existent service.
func (egrp *EntryGroup) UpdateServiceTxt(
	ifindex IfIndex,
	proto Protocol,
	flags PublishFlags,
	svcid *EntryGroupServiceIdent,
	txt []string) error {

	// Convert strings from Go to C
	cinstancename := C.CString(svcid.InstanceName)
	defer C.free(unsafe.Pointer(cinstancename))

	ctype := C.CString(svcid.Type)
	defer C.free(unsafe.Pointer(ctype))

	var cdomain *C.char
	if svcid.Domain != "" {
		cdomain = C.CString(svcid.Domain)
		defer C.free(unsafe.Pointer(cdomain))
	}

	// Convert TXT from Go to C
	ctxt, err := makeAvahiStringList(txt)
	if err != nil {
		return err
	}
	defer C.avahi_string_list_free(ctxt)

	// Call Avahi
	egrp.clnt.begin()
	defer egrp.clnt.end()

	rc := C.avahi_entry_group_update_service_txt_strlst(
		egrp.avahiEntryGroup,
		C.AvahiIfIndex(ifindex),
		C.AvahiProtocol(proto),
		C.AvahiPublishFlags(flags),
		cinstancename,
		ctype,
		cdomain,
		ctxt,
	)

	if rc < 0 {
		return ErrCode(rc)
	}

	return nil
}

// AddAddress adds host/address pair.
func (egrp *EntryGroup) AddAddress(
	ifindex IfIndex,
	proto Protocol,
	flags PublishFlags,
	hostname string, addr netip.Addr) error {

	// Convert address from Go to C
	caddr, err := makeAvahiAddress(addr)
	if err != nil {
		return err
	}

	// Convert strings from Go to C
	chostname := C.CString(hostname)
	defer C.free(unsafe.Pointer(chostname))

	// Call Avahi
	egrp.clnt.begin()
	defer egrp.clnt.end()

	rc := C.avahi_entry_group_add_address(
		egrp.avahiEntryGroup,
		C.AvahiIfIndex(ifindex),
		C.AvahiProtocol(proto),
		C.AvahiPublishFlags(flags),
		chostname,
		&caddr,
	)

	if rc < 0 {
		return ErrCode(rc)
	}

	return nil
}

// AddRecord adds a raw DNS record
func (egrp *EntryGroup) AddRecord(
	ifindex IfIndex,
	proto Protocol,
	flags PublishFlags,
	rec *EntryGroupRecord) error {
	// Convert TTL from Go to C
	if rec.TTL < 0 || rec.TTL > time.Second*math.MaxInt32 {
		return ErrInvalidTTL
	}

	cttl := C.uint32_t((rec.TTL + time.Second/2) / time.Second)

	// Convert strings from Go to C
	cname := C.CString(rec.Name)
	defer C.free(unsafe.Pointer(cname))

	// Convert record data from Go to C
	csize := C.size_t(len(rec.Data))
	cdata := C.CBytes(rec.Data)
	defer C.free(cdata)

	// Call Avahi
	egrp.clnt.begin()
	defer egrp.clnt.end()

	rc := C.avahi_entry_group_add_record(
		egrp.avahiEntryGroup,
		C.AvahiIfIndex(ifindex),
		C.AvahiProtocol(proto),
		C.AvahiPublishFlags(flags),
		cname,
		C.uint16_t(rec.Class),
		C.uint16_t(rec.Type),
		cttl,
		cdata,
		csize,
	)

	if rc < 0 {
		return ErrCode(rc)
	}

	return nil
}

// entryGroupCallback called by AvahiClient to report client state change
//
//export entryGroupCallback
func entryGroupCallback(
	g *C.AvahiEntryGroup,
	s C.AvahiClientState,
	p unsafe.Pointer) {

	clntHandle := *(*cgo.Handle)(p)
	egrp := clntHandle.Value().(*EntryGroup)

	state := EntryGroupState(s)
	evnt := &EntryGroupEvent{State: state}

	if state == EntryGroupStateFailure {
		evnt.Err = egrp.clnt.errno()
	}

	egrp.queue.Push(evnt)
}
