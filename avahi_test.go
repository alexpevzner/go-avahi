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
	"context"
	"crypto/rand"
	"fmt"
	"net/netip"
	"reflect"
	"sort"
	"testing"
	"time"
)

// randName creates a random name
func randName() string {
	// Create random name
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%x", buf)
}

// addServiceTypeBrowser adds ServiceTypeBrowser
//
// On success, Browser's event channel added into the
// list of channels, represented by []reflect.SelectCase.
func addServiceTypeBrowser(cases *[]reflect.SelectCase, clnt *Client) error {
	browser, err := NewServiceTypeBrowser(
		clnt,
		MustLoopback(),
		ProtocolUnspec,
		"",
		LookupUseMulticast)

	if err != nil {
		return err
	}

	*cases = append(*cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(browser.Chan()),
	})

	return nil
}

// addServiceBrowser adds ServiceBrowser for the service defined by
// the ServiceBrowserEvent.
//
// On success, Browser's event channel added into the
// list of channels, represented by []reflect.SelectCase.
func addServiceBrowser(cases *[]reflect.SelectCase,
	clnt *Client, evnt *ServiceTypeBrowserEvent) error {

	browser, err := NewServiceBrowser(
		clnt,
		evnt.IfIndex,
		evnt.Protocol,
		evnt.Type,
		evnt.Domain,
		LookupUseMulticast)

	if err != nil {
		return err
	}

	*cases = append(*cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(browser.Chan()),
	})

	return nil
}

// addServiceResolver adds ServiceResolver for the service defined
// by the ServiceBrowserEvent.
//
// On success, Browser's event channel added into the
// list of channels, represented by []reflect.SelectCase.
func addServiceResolver(cases *[]reflect.SelectCase,
	clnt *Client, evnt *ServiceBrowserEvent) error {

	resolver, err := NewServiceResolver(
		clnt,
		evnt.IfIndex,
		evnt.Protocol,
		evnt.InstanceName, evnt.Type, evnt.Domain,
		evnt.Protocol,
		LookupUseMulticast)

	if err != nil {
		return err
	}

	*cases = append(*cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(resolver.Chan()),
	})

	return nil
}

// addAddressResolver adds AddressResolver for the address defined
// by the ServiceResolverEvent.
//
// On success, Browser's event channel added into the
// list of channels, represented by []reflect.SelectCase.
func addAddressResolver(cases *[]reflect.SelectCase,
	clnt *Client, evnt *ServiceResolverEvent) error {

	addr := evnt.AddrPort.Addr()
	addr = netip.MustParseAddr("127.0.0.1")

	resolver, err := NewAddressResolver(
		clnt,
		evnt.IfIndex,
		evnt.Protocol,
		addr,
		LookupUseMulticast)

	if err != nil {
		return err
	}

	*cases = append(*cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(resolver.Chan()),
	})

	return nil
}

// addHostNameResolver adds HostNameResolver for the hostname defined
// by the ServiceResolverEvent.
//
// On success, Browser's event channel added into the
// list of channels, represented by []reflect.SelectCase.
func addHostNameResolver(cases *[]reflect.SelectCase,
	clnt *Client, evnt *ServiceResolverEvent) error {

	resolver, err := NewHostNameResolver(
		clnt,
		evnt.IfIndex,
		evnt.Protocol,
		evnt.FQDN(),
		evnt.Protocol,
		LookupUseMulticast)

	if err != nil {
		return err
	}

	*cases = append(*cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(resolver.Chan()),
	})

	return nil
}

