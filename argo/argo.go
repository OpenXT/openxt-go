// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Solutions, LLC
//

// Package argo provides a Go interface for interacting with the Argo character
// driver used by OpenXT.
package argo

import (
	"os"
)

const (
	XEN_ARGO_MAX_RING_SIZE = 0x1000000
	XEN_ARGO_DOMID_ANY     = 0x7FF4
	XEN_ARGO_PORT_ANY      = 0xFFFFFFFF
)

// Common functions for Conn struct

// Returns an os.File pointer that wraps the underlying file descriptor for the
// connection.
func (c *Conn) File() *os.File {
	return c.file
}

// Returns the file descriptor for the connection.
func (c *Conn) Fd() uintptr {
	return c.file.Fd()
}

// Read implements the standard Read interface: it reads data from the
// connection. If the remote end is closed with an error, that error is
// returned as err; otherwise err is EOF.
func (c *Conn) Read(p []byte) (int, error) {
	return c.file.Read(p)
}

// Write implements the standard Write interface: it writes data to the
// connection. If the remote end is closed with an error, that err is returned
// as err.
func (c *Conn) Write(p []byte) (int, error) {
	return c.file.Write(p)
}

// Closes the connection.
func (c *Conn) Close() error {
	return c.file.Close()
}

