// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Avahi Client test
//
//go:build linux || freebsd

package avahi

import "testing"

func TestClient(t *testing.T) {
	clnt, err := NewClient()
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	state := <-clnt.Chan()
	println(state.String())
}
