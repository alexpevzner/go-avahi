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
import "unsafe"

// Domain represents a domain name.
//
// As Multicast DNS allows domain labels to contain any valid
// UTF characters, when domain name is constructed from a sequence
// of labels, the proper escaping is required, so the dot character
// within a label will not be interpreted as label separator.
//
// Note, it affects only local representation of domain names.
// The wire representation
type Domain string

// DomainFrom makes a domain string from a sequence of labels.
//
// Labels are properly escaped but overall validity check is not
// performed (Avahi will do it for us when receive Domain as input).
//
// Note, this function is not guaranteed to escape labels exactly
// as Avahi does, but output is anyway correct.
func DomainFrom(labels []string) Domain {
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

	return Domain(buf)
}

// Split splits domain name into a sequence of labels.
//
// In a case of error it returns nil.
func (d Domain) Split() []string {
	// Convert input from Go to C
	in := C.CString(string(d))
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

// Equal reports if two Domain names are equal.
//
// Note, invalid domain names are never equal to anything
// else, including itself.
func (d Domain) Equal(d2 Domain) bool {
	labels1 := d.Split()
	labels2 := d2.Split()

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

// Normalize normalizes the domain name by removing unneeded escaping.
//
// In a case of error it returns empty string.
func (d Domain) Normalize() Domain {
	return DomainFrom(d.Split())
}
