// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Soutions, LLC
//

package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/godbus/dbus/v5"
	argoDbus "github.com/openxt/openxt-go/argo/dbus"
	"github.com/openxt/openxt-go/db"
	flag "github.com/spf13/pflag"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s <command> [<args>]:\n", os.Args[0])
	flag.PrintDefaults()
	die(`
Available commands are:
  cat <key>		Dump raw value for <key>
  exists <key>		Check if <key> exist
  ls <key>		List tree start at <key>
  nodes <key>		List immediate childtren of <key>
  read <key>		Retrieve string value for <key>
  rm <key>		Delete <key>
  write <key> <value>	Store <value> for <key>`)
}

func list(c *db.DbClient, fullPath bool, indent int, path string) (string, error) {
	path = strings.TrimRight(path, db.PathDelimiter)
	result, err := c.List(path)
	if err != nil {
		return "", err
	}

	var key string
	if fullPath {
		key = path
	} else {
		key = strings.Repeat(" ", indent)
		if path != "" {
			key += filepath.Base(path)
		}
	}

	if len(result) == 0 {
		value, err := c.Read(path)
		if err != nil {
			return "", fmt.Errorf("failed reading %s: %v\n", path, err)
		}
		return fmt.Sprintf("%s = \"%s\"", key, value), nil
	}

	out := key + " ="
	for _, elem := range result {
		r, err := list(c, fullPath, indent+1, path+"/"+elem)
		if err != nil {
			return "", err
		}

		out += "\n" + r
	}

	return out, nil
}

func main() {
	var conn *dbus.Conn

	helpFlag := flag.BoolP("help", "h", false, "Print help")
	fullPathFlag := flag.BoolP("full", "f", false, "Full path")
	platBusFlag := flag.BoolP("platform", "p", false, "Connect to the platform bus")
	flag.CommandLine.MarkHidden("full")
	flag.Parse()

	if *helpFlag {
		usage()
	}

	if *platBusFlag {
		var err error
		conn, err = argoDbus.ConnectPlatformBus()
		if err != nil {
			die("Error connecting to platform bus: %v\n", err)
		}
	} else {
		var err error
		conn, err = dbus.SystemBus()
		if err != nil {
			die("Error connecting to system bus: %v\n", err)
		}
	}
	defer conn.Close()

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	client := db.NewDbClient(conn, db.DbServiceName, "/")

	operation := args[0]

	args = args[1:]
	arglen := len(args)

	switch operation {
	case "cat":
		if arglen != 1 {
			fmt.Fprintf(os.Stderr,
				"Error: incorrect number of arguments.\n")
			usage()
		}
		result, err := client.ReadBinary(args[1])
		if err != nil {
			die("DB read binary error: %v", err)
		}
		binary.Write(os.Stdout, binary.LittleEndian, result)

	case "exists":
		if arglen != 1 {
			fmt.Fprintf(os.Stderr,
				"Error: incorrect number of arguments.\n")
			usage()
		}
		result, err := client.Exists(os.Args[0])
		if err != nil {
			die("DB exists error: %v", err)
		}
		fmt.Printf("%t", result)
	case "ls":
		path := "/"
		if len(args) != 0 {
			path = args[0]
		}
		result, err := list(client, *fullPathFlag, 0, path)
		if err != nil {
			die("DB list error: %v", err)
		}
		fmt.Printf("%s\n", result)

	case "nodes":
		if arglen != 1 {
			fmt.Fprintf(os.Stderr,
				"Error: incorrect number of arguments.\n")
			usage()
		}
		result, err := client.List(args[0])
		if err != nil {
			die("DB read error: %v", err)
		}
		fmt.Printf("%s\n", strings.Join(result, " "))
	case "read":
		if arglen != 1 {
			fmt.Fprintf(os.Stderr,
				"Error: incorrect number of arguments.\n")
			usage()
		}
		result, err := client.Read(args[0])
		if err != nil {
			die("DB read error: %v", err)
		}
		fmt.Printf("%s\n", result)
	case "rm":
		if arglen != 1 {
			fmt.Fprintf(os.Stderr,
				"Error: incorrect number of arguments.\n")
			usage()
		}
		err := client.Rm(args[0])
		if err != nil {
			die("DB rm error: %v", err)
		}
	case "write":
		if arglen != 2 {
			fmt.Fprintf(os.Stderr,
				"Error: incorrect number of arguments.\n")
			usage()
		}
		err := client.Write(args[0], args[1])
		if err != nil {
			die("DB write error: %v", err)
		}
	default:
		usage()
	}
}
