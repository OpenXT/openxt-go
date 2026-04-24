// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Soutions, LLC
//

package db

import (
	"github.com/godbus/dbus/v5"
)

const DbServiceName = "com.citrix.xenclient.db"

type DbClient struct {
	dbus.BusObject
}

func NewDbClient(conn *dbus.Conn, dest, path string) *DbClient {
	return &DbClient{conn.Object(dest, dbus.ObjectPath(path))}
}

/* Interface com.citrix.xenclient.db */
func (d *DbClient) Dump(path string) (value string, err error) {

	err = d.Call("com.citrix.xenclient.db.dump", 0, path).Store(&value)

	return
}

func (d *DbClient) Exists(path string) (ex bool, err error) {

	err = d.Call("com.citrix.xenclient.db.exists", 0, path).Store(&ex)

	return
}

func (d *DbClient) Inject(path string, value string) error {

	call := d.Call("com.citrix.xenclient.db.inject", 0, path, value)

	return call.Err
}

func (d *DbClient) List(path string) (value []string, err error) {

	err = d.Call("com.citrix.xenclient.db.list", 0, path).Store(&value)

	return
}

func (d *DbClient) Read(path string) (value string, err error) {

	err = d.Call("com.citrix.xenclient.db.read", 0, path).Store(&value)

	return
}

func (d *DbClient) ReadBinary(path string) (value []byte, err error) {

	err = d.Call("com.citrix.xenclient.db.read_binary", 0, path).Store(&value)

	return
}

func (d *DbClient) Rm(path string) error {

	call := d.Call("com.citrix.xenclient.db.rm", 0, path)

	return call.Err
}

func (d *DbClient) Write(path string, value string) error {

	call := d.Call("com.citrix.xenclient.db.write", 0, path, value)

	return call.Err
}
