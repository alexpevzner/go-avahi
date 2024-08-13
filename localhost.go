// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Localhost handling
//
//go:build linux || freebsd

package avahi

import "strings"

// isLocalhost tells if hostname is localhost
func isLocalhost(hostname string) bool {
	ret := false

	switch strings.ToLower(hostname) {
	case "localhost", "localhost.localdomain":
		ret = true
	}

	return ret
}
