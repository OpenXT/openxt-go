//
// Copyright 2020 Apertus Soutions, LLC
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file
//

package argo

// Common functions for Conn struct

import (
	"os"
)

func (c *Conn) File() *os.File {
	return c.file
}

func (c *Conn) Fd() uintptr {
	return c.file.Fd()
}

func (c *Conn) Read(p []byte) (int, error) {
	return c.file.Read(p)
}

func (c *Conn) Write(p []byte) (int, error) {
	return c.file.Write(p)
}

func (c *Conn) Close() error {
	return c.file.Close()
}

