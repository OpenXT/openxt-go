//go:build !libargo
// +build !libargo

// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Solutions, LLC
//

package argo

import (
	"bytes"
	"syscall"
	"testing"
)

func TestAddrCodec(t *testing.T) {
	in := Addr{Port: 5555, Domain: 7}
	var buf bytes.Buffer
	if err := in.toC(&buf); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != xenArgoAddrSize {
		t.Fatalf("encoded address length = %d, want %d", buf.Len(), xenArgoAddrSize)
	}
	raw := append([]byte(nil), buf.Bytes()...)
	if raw[6] != 0 || raw[7] != 0 {
		t.Fatalf("address pad = %v, want zero", raw[6:])
	}

	var out Addr
	if err := addrFromC(&buf, &out); err != nil {
		t.Fatal(err)
	}
	if out.Port != in.Port || out.Domain != in.Domain {
		t.Fatalf("decoded address = %#v, want %#v", out, in)
	}
}

func TestRingIDCodec(t *testing.T) {
	in := RingId{Domain: XEN_ARGO_DOMID_ANY, Partner: 3, Port: 5555}
	var buf bytes.Buffer
	if err := in.toC(&buf); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != argoRingIdSize {
		t.Fatalf("encoded ring length = %d, want %d", buf.Len(), argoRingIdSize)
	}
}

func TestIoctlErr(t *testing.T) {
	if err := ioctlErr(0); err != nil {
		t.Fatalf("ioctlErr(0) = %v, want nil", err)
	}
	err := ioctlErr(syscall.EINVAL)
	errno, ok := err.(syscall.Errno)
	if !ok || errno != syscall.EINVAL {
		t.Fatalf("ioctlErr(EINVAL) = %T(%v), want syscall.Errno", err, err)
	}
}
