// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// CGo glue
//
//go:build linux || freebsd

package avahi

import (
	"net/netip"
	"unsafe"
)

// #cgo pkg-config: avahi-client
//
// #include <avahi-client/client.h>
import "C"

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

// strcaseequal compares two strings ignoring case, as C does,
// i.e. without any special interpretation of UTF-8 sequences.
func strcaseequal(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := 0; i < len(s1); i++ {
		c1 := s1[i]
		c2 := s2[i]

		switch {
		case c1 == c2:
		case toupper(c1) == toupper(c2):
		default:
			return false
		}
	}

	return true
}

// toupper converts ASCII character to uppercase
func toupper(c byte) byte {
	if 'a' <= c && c <= 'z' {
		c = c - 'a' + 'A'
	}
	return c
}
