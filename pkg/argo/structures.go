//
// Copyright 2020 Apertus Soutions, LLC
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file
//

package argo

import (
	"os"
)

// DomainId is a Xen domain id
type DomainId uint16

// Addr is an Argo address
type Addr struct {
	Port   uint32
	Domain DomainId
	pad    uint16
}

// RingId represent an Argo ring
type RingId struct {
	Domain  DomainId
	Partner DomainId
	Port    uint32
}

// VIpTablesRule is a rule in Argo firewall
type VIpTablesRule struct {
	Src    Addr
	Dst    Addr
	Accept uint32
}

// VIpTablesRulePosition provides the position of a rule in the Argo firewall
type VIpTablesRulePosition struct {
	Rule     VIpTablesRule
	Position uint32
}

// A Conn is an active connection over Argo.
type Conn struct {
	file *os.File
	addr Addr
}

// A Listener is an active listening connection over Argo
type Listener struct {
	conn *Conn
	ring RingId
}