// TestAvahi performs overall test of most of Avahi API.
func TestAvahi(t *testing.T) {
	// Create context with timeout
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	// Prepare test data
	loopback := MustLoopback()
	instancename := randName()
	hostname := randName() + ".local"
	svctype := "_avahi-test._tcp"

	services := []*EntryGroupService{
		&EntryGroupService{
			EntryGroupServiceIdent: EntryGroupServiceIdent{
				IfIndex:      loopback,
				Protocol:     ProtocolUnspec,
				InstanceName: instancename,
				Type:         svctype,
				Domain:       "",
			},
			Hostname: "",
			Port:     0,
			Txt: []string{
				"foo=bar",
			},
		},
	}

	addresses := []*EntryGroupAddress{
		{
			IfIndex:  loopback,
			Protocol: ProtocolUnspec,
			Hostname: hostname,
			Addr:     netip.MustParseAddr("127.0.0.1"),
		},
	}

	// Create a client
	clnt, err := NewClient(ClientLoopbackWorkarounds)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	defer clnt.Close()

	// Wait until ClientStateRunning
	evntClnt, err := clnt.Get(ctx)
	for evntClnt != nil && evntClnt.State != ClientStateRunning {
		t.Logf("%#v\n", evntClnt)
		evntClnt, err = clnt.Get(ctx)
	}

	if err != nil {
		t.Errorf("%s", err)
		return
	}

	// Create EntryGroup
	egrp, err := NewEntryGroup(clnt)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	// Create test entries
	for _, ent := range services {
		err = egrp.AddService(ent, 0)
		if err != nil {
			t.Errorf("%s", err)
			return
		}
	}

	for _, ent := range addresses {
		err = egrp.AddAddress(ent, PublishNoReverse)
		if err != nil {
			t.Errorf("%s", err)
			return
		}
	}

	// Commit the EntryGroup
	err = egrp.Commit()
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	// Wait until EntryGroupStateEstablished
	egrpEvnt, err := egrp.Get(ctx)
	for egrpEvnt != nil && egrpEvnt.State != EntryGroupStateEstablished {
		t.Logf("%#v\n", egrpEvnt)
		egrpEvnt, err = egrp.Get(ctx)
	}

	if err != nil {
		t.Errorf("%s", err)
		return
	}

	// Prepare to keep track on resolved entries
	expectAddressResolver := make(map[string]bool)
	expectHostNameResolver := make(map[string]bool)
	expectServiceBrowser := make(map[string]bool)
	expectServiceResolver := make(map[string]bool)
	expectServiceTypeBrowser := make(map[string]bool)

	//expectServiceTypeBrowser["not-found"] = true

	expectCount := func() int {
		return len(expectAddressResolver) +
			len(expectHostNameResolver) +
			len(expectServiceBrowser) +
			len(expectServiceResolver) +
			len(expectServiceTypeBrowser)
	}

	expectMissed := func() (missed []string) {
		expect := []struct {
			n string
			m map[string]bool
		}{
			{"AddressResolver", expectAddressResolver},
			{"HostNameResolver", expectHostNameResolver},
			{"ServiceBrowser", expectServiceBrowser},
			{"ServiceResolver", expectServiceResolver},
			{"ServiceTypeBrowser", expectServiceTypeBrowser},
		}

		for _, exp := range expect {
			if len(exp.m) > 0 {
				missed = append(missed, exp.n+":")
				names := make([]string, 0, len(exp.m))
				for name := range exp.m {
					names = append(names,
						fmt.Sprintf("  %q", name))
				}
				sort.Slice(names, func(i, j int) bool {
					return names[i] < names[j]
				})
				missed = append(missed, names...)
			}
		}

		return
	}

	for _, svc := range services {
		expectServiceTypeBrowser[svc.Type] = true
		expectServiceBrowser[svc.Type] = true
		expectServiceResolver[svc.InstanceName] = true
	}

	expectAddressResolver["localhost.localdomain"] = true
	expectHostNameResolver["127.0.0.1"] = true

	// Resolve everything we've just published
	cases := []reflect.SelectCase{
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		},
	}

	err = addServiceTypeBrowser(&cases, clnt)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	for err == nil && expectCount() > 0 {

		_, recv, _ := reflect.Select(cases)

		switch evnt := recv.Interface().(type) {
		case struct{}:
			err = ctx.Err()

		case *ServiceTypeBrowserEvent:
			t.Logf("%#v", evnt)
			switch evnt.Event {
			case BrowserNew:
				if evnt.IfIndex != loopback {
					continue
				}

				n := evnt.Type
				if expectServiceTypeBrowser[n] {
					delete(expectServiceTypeBrowser, n)
					err = addServiceBrowser(&cases,
						clnt, evnt)
				}
			case BrowserFailure:
				err = evnt.Err
			}

		case *ServiceBrowserEvent:
			t.Logf("%#v", evnt)
			switch evnt.Event {
			case BrowserNew:
				if evnt.IfIndex != loopback {
					continue
				}

				n := evnt.Type
				if expectServiceBrowser[n] {
					delete(expectServiceBrowser, n)
					err = addServiceResolver(&cases,
						clnt, evnt)
				}
			case BrowserFailure:
				err = evnt.Err
			}

		case *ServiceResolverEvent:
			t.Logf("%#v", evnt)
			t.Logf("%s", evnt.AddrPort)

			switch evnt.Event {
			case ResolverFound:
				if evnt.IfIndex != loopback {
					continue
				}

				n := evnt.InstanceName
				if expectServiceResolver[n] {
					delete(expectServiceResolver, n)

					err = addAddressResolver(&cases,
						clnt, evnt)
					if err == nil {
						err = addHostNameResolver(&cases,
							clnt, evnt)
					}
				}

			case ResolverFailure:
				err = evnt.Err
			}

		case *AddressResolverEvent:
			t.Logf("%#v", evnt)

			switch evnt.Event {
			case ResolverFound:
				if evnt.IfIndex != loopback {
					continue
				}

				n := evnt.Hostname
				if expectAddressResolver[n] {
					delete(expectAddressResolver, n)
				}

			case ResolverFailure:
				err = evnt.Err
			}

		case *HostNameResolverEvent:
			t.Logf("%#v", evnt)
			t.Logf("%s", evnt.Addr)
			t.Logf("%s", evnt.Flags)

			switch evnt.Event {
			case ResolverFound:
				if evnt.IfIndex != loopback {
					continue
				}

				n := evnt.Addr.String()
				if expectHostNameResolver[n] {
					delete(expectHostNameResolver, n)
				}

			case ResolverFailure:
				err = evnt.Err
			}
		}
	}

	if missed := expectMissed(); missed != nil {
		s := ""
		for _, ent := range missed {
			s += "\n" + ent
		}
		t.Errorf("unresolved entries:%s", s)
	}

	if err != nil {
		t.Errorf("%s", err)
		return
	}

	// Close the Client
	clnt.Close()
	evntClnt, err = clnt.Get(ctx)
	for evntClnt != nil {
		t.Logf("%#v\n", evntClnt)
		evntClnt, err = clnt.Get(ctx)
	}

	if err != nil {
		t.Errorf("%s", err)
		return
	}

	// EntryGroup must die
	egrpEvnt, err = egrp.Get(ctx)
	for egrpEvnt != nil && egrpEvnt.State != EntryGroupStateFailure {
		t.Logf("%#v\n", egrpEvnt)
		egrpEvnt, err = egrp.Get(ctx)
	}

	if err != nil {
		t.Errorf("%s", err)
		return
	}
}
