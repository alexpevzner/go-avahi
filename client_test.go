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

import (
	"fmt"
	"testing"
)

func TestClient(t *testing.T) {
	clnt, err := NewClient()
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	fmt.Printf("avahi version: %q\n", clnt.GetVersionString())
	fmt.Printf("Host name:     %q\n", clnt.GetHostName())
	fmt.Printf("Domain name:   %q\n", clnt.GetDomainName())
	fmt.Printf("Host FQDN:     %q\n", clnt.GetHostFQDN())

	state := <-clnt.Chan()
	println(state.String())

	browser, err := NewServiceBrowser(clnt,
		IfIndexUnspec, ProtocolUnspec,
		"_http._tcp", "", 0)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	for evnt := range browser.Chan() {
		fmt.Printf("%#v\n", evnt)
		if evnt.Event == BrowserAllForNow {
			break
		}
	}
}
