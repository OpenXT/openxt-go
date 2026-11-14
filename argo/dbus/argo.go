// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Solutions, LLC
//

package dbus

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	godbus "github.com/godbus/dbus/v5"
	"github.com/openxt/openxt-go/argo"
)

const (
	PlatformBusAddress = "argo:domain=0,port=5555"
	SystemBusAddress   = "unix:path=/var/run/dbus/system_bus_socket"
)

// ConnectPlatformBus opens a D-Bus connection to Dom0 on Argo port 5555.
// DBUS_SYSTEM_BUS_ADDRESS overrides that address when it is set.
func ConnectPlatformBus(opts ...godbus.ConnOption) (*godbus.Conn, error) {
	address := os.Getenv("DBUS_SYSTEM_BUS_ADDRESS")
	if address == "" {
		address = PlatformBusAddress
	}

	return Connect(address, opts...)
}

// Connect expands godbus.Dial to accept an Argo address. An Argo address has
// the form argo:domain=<id>,port=<number>. domid is accepted as an alias for
// domain.
func Connect(address string, opts ...godbus.ConnOption) (*godbus.Conn, error) {
	i := strings.IndexRune(address, ':')
	if i == -1 {
		return nil, errors.New("dbus: invalid bus address (no transport)")
	}

	var conn *godbus.Conn
	var err error
	if address[:i] == "argo" {
		domid, port, perr := parseArgoAddress(address[i+1:])
		if perr != nil {
			return nil, perr
		}

		ac, derr := argo.Dial(syscall.SOCK_STREAM, domid, port)
		if derr != nil {
			return nil, derr
		}
		conn, err = godbus.NewConn(ac.File(), opts...)
		if err != nil {
			ac.Close()
			return nil, err
		}
	} else {
		conn, err = godbus.Dial(address, opts...)
		if err != nil {
			return nil, err
		}
	}

	// Should pass conn.auth but it is private and cannot access
	if err = conn.Auth([]godbus.Auth{godbus.AuthAnonymous(), godbus.AuthExternal("root")}); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err = conn.Hello(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

// parseArgoAddress parses the portion of an Argo bus address after "argo:".
func parseArgoAddress(spec string) (domid, port int, err error) {
	var haveDomain, havePort bool

	for _, field := range strings.Split(spec, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		eq := strings.IndexRune(field, '=')
		if eq <= 0 || eq == len(field)-1 {
			return 0, 0, fmt.Errorf("dbus: invalid argo address field %q", field)
		}
		key := strings.TrimSpace(field[:eq])
		raw := strings.TrimSpace(field[eq+1:])
		n, nerr := strconv.Atoi(raw)
		if nerr != nil {
			return 0, 0, fmt.Errorf("dbus: invalid argo %s %q", key, raw)
		}
		if n < 0 {
			return 0, 0, fmt.Errorf("dbus: invalid argo %s %d", key, n)
		}

		switch key {
		case "domain", "domid":
			if haveDomain {
				return 0, 0, errors.New("dbus: duplicate argo domain")
			}
			if n > 0xFFFF {
				return 0, 0, fmt.Errorf("dbus: invalid argo domain %d", n)
			}
			haveDomain = true
			domid = n
		case "port":
			if havePort {
				return 0, 0, errors.New("dbus: duplicate argo port")
			}
			if uint64(n) > 0xFFFFFFFF {
				return 0, 0, fmt.Errorf("dbus: invalid argo port %d", n)
			}
			havePort = true
			port = n
		default:
			return 0, 0, fmt.Errorf("dbus: unknown argo address key %q", key)
		}
	}

	if !haveDomain || !havePort {
		return 0, 0, errors.New("dbus: argo address requires domain and port")
	}
	return domid, port, nil
}
