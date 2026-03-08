// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Soutions, LLC
//

package main

import (
	"fmt"
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/openxt/openxt-go/db"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}

func usage() {
	die(
		`Usage: dbcmd <command> [<args>]

Available commands are:
  read <key>		Retrieve <key> from db
  write <key> <value>	Store <value> for <key> in the db
  rm <key>		Delete <key> from db
  exists <key>		Check if <key> exist in the db
  help			Print this help`)
}

func main() {
	arglen := len(os.Args)
	if arglen < 2 {
		usage()
	}

	conn, err := dbus.SystemBus()
	if err != nil {
		die("Error connecting to system bus: %v\n", err)
	}
	defer conn.Close()

	db := db.NewDbClient(conn, db.DbServiceName, "/")

	operation := os.Args[1]

	switch operation {
	case "read":
		if arglen != 3 {
			fmt.Fprintf(os.Stderr,
				"Error: incorrect number of arguments.\n")
			usage()
		}
		result, err := db.Read(os.Args[2])
		if err != nil {
			die("DB read error: %v", err)
		}
		fmt.Println(os.Stdout, "%s", result)
	case "write":
		if arglen != 4 {
			fmt.Fprintf(os.Stderr,
				"Error: incorrect number of arguments.\n")
			usage()
		}
		err := db.Write(os.Args[2], os.Args[3])
		if err != nil {
			die("DB write error: %v", err)
		}
	case "rm":
		if arglen != 3 {
			fmt.Fprintf(os.Stderr,
				"Error: incorrect number of arguments.\n")
			usage()
		}
		err := db.Rm(os.Args[2])
		if err != nil {
			die("DB rm error: %v", err)
		}
	case "exists":
		if arglen != 3 {
			fmt.Fprintf(os.Stderr,
				"Error: incorrect number of arguments.\n")
			usage()
		}
		result, err := db.Exists(os.Args[2])
		if err != nil {
			die("DB exists error: %v", err)
		}
		fmt.Println(os.Stdout, "%t", result)
	default:
		usage()
	}
}
