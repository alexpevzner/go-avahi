// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Avahi IP4/IP6 protocol
//
//go:build linux || freebsd

package avahi

// #include <avahi-common/address.h>
import "C"
import "fmt"

// Protocol specifies IP4/IP6 protocol
type Protocol int

// Protocol values:
const (
	ProtocolIP4    Protocol = C.AVAHI_PROTO_INET
	ProtocolIP6    Protocol = C.AVAHI_PROTO_INET6
	ProtocolUnspec Protocol = C.AVAHI_PROTO_UNSPEC
)

// protocolNames contains names for valid Protocol values.
var protocolNames = map[Protocol]string{
	ProtocolIP4:    "ip4",
	ProtocolIP6:    "ip6",
	ProtocolUnspec: "unspec",
}

// String returns name of the Protocol.
func (proto Protocol) String() string {
	n := protocolNames[proto]
	if n == "" {
		n = fmt.Sprintf("UNKNOWN %d", int(proto))
	}
	return n

}
