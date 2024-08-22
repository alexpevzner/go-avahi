// CGo binding for Avahi
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
// On success, Browser is added to the Poller.
func addServiceTypeBrowser(poller *Poller, clnt *Client) error {
	browser, err := NewServiceTypeBrowser(
		clnt,
		MustLoopback(),
		ProtocolUnspec,
		"",
		LookupUseMulticast)

	if err != nil {
		return err
	}

	poller.AddServiceTypeBrowser(browser)
	return nil
}

// addServiceBrowser adds ServiceBrowser for the service defined by
// the ServiceBrowserEvent.
//
// On success, Browser is added to the Poller.
func addServiceBrowser(poller *Poller,
	clnt *Client, evnt *ServiceTypeBrowserEvent) error {

	browser, err := NewServiceBrowser(
		clnt,
		evnt.IfIdx,
		evnt.Proto,
		evnt.SvcType,
		evnt.Domain,
		LookupUseMulticast)

	if err != nil {
		return err
	}

	poller.AddServiceBrowser(browser)
	return nil
}

// addServiceResolver adds ServiceResolver for the service defined
// by the ServiceBrowserEvent.
//
// On success, Browser is added to the Poller.
func addServiceResolver(poller *Poller,
	clnt *Client, evnt *ServiceBrowserEvent) error {

	resolver, err := NewServiceResolver(
		clnt,
		evnt.IfIdx,
		evnt.Proto,
		evnt.InstanceName, evnt.SvcType, evnt.Domain,
		evnt.Proto,
		LookupUseMulticast)

	if err != nil {
		return err
	}

	poller.AddServiceResolver(resolver)
	return nil
}

// addAddressResolver adds AddressResolver for the address defined
// by the ServiceResolverEvent.
//
// On success, Browser is added to the Poller.
func addAddressResolver(poller *Poller,
	clnt *Client, evnt *ServiceResolverEvent) error {

	addr := evnt.Addr

	resolver, err := NewAddressResolver(
		clnt,
		evnt.IfIdx,
		evnt.Proto,
		addr,
		LookupUseMulticast)

	if err != nil {
		return err
	}

	poller.AddAddressResolver(resolver)
	return nil
}

// addHostNameResolver adds HostNameResolver for the hostname defined
// by the ServiceResolverEvent.
//
// On success, Browser is added to the Poller.
func addHostNameResolver(poller *Poller,
	clnt *Client, evnt *ServiceResolverEvent) error {

	resolver, err := NewHostNameResolver(
		clnt,
		evnt.IfIdx,
		evnt.Proto,
		evnt.FQDN(),
		evnt.Proto,
		LookupUseMulticast)

	if err != nil {
		return err
	}

	poller.AddHostNameResolver(resolver)
	return nil
}

// TestAvahi performs overall test of most of Avahi API.
func TestAvahi(t *testing.T) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare test data
	loopback := MustLoopback()
	instancename := randName()
	hostname := randName() + ".local"
	svctype := "_avahi-test._tcp"

	services := []*EntryGroupService{
		{
			IfIdx:        loopback,
			Proto:        ProtocolUnspec,
			InstanceName: instancename,
			SvcType:      svctype,
			Domain:       "",
			Hostname:     "",
			Port:         0,
			Txt: []string{
				"foo=bar",
			},
		},
	}

	addresses := []*EntryGroupAddress{
		{
			IfIdx:    loopback,
			Proto:    ProtocolUnspec,
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
		expectServiceTypeBrowser[svc.SvcType] = true
		expectServiceBrowser[svc.SvcType] = true
		expectServiceResolver[svc.InstanceName] = true
	}

	expectAddressResolver["localhost.localdomain"] = true
	expectHostNameResolver["127.0.0.1"] = true

	// Resolve everything we've just published
	poller := NewPoller()
	err = addServiceTypeBrowser(poller, clnt)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	for err == nil && expectCount() > 0 {
		var evnt any
		evnt, err = poller.Poll(ctx)

		switch evnt := evnt.(type) {
		case *ServiceTypeBrowserEvent:
			t.Logf("%#v", evnt)
			switch evnt.Event {
			case BrowserNew:
				if evnt.IfIdx != loopback {
					continue
				}

				n := evnt.SvcType
				if expectServiceTypeBrowser[n] {
					delete(expectServiceTypeBrowser, n)
					err = addServiceBrowser(poller,
						clnt, evnt)
				}
			case BrowserFailure:
				err = evnt.Err
			}

		case *ServiceBrowserEvent:
			t.Logf("%#v", evnt)
			switch evnt.Event {
			case BrowserNew:
				if evnt.IfIdx != loopback {
					continue
				}

				n := evnt.SvcType
				if expectServiceBrowser[n] {
					delete(expectServiceBrowser, n)
					err = addServiceResolver(poller,
						clnt, evnt)
				}
			case BrowserFailure:
				err = evnt.Err
			}

		case *ServiceResolverEvent:
			t.Logf("%#v", evnt)
			t.Logf("%s:%d", evnt.Addr, evnt.Port)

			switch evnt.Event {
			case ResolverFound:
				if evnt.IfIdx != loopback {
					continue
				}

				n := evnt.InstanceName
				if expectServiceResolver[n] {
					delete(expectServiceResolver, n)

					err = addAddressResolver(poller,
						clnt, evnt)
					if err == nil {
						err = addHostNameResolver(
							poller, clnt, evnt)
					}
				}

			case ResolverFailure:
				err = evnt.Err
			}

		case *AddressResolverEvent:
			t.Logf("%#v", evnt)

			switch evnt.Event {
			case ResolverFound:
				if evnt.IfIdx != loopback {
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
				if evnt.IfIdx != loopback {
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
