// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Solutions, LLC
//

package dbus

import "testing"

func TestParseArgoAddress(t *testing.T) {
	domid, port, err := parseArgoAddress("domain=0,port=5555")
	if err != nil {
		t.Fatal(err)
	}
	if domid != 0 || port != 5555 {
		t.Fatalf("parseArgoAddress = (%d, %d), want (0, 5555)", domid, port)
	}

	domid, port, err = parseArgoAddress(" domid = 4 , port = 9 ")
	if err != nil {
		t.Fatal(err)
	}
	if domid != 4 || port != 9 {
		t.Fatalf("parseArgoAddress alias = (%d, %d), want (4, 9)", domid, port)
	}
}

func TestParseArgoAddressRejects(t *testing.T) {
	cases := []string{
		"",
		"domain",
		"domain=",
		"=1",
		"domain=x,port=1",
		"domain=-1,port=1",
		"domain=1,port=-5",
		"domain=65536,port=1",
		"port=1",
		"domain=1",
		"foo=1,port=1",
		"domain=1,port=1,domain=2",
		"domain=1,domid=2,port=1",
	}
	for _, spec := range cases {
		if _, _, err := parseArgoAddress(spec); err == nil {
			t.Fatalf("parseArgoAddress(%q) succeeded, want error", spec)
		}
	}
}

func TestConnectRejectsBadArgoAddress(t *testing.T) {
	cases := []string{
		"argo",
		"argo:",
		"argo:domain=nope,port=1",
		"argo:port=1",
	}
	for _, address := range cases {
		if _, err := Connect(address); err == nil {
			t.Fatalf("Connect(%q) succeeded, want error", address)
		}
	}
}
