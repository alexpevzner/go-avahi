# CGo binding for Avahi

[![godoc.org](https://godoc.org/github.com/alexpevzner/go-avahi?status.svg)](https://godoc.org/github.com/alexpevzner/go-avahi)
![GitHub](https://img.shields.io/github/license/alexpevzner/go-avahi)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexpevzner/go-avahi)](https://goreportcard.com/report/github.com/alexpevzner/go-avahi)

```
import "github.com/alexpevzner/go-avahi"
```

Package avahi provides a fairly complete CGo binding for Avahi client.

Avahi is the standard implementation of Multicast DNS and DNS-SD for Linux, and
likely for some BSD systems as well. This technology is essential for automatic
network configuration, service discovery on local networks, and driverless
printing and scanning. It also can be useful for the peer services discovery
in a cloud.

Please notice, there is an alternative Avahi binding in pure Go:

  - GitHub project: https://github.com/holoplot/go-avahi
  - The documentation: https://pkg.go.dev/github.com/holoplot/go-avahi

This package has the following key differences:

  - This is CGo binding, not pure Go
  - It uses native/stdlib types, where appropriate. For example,
    IP addresses returned as [netip.AddrPort]
  - It uses a single channel for all events reported by an object,
    so add/remove events cannot be reordered
  - It survives Avahi restart
  - Integer values, like various flags, DNS class and type and
    so own, have their own type, not a generic int16/int32
  - And the last but not least, it attempts to fill the gaps
    in Avahi documentation, which is not very detailed

This library is comprehensive, high-quality, and quite popular. It is possible
(and not very difficult) to implement MDNS/DNS-SD directly on top of it,
allowing the entire protocol to run within the user process without relying on
a system daemon like Avahi.

There are several existing implementations; however, I don't have experience
with them, so I can't provide a review.

One inherent disadvantage of all these implementations is that they do not work
with local services operating via the loopback network interface. MDNS is a
multicast-based protocol, and the loopback interface does not support
multicasting. System daemons like Avahi do not actually use multicasting for
loopback services; instead, they emulate the publishing and discovery
functionality for those services. An in-process implementation cannot achieve
this.

# Avahi documentation

[Avahi API documentation](https://avahi.org/doxygen/html/), to be
honest, is not easy to read. It lacks significant details and hard to
understand unless you have a lot of a-priory knowledge in the subject.

Among other things, this package attempts to fill this gap. As its
exported objects map very closely to the native C API objects (except
Go vs C naming conventions and using channels instead of callbacks),
[the package reference](https://godoc.org/github.com/alexpevzner/go-avahi)
may be useful as a generic Avahi API reference, regardless of
programming language you use.

So even if you are the C or Python programmer, you may find package
reference useful for you.

# Build requirements

This package requires Go 1.18 or newer. This is an easy requirement,
because Go 1.18 was released at March 2022, so must distros should
be up to date.

As it is CGo binding, it requires avahi-devel (or avahi-client, the
exact name may depend on your distro) package to be installed. On
most Linux distros it is an easy to achieve.

You will need also a working C compiler. This is easy in a case of
native build, but in a case of cross-compiling may require some
additional effort.

This package was developed and tested at Fedora 40, but expected
to work at all other distros.

# Runtime requirements

This package requires a working Avahi daemon and libavahi-client dynamic
libraries installed on a system. In most cases it should work out of
box.

# An Example

The following simple example demonstrates usage of the API provided by
this package. This simple program scans local network for available network
printers and outputs found devices.

```
// github.com/alexpevzner/go-avahi example

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/alexpevzner/go-avahi"
)

// checkErr terminates a program if err is not nil.
func checkErr(err error, format string, args ...any) {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		fmt.Printf("%s: %s\n", msg, err)
		os.Exit(1)
	}
}

// The main function.
func main() {
	// Create a Client with enabled workarounds for Avahi bugs
	clnt, err := avahi.NewClient(avahi.ClientLoopbackWorkarounds)
	checkErr(err, "avahi.NewClient")

	defer clnt.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create poller to simplify event loop.
	poller := avahi.NewPoller()

	// Create ServiceBrowsers for a variety of printer types
	svctypes := []string{
		"_printer._tcp",        // LPD protocol
		"_pdl-datastream._tcp", // HP JetDirect
		"_ipp._tcp",            // IPP over HTTPS
		"_ipps._tcp",           // IPP over HTTPS
	}

	for _, svctype := range svctypes {
		browser, err := avahi.NewServiceBrowser(
			clnt,
			avahi.IfIndexUnspec,
			avahi.ProtocolUnspec,
			svctype,
			"",
			avahi.LookupUseMulticast)

		checkErr(err, "browse %q", svctype)
		poller.AddServiceBrowser(browser)
	}

	// Wait for Browser events. Create resolvers on a fly.
	//
	// Here we use asynchronous API, so we can start resolvers
	// early to run in background.
	//
	// Run until we found/resolve all we expect or timeout occurs.
	wanted := make(map[string]struct{})
	for _, svctype := range svctypes {
		wanted[svctype] = struct{}{}
	}

	for ctx.Err() == nil && len(wanted) > 0 {
		evnt, _ := poller.Poll(ctx)

		switch evnt := evnt.(type) {
		case *avahi.ServiceBrowserEvent:
			switch evnt.Event {
			case avahi.BrowserNew:
				resolver, err := avahi.NewServiceResolver(
					clnt,
					evnt.IfIdx,
					evnt.Proto,
					evnt.InstanceName,
					evnt.SvcType,
					evnt.Domain,
					avahi.ProtocolUnspec,
					avahi.LookupUseMulticast)

				checkErr(err, "resolve %q", evnt.InstanceName)
				poller.AddServiceResolver(resolver)
				wanted[evnt.InstanceName] = struct{}{}

			case avahi.BrowserAllForNow:
				delete(wanted, evnt.SvcType)

			case avahi.BrowserFailure:
				err = evnt.Err
				checkErr(err, "browse %q", evnt.SvcType)
			}

		case *avahi.ServiceResolverEvent:
			switch evnt.Event {
			case avahi.ResolverFound:
				fmt.Printf("Found new device:\n"+
					"  Name:       %s:\n"+
					"  Type:       %s\n"+
					"  IP address: %s:%d\n",
					evnt.InstanceName,
					evnt.SvcType,
					evnt.Addr,
					evnt.Port)

				delete(wanted, evnt.InstanceName)

			case avahi.ResolverFailure:
				err = evnt.Err
				checkErr(err, "resolve %q", evnt.InstanceName)
			}
		}
	}
}
```

<!-- vim:ts=8:sw=4:et:textwidth=72
-->
