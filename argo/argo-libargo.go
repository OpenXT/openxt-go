//go:build libargo
// +build libargo

// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Solutions, LLC
//

package argo

// #cgo LDFLAGS: -largo
// #include <libargo.h>
import "C"

import (
	"errors"
	"os"
	"syscall"
)

func socketName(sockType int) (string, error) {
	switch sockType {
	case syscall.SOCK_STREAM:
		return "/dev/argo_stream", nil
	case syscall.SOCK_DGRAM:
		return "/dev/argo_dgram", nil
	default:
		return "", errors.New("unsupported socket type")
	}
}

// callErrno returns the cgo errno when a libargo call fails. A negative
// return is a failure even if errno was not set. A non-negative return is
// success; a stale errno is ignored.
func callErrno(rv C.int, err error) error {
	if rv >= 0 {
		return nil
	}
	if err != nil {
		return err
	}
	return errors.New("argo call failed")
}

func closeFD(fd C.int) {
	C.argo_close(fd)
}

func newConn(fd C.int, sockType int, addr Addr) (*Conn, error) {
	name, err := socketName(sockType)
	if err != nil {
		closeFD(fd)
		return nil, err
	}
	if err := syscall.SetNonblock(int(fd), false); err != nil {
		closeFD(fd)
		return nil, err
	}
	f := os.NewFile(uintptr(fd), name)
	if f == nil {
		closeFD(fd)
		return nil, errors.New("invalid argo descriptor")
	}
	return &Conn{file: f, addr: addr}, nil
}

// Dial initiates an Argo connection.
func Dial(sockType, domid, port int) (*Conn, error) {
	if _, err := socketName(sockType); err != nil {
		return nil, err
	}

	fd, err := C.argo_socket(C.int(sockType))
	if err = callErrno(fd, err); err != nil {
		return nil, err
	}

	peer := C.xen_argo_addr_t{
		aport:     C.xen_argo_port_t(port),
		domain_id: C.domid_t(domid),
	}
	rv, cerr := C.argo_connect(fd, &peer)
	if err = callErrno(rv, cerr); err != nil {
		closeFD(fd)
		return nil, err
	}

	return newConn(fd, sockType, Addr{
		Port:   uint32(port),
		Domain: DomainId(domid),
	})
}

// Listen opens an Argo socket, binds it, and listens for connections.
func Listen(sockType, port int, partner DomainId) (*Listener, error) {
	if _, err := socketName(sockType); err != nil {
		return nil, err
	}

	fd, err := C.argo_socket(C.int(sockType))
	if err = callErrno(fd, err); err != nil {
		return nil, err
	}

	addr := C.xen_argo_addr_t{
		aport:     C.xen_argo_port_t(port),
		domain_id: C.domid_t(XEN_ARGO_DOMID_ANY),
	}
	rv, cerr := C.argo_bind(fd, &addr, C.domid_t(partner))
	if err = callErrno(rv, cerr); err != nil {
		closeFD(fd)
		return nil, err
	}
	rv, cerr = C.argo_listen(fd, 5)
	if err = callErrno(rv, cerr); err != nil {
		closeFD(fd)
		return nil, err
	}

	c, err := newConn(fd, sockType, Addr{
		Port:   uint32(port),
		Domain: XEN_ARGO_DOMID_ANY,
	})
	if err != nil {
		return nil, err
	}
	return &Listener{
		conn: c,
		ring: RingId{
			Domain:  XEN_ARGO_DOMID_ANY,
			Partner: partner,
			Port:    uint32(port),
		},
	}, nil
}

// Accept waits for an incoming connection.
func (l *Listener) Accept() (*Conn, error) {
	var peer C.xen_argo_addr_t
	nfd, err := C.argo_accept(C.int(l.conn.Fd()), &peer)
	if err = callErrno(nfd, err); err != nil {
		return nil, err
	}

	sockType := syscall.SOCK_STREAM
	if l.conn.file.Name() == "/dev/argo_dgram" {
		sockType = syscall.SOCK_DGRAM
	}
	return newConn(nfd, sockType, Addr{
		Port:   uint32(peer.aport),
		Domain: DomainId(peer.domain_id),
	})
}
