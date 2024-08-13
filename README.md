# CGo binding for Avahi

[![godoc.org](https://pkg.go.dev/github.com/alexpevzner/go-avahi?status.svg)](https://pkg.go.dev/github.com/alexpevzner/go-avahi)
![GitHub](https://img.shields.io/github/license/alexpevzner/go-avahi)

```
import "github.com/alexpevzner/go-avahi"
```

Package avahi provides a fairly complete CGo binding for [Avahi] client.

Avahi is the standard implementation of Multicast DNS and DNS-SD for Linux, and
likely for some BSD systems as well. This technology is essential for automatic
network configuration, service discovery on local networks, and driverless
printing and scanning.

Please notice, there is an alternative Avahi binding for Go:

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

# Build requirements

This package requires Go 1.18 or newer. This is an easy requirement,
because Go 1.18 was released at March 2022, so much distros should
be up to date.

As it is CGo binding, it requires avahi-devel (or avahi-client, the
exact name may depend on your distro) package to be installed. On
most Linux distros it is an easy to achieve.

You will need also a working C compiler. This is easy in a case of
native build, but in a case of cross-compiling may require some
additional effort.

This package was developed and tested at Fedora 40, but expected
to work at all other distros.

<!-- vim:ts=8:sw=4:et:textwidth=72
-->
