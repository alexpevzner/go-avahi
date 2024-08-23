// CGo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Domain name test
//
//go:build linux || freebsd

package avahi

import (
	"reflect"
	"testing"
)

// TestDomainFrom tests DomainFrom function
func TestDomainFrom(t *testing.T) {
	type testData struct {
		labels []string
		domain string
	}

	tests := []testData{
		{
			labels: []string{`example`, `com`},
			domain: `example.com`,
		},

		{
			labels: []string{`My.Service`, `example`, `com`},
			domain: `My\.Service.example.com`,
		},

		{
			labels: []string{`My\Service`, `example`, `com`},
			domain: `My\\Service.example.com`,
		},
	}

	for _, test := range tests {
		domain := DomainFrom(test.labels)
		if domain != test.domain {
			t.Errorf("%q:\n"+
				"expected: %q\n"+
				"present:  %q\n",
				test.labels, test.domain, domain)
		}
	}
}

// TestDomainSlice tests DomainSlice function
func TestDomainSlice(t *testing.T) {
	type testData struct {
		domain string
		labels []string
	}

	tests := []testData{
		{
			domain: `example.com`,
			labels: []string{"example", "com"},
		},

		{
			domain: `ex\?ample.com`,
			labels: nil,
		},
	}

	for _, test := range tests {
		labels := DomainSlice(test.domain)
		if !reflect.DeepEqual(labels, test.labels) {
			t.Errorf("%q:\n"+
				"expected: %q\n"+
				"present:  %q\n",
				test.domain, test.labels, labels)
		}
	}
}

// TestDomainEqual tests DomainEqual function
func TestDomainEqual(t *testing.T) {
	type testData struct {
		d1, d2 string
		equal  bool
	}

	tests := []testData{
		{
			// Equal domains
			d1:    `example.com`,
			d2:    `example.com`,
			equal: true,
		},
		{
			// Different domains
			d1:    `www.example.com`,
			d2:    `xxx.example.com`,
			equal: false,
		},
		{
			// Different number of labels
			d1:    `www.example.com`,
			d2:    `example.com`,
			equal: false,
		},
		{
			// ASCII case must be ignored
			d1:    `ExAmPlE.CoM`,
			d2:    `eXaMpLe.cOm`,
			equal: true,
		},
		{
			// UTF-8 must not be a problem
			d1:    `пример.example.com`,
			d2:    `пример.example.com`,
			equal: true,
		},
		{
			// UTF-8 case must not be ignored
			d1:    `пример.example.com`,
			d2:    `ПРИМЕР.example.com`,
			equal: false,
		},
		{
			// Invalid domains are never equal
			d1:    `ex\?ample.com`,
			d2:    `ex\?ample.com`,
			equal: false,
		},
	}

	for _, test := range tests {
		equal := DomainEqual(test.d1, test.d2)
		if equal != test.equal {
			t.Errorf("%q vs %q:\n"+
				"expected: %v\n"+
				"present:  %v\n",
				test.d1, test.d2, test.equal, equal)
		}
	}
}

// TestDomainNormalize tests DomainNormalize function
func TestDomainNormalize(t *testing.T) {
	type testData struct {
		in, out string
	}

	tests := []testData{
		{
			in:  `example.com`,
			out: `example.com`,
		},
		{
			in:  `ex\?ample.com`,
			out: ``,
		},
		{
			in:  `ex ample.com`,
			out: `ex ample.com`,
		},
	}

	for _, test := range tests {
		out := DomainNormalize(test.in)
		if out != test.out {
			t.Errorf("%q:\n"+
				"expected: %q\n"+
				"present:  %q\n",
				test.in, test.out, out)
		}
	}
}

// TestDomainToLowerUpper tests DomainToLower and DomainToUpper function
func TestDomainToLowerUpper(t *testing.T) {
	type testData struct {
		domain       string
		lower, upper string
	}

	tests := []testData{
		{
			domain: `1.2.3.example.com`,
			lower:  `1.2.3.example.com`,
			upper:  `1.2.3.EXAMPLE.COM`,
		},
		{
			domain: `1.2.3.EXAMPLE.COM`,
			lower:  `1.2.3.example.com`,
			upper:  `1.2.3.EXAMPLE.COM`,
		},
		{
			domain: `привет.example.com`,
			lower:  `привет.example.com`,
			upper:  `привет.EXAMPLE.COM`,
		},
		{
			domain: `ПРИВЕТ.example.com`,
			lower:  `ПРИВЕТ.example.com`,
			upper:  `ПРИВЕТ.EXAMPLE.COM`,
		},
	}

	for _, test := range tests {
		lower := DomainToLower(test.domain)
		upper := DomainToUpper(test.domain)

		if lower != test.lower {
			t.Errorf("DomainToLower(%q):\n"+
				"expected: %q\n"+
				"present:  %q\n",
				test.domain, test.lower, lower)
		}

		if upper != test.upper {
			t.Errorf("DomainToUpper(%q):\n"+
				"expected: %q\n"+
				"present:  %q\n",
				test.domain, test.upper, upper)
		}
	}
}
