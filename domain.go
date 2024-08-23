// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Domain name string
//
//go:build linux || freebsd

package avahi

// #include <stdlib.h>
// #include <avahi-common/domain.h>
import "C"
import (
	"unsafe"
)

// DomainFrom makes escaped domain name string from a sequence of unescaped
// labels:
//
//	["Ex.Ample", "com"] -> "Ex\.Ample.com"
//
// Note, this function is not guaranteed to escape labels exactly
// as Avahi does, but output is anyway correct.
func DomainFrom(labels []string) string {
	buf := make([]byte, 0, 256)

	for n, label := range labels {
		if n != 0 {
			buf = append(buf, '.')
		}

		for i := 0; i < len(label); i++ {
			c := label[i]
			switch c {
			case '.', '\\':
				buf = append(buf, '\\', c)
			default:
				buf = append(buf, c)
			}
		}
	}

	return string(buf)
}

// DomainSlice splits escaped domain name into a sequence of unescaped
// labels:
//
//	"Ex\.Ample.com" -> ["Ex.Ample", "com"]
//
// In a case of error it returns nil.
func DomainSlice(d string) []string {
	// Convert input from Go to C
	in := C.CString(d)
	defer C.free(unsafe.Pointer(in))

	// Allocate decode buffer. len(d) must be enough.
	buflen := C.size_t(len(d))
	buf := C.malloc(buflen)

	// Decode label by label
	labels := []string{}

	next := in
	for *next != 0 {
		clabel := C.avahi_unescape_label(&next, (*C.char)(buf), buflen)
		if clabel == nil {
			return nil
		}

		labels = append(labels, C.GoString(clabel))
	}

	return labels
}

// DomainEqual reports if two domain names are equal.
//
// Note, invalid domain names are never equal to anything
// else, including itself.
func DomainEqual(d1, d2 string) bool {
	labels1 := DomainSlice(d1)
	labels2 := DomainSlice(d2)

	if labels1 == nil || labels2 == nil {
		return false
	}

	if len(labels1) != len(labels2) {
		return false
	}

	for i := range labels1 {
		if !strcaseequal(labels1[i], labels2[i]) {
			return false
		}
	}

	return true
}

// DomainNormalize normalizes the domain name by removing unneeded escaping.
//
// In a case of error it returns empty string.
func DomainNormalize(d string) string {
	return DomainFrom(DomainSlice(d))
}

// DomainToLower converts domain name to upper case.
//
// It only touches ASCII uppercase letters "a" to "z",
// leaving all other bytes untouched.
//
// See [RFC6762, 16. Multicast DNS Character Set] for details.
//
// [RFC6762, 16. Multicast DNS Character Set]: https://datatracker.ietf.org/doc/html/rfc6762#section-16
func DomainToLower(d string) string {
	buf := []byte(d)

	for i := range buf {
		if c := buf[i]; 'A' <= c && c <= 'Z' {
			buf[i] = c - 'A' + 'a'
		}
	}

	return string(buf)
}

// DomainToUpper converts domain name to upper case.
//
// It only touches ASCII lowercase letters "a" to "z",
// leaving all other bytes untouched.
//
// See [RFC6762, 16. Multicast DNS Character Set] for details.
//
// [RFC6762, 16. Multicast DNS Character Set]: https://datatracker.ietf.org/doc/html/rfc6762#section-16
func DomainToUpper(d string) string {
	buf := []byte(d)

	for i := range buf {
		if c := buf[i]; 'a' <= c && c <= 'z' {
			buf[i] = c - 'a' + 'A'
		}
	}

	return string(buf)
}

// DomainServiceNameSplit splits service name into instance, service type
// and domain components:
//
//	"Kyocera ECOSYS M2040dn._ipp._tcp.local" -->
//	    --> ["Kyocera ECOSYS M2040dn", "_ipp._tcp", "local"]
//
// In a case of error it returns empty strings
func DomainServiceNameSplit(nm string) (instance, svctype, domain string) {
	// Slice domain name into labels
	labels := DomainSlice(nm)
	if len(labels) < 3 {
		// At least 3 labels are required: instance name
		// plus service type, which is two labels at least
		return
	}

	// First label is service name. Then some labels are
	// service type. We consider every label in sequence, starting
	// with the underscore character, a part of service type.
	// The reminder is domain.
	//
	// So find range of labels that belong to the service type.
	svcTypeBeg := 1
	svcTypeEnd := 1

	for svcTypeEnd < len(labels) &&
		len(labels[svcTypeEnd]) > 1 && labels[svcTypeEnd][0] == '_' {
		svcTypeEnd++
	}

	if svcTypeEnd-svcTypeBeg < 2 {
		// At least 2 labels required
		return
	}

	instance = labels[0]
	svctype = DomainFrom(labels[svcTypeBeg:svcTypeEnd])
	domain = DomainFrom(labels[svcTypeEnd:])

	return
}

// DomainServiceNameJoin merges two parts of the full service
// name (instance name, service type and domain name) into
// the full service name.
//
//   - instance MUST be unescaped label
//   - svctype and domain MUST be escaped domain names
//   - instance and svctype MUST NOT be empty
//
// In a case of error it returns empty strings.
// Strong validation of input strings is not performed here.
func DomainServiceNameJoin(instance, svctype, domain string) string {
	// instance and svctype must not be empty
	if instance == "" || svctype == "" {
		return ""
	}

	// Escape instance name
	instance = DomainFrom([]string{instance})

	// Join parts together
	out := instance + "." + svctype
	if domain != "" {
		out += "." + domain
	}

	return out
}
