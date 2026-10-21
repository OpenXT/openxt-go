// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Solutions, LLC
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file
//

package main

import (
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/openxt/openxt-go/argo"
	flag "github.com/spf13/pflag"
)

var (
	domid  = flag.IntP("domain", "d", 0, "destination domain id")
	port   = flag.IntP("port", "p", 5555, "destination port")
	listen = flag.BoolP("listen", "l", false, "listen for incoming connections")
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}

func sender(sockType, domid, port int) {
	s, err := argo.Dial(sockType, domid, port)
	if err != nil {
		die("err: argo connection error: %v", err)
	}

	_, copyErr := io.Copy(s, os.Stdin)
	closeErr := s.Close()
	if copyErr != nil {
		die("err: argo connection error: %v", copyErr)
	}
	if closeErr != nil {
		die("err: argo connection error: %v", closeErr)
	}
}

func listener(sockType, port int) {
	l, err := argo.Listen(sockType, port, argo.XEN_ARGO_DOMID_ANY)
	if err != nil {
		die("err: argo listen error: %v", err)
	}

	for {
		incoming, err := l.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "err: argo listen error:", err)
			continue
		}

		_, copyErr := io.Copy(os.Stdout, incoming)
		closeErr := incoming.Close()
		if copyErr != nil && copyErr != io.EOF && copyErr != io.ErrClosedPipe {
			fmt.Fprintln(os.Stderr, "err: argo connection error:", copyErr)
			continue
		}
		if closeErr != nil {
			fmt.Fprintln(os.Stderr, "err: argo connection error:", closeErr)
		}
	}
}

func main() {

	flag.Parse()

	if *listen {
		listener(syscall.SOCK_STREAM, *port)
	} else {
		sender(syscall.SOCK_STREAM, *domid, *port)
	}
}
