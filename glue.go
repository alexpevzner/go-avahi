// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// CGo glue
//
//go:build linux || freebsd

package avahi

// #cgo pkg-config: avahi-client
//
// #include <avahi-client/client.h>
import "C"
import (
	"net/netip"
	"unsafe"
)

// makeAvahiAddress makes C.AvahiAddress
func makeAvahiAddress(addr netip.Addr) (C.AvahiAddress, error) {
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
		return caddr, ErrInvalidAddress
	}

	return caddr, nil
}

// decodeAvahiAddress decodes C.AvahiAddress
func decodeAvahiAddress(caddr *C.AvahiAddress) netip.Addr {
	var ip netip.Addr

	switch {
	case caddr == nil:
		// Do nothing

	case caddr.proto == C.AVAHI_PROTO_INET:
		ip = netip.AddrFrom4(*(*[4]byte)(unsafe.Pointer(&caddr.data)))
	case caddr.proto == C.AVAHI_PROTO_INET6:
		ip = netip.AddrFrom16(*(*[16]byte)(unsafe.Pointer(&caddr.data)))
	}

	return ip
}

// makeAvahiStringList makes C.AvahiStringList
func makeAvahiStringList(txt []string) (*C.AvahiStringList, error) {
	var ctxt *C.AvahiStringList

	for i := len(txt) - 1; i > 0; i-- {
		b := []byte(txt[i])

		prev := ctxt
		ctxt = C.avahi_string_list_add_arbitrary(
			ctxt,
			(*C.uint8_t)(unsafe.Pointer(&b[0])),
			C.size_t(len(b)),
		)

		if ctxt == nil {
			C.avahi_string_list_free(prev)
			return nil, ErrNoMemory
		}
	}

	return ctxt, nil
}

// decodeAvahiStringList decodes C.AvahiStringList
func decodeAvahiStringList(ctxt *C.AvahiStringList) []string {
	var txt []string

	for ctxt != nil {
		t := C.GoStringN((*C.char)(unsafe.Pointer(&ctxt.text)),
			C.int(ctxt.size))
		txt = append(txt, t)

		ctxt = ctxt.next
	}

	return txt
}
