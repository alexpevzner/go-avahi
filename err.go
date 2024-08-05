// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Avahi error codes
//
//go:build linux || freebsd

package avahi

// #include <avahi-common/error.h>
import "C"

// ErrCode represents an Avahi error code
type ErrCode int

// Error codes:
const (
	// No error
	NoError ErrCode = C.AVAHI_OK
	// Generic error code
	ErrFailure ErrCode = C.AVAHI_ERR_FAILURE
	// Object was in a bad state
	ErrBadState ErrCode = C.AVAHI_ERR_BAD_STATE
	// Invalid host name
	ErrInvalidHostName ErrCode = C.AVAHI_ERR_INVALID_HOST_NAME
	// Invalid domain name
	ErrInvalidDomainName ErrCode = C.AVAHI_ERR_INVALID_DOMAIN_NAME
	// No suitable network protocol available
	ErrNoNetwork ErrCode = C.AVAHI_ERR_NO_NETWORK
	// Invalid DNS TTL
	ErrInvalidTTL ErrCode = C.AVAHI_ERR_INVALID_TTL
	// RR key is pattern
	ErrIsPattern ErrCode = C.AVAHI_ERR_IS_PATTERN
	// Name collision
	ErrCollision ErrCode = C.AVAHI_ERR_COLLISION
	// Invalid RR
	ErrInvalidRecord ErrCode = C.AVAHI_ERR_INVALID_RECORD

	// Invalid service name
	ErrInvalidServiceName ErrCode = C.AVAHI_ERR_INVALID_SERVICE_NAME
	// Invalid service type
	ErrInvalidServiceType ErrCode = C.AVAHI_ERR_INVALID_SERVICE_TYPE
	// Invalid port number
	ErrInvalidPort ErrCode = C.AVAHI_ERR_INVALID_PORT
	// Invalid key
	ErrInvalidKey ErrCode = C.AVAHI_ERR_INVALID_KEY
	// Invalid address
	ErrInvalidAddress ErrCode = C.AVAHI_ERR_INVALID_ADDRESS
	// Timeout reached
	ErrTimeout ErrCode = C.AVAHI_ERR_TIMEOUT
	// Too many clients
	ErrTooManyClients ErrCode = C.AVAHI_ERR_TOO_MANY_CLIENTS
	// Too many objects
	ErrTooManyObjects ErrCode = C.AVAHI_ERR_TOO_MANY_OBJECTS
	// Too many entries
	ErrTooManyEntries ErrCode = C.AVAHI_ERR_TOO_MANY_ENTRIES
	// OS error
	ErrOS ErrCode = C.AVAHI_ERR_OS

	// Access denied
	ErrAccessDenied ErrCode = C.AVAHI_ERR_ACCESS_DENIED
	// Invalid operation
	ErrInvalidOperation ErrCode = C.AVAHI_ERR_INVALID_OPERATION
	// An unexpected D-Bus error occurred
	ErrDbusError ErrCode = C.AVAHI_ERR_DBUS_ERROR
	// Daemon connection failed
	ErrDisconnected ErrCode = C.AVAHI_ERR_DISCONNECTED
	// Memory exhausted
	ErrNoMemory ErrCode = C.AVAHI_ERR_NO_MEMORY
	// The object passed to this function was invalid
	ErrInvalidObject ErrCode = C.AVAHI_ERR_INVALID_OBJECT
	// Daemon not running
	ErrNoDaemon ErrCode = C.AVAHI_ERR_NO_DAEMON
	// Invalid interface
	ErrInvalidInterface ErrCode = C.AVAHI_ERR_INVALID_INTERFACE
	// Invalid protocol
	ErrInvalidProtocol ErrCode = C.AVAHI_ERR_INVALID_PROTOCOL
	// Invalid flags
	ErrInvalidFlags ErrCode = C.AVAHI_ERR_INVALID_FLAGS

	// Not found
	ErrNotFound ErrCode = C.AVAHI_ERR_NOT_FOUND
	// Configuration error
	ErrInvalidConfig ErrCode = C.AVAHI_ERR_INVALID_CONFIG
	// Verson mismatch
	ErrVersionMismatch ErrCode = C.AVAHI_ERR_VERSION_MISMATCH
	// Invalid service subtype
	ErrInvalidServiceSubtype ErrCode = C.AVAHI_ERR_INVALID_SERVICE_SUBTYPE
	// Invalid packet
	ErrInvalidPacket ErrCode = C.AVAHI_ERR_INVALID_PACKET
	// Invlaid DNS return code
	ErrInvalidDNSError ErrCode = C.AVAHI_ERR_INVALID_DNS_ERROR
	// DNS Error: Form error
	ErrDNSFormerr ErrCode = C.AVAHI_ERR_DNS_FORMERR
	// DNS Error: Server Failure
	ErrDNSSERVFAIL ErrCode = C.AVAHI_ERR_DNS_SERVFAIL
	// DNS Error: No such domain
	ErrDNSNXDOMAIN ErrCode = C.AVAHI_ERR_DNS_NXDOMAIN
	// DNS Error: Not implemented
	ErrDNSNotimp ErrCode = C.AVAHI_ERR_DNS_NOTIMP

	// DNS Error: Operation refused
	ErrDNSREFUSED ErrCode = C.AVAHI_ERR_DNS_REFUSED
	// DNS Error: YXDOMAIN
	ErrDNSYXDOMAIN ErrCode = C.AVAHI_ERR_DNS_YXDOMAIN
	// DNS Error: YXRRSET
	ErrDNSYXRRSET ErrCode = C.AVAHI_ERR_DNS_YXRRSET
	// DNS Error: NXRRSET
	ErrDNSNXRRSET ErrCode = C.AVAHI_ERR_DNS_NXRRSET
	// DNS Error: Not authorized
	ErrDNSNOTAUTH ErrCode = C.AVAHI_ERR_DNS_NOTAUTH
	// DNS Error: NOTZONE
	ErrDNSNOTZONE ErrCode = C.AVAHI_ERR_DNS_NOTZONE

	// Invalid RDATA
	ErrInvalidRDATA ErrCode = C.AVAHI_ERR_INVALID_RDATA
	// Invalid DNS class
	ErrInvalidDNSClass ErrCode = C.AVAHI_ERR_INVALID_DNS_CLASS
	// Invalid DNS type
	ErrInvalidDNSType ErrCode = C.AVAHI_ERR_INVALID_DNS_TYPE
	// Not supported
	ErrNotSupported ErrCode = C.AVAHI_ERR_NOT_SUPPORTED

	// Operation not permitted
	ErrNotPermitted ErrCode = C.AVAHI_ERR_NOT_PERMITTED
	// Invalid argument
	ErrInvalidArgument ErrCode = C.AVAHI_ERR_INVALID_ARGUMENT
	// Is empty
	ErrIsEmpty ErrCode = C.AVAHI_ERR_IS_EMPTY
	// The requested operation is invalid because it is redundant
	ErrNoChange ErrCode = C.AVAHI_ERR_NO_CHANGE
)

// Error returns error string.
// It implements error interface.
func (err ErrCode) Error() string {
	s := C.avahi_strerror(C.int(err))
	return "avahi: " + C.GoString(s)
}
