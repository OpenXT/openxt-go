//
// Copyright 2020 Apertus Soutions, LLC
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file
//

// Package argo provides a Go interface for interacting with the Argo character
// driver used by OpenXT.
package argo

// Common functions for Conn struct

import (
	"os"
)

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

