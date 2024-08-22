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

// DomainFrom makes a domain name string from a sequence of labels.
//
// Labels are properly escaped but overall validity check is not
// performed (Avahi will do it for us when receive Domain as input).
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

// DomainSplit splits domain name into a sequence of labels.
//
// In a case of error it returns nil.
func DomainSplit(d string) []string {
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
	labels1 := DomainSplit(d1)
	labels2 := DomainSplit(d2)

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
	return DomainFrom(DomainSplit(d))
}