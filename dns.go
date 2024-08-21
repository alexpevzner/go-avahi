// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// DNS constants
//
//go:build linux || freebsd

package avahi

import "net/netip"

// DNSClass represents a DNS record class. See [RFC1035, 3.2.4.] for details.
//
// [RFC1035, 3.2.4.]: https://datatracker.ietf.org/doc/html/rfc1035#section-3.2.4
type DNSClass int

// DNSClass values:
const (
	DNSClassIN DNSClass = 1
)

// DNSType represents a DNS record type.
//
// For details, see:
//
//   - [RFC1035, 3.2.2.] - common record types
//   - [RFC2782] - SRV record
//   - [RFC3596] - AAAA record
//
// [RFC1035, 3.2.2.]: https://datatracker.ietf.org/doc/html/rfc1035#section-3.2.2
// [RFC2782]: https://datatracker.ietf.org/doc/html/rfc2782
// [RFC3596]: https://datatracker.ietf.org/doc/html/rfc3596
type DNSType int

// DNSType values
const (
	DNSTypeA     DNSType = 1  // IP4 host address
	DNSTypeNS    DNSType = 2  // An authoritative name server
	DNSTypeCNAME DNSType = 5  // The canonical name for an alias
	DNSTypeSOA   DNSType = 6  // SOA record
	DNSTypePTR   DNSType = 12 // A domain name pointer
	DNSTypeHINFO DNSType = 13 // Host information
	DNSTypeMX    DNSType = 15 // Mail exchange
	DNSTypeTXT   DNSType = 16 // Text strings
	DNSTypeAAAA  DNSType = 28 // IP6 host address (RFC3596)
	DNSTypeSRV   DNSType = 33 // Service record (RFC2782)
)

// DNSDecodeA decodes A type resource record.
//
// It returns a real IPv4 (not IPv6-encoded IPv4) address.
//
// [RecordBrowserEvent].RData can be used as input.
// Errors reported by returning zero [netip.Addr]
func DNSDecodeA(rdata []byte) netip.Addr {
	var addr netip.Addr
	if len(rdata) == 4 {
		addr, _ = netip.AddrFromSlice(rdata)
		addr = addr.Unmap()
	}
	return addr
}

// DNSDecodeAAAA decodes AAAA type resource record.
//
// [RecordBrowserEvent].RData can be used as input.
// Errors reported by returning zero [netip.Addr]
func DNSDecodeAAAA(rdata []byte) netip.Addr {
	var addr netip.Addr
	if len(rdata) == 16 {
		addr, _ = netip.AddrFromSlice(rdata)
	}
	return addr
}

// DNSDecodeTXT decodes TXT type resource record.
//
// [RecordBrowserEvent].RData can be used as input.
// Errors reported by returning nil slice.
func DNSDecodeTXT(rdata []byte) []string {
	txt := []string{}

	for len(rdata) > 0 {
		// Extract size of the next string
		sz := int(rdata[0])
		rdata = rdata[1:]

		// Size exceeds available data
		if sz > len(rdata) {
			return nil
		}

		// Extract next string. Ignore empty ones.
		if sz > 0 {
			s := string(rdata[:sz])
			rdata = rdata[sz:]
			txt = append(txt, s)
		}
	}

	return txt
}
