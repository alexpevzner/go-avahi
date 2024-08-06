// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// DNS constants
//
//go:build linux || freebsd

package avahi

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
