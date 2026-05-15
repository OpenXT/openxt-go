// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Soutions, LLC
//

package db

import (
	"github.com/godbus/dbus/v5/introspect"
)

const DbdIntrospection = `
<node name="/">
  <!--~~~~~~~~~~~~~~~~~~~~~~
  ~~~~  DBus Interface  ~~~~
  ~~~~~~~~~~~~~~~~~~~~~~~-->
  <interface name="com.citrix.xenclient.db">
    <!--~~~~~~~~~~~~~~~~~~~~~~
    ~~~~  DBus Methods    ~~~~
    ~~~~~~~~~~~~~~~~~~~~~~~-->
    <method name="dump">
      <arg name="path" type="s" direction="in"/>
      <arg name="value" type="s" direction="out"/>
    </method>
    <method name="exists">
      <arg name="path" type="s" direction="in"/>
      <arg name="ex" type="b" direction="out"/>
    </method>
    <method name="inject">
      <arg name="path" type="s" direction="in"/>
      <arg name="value" type="s" direction="in"/>
    </method>
    <method name="list">
      <arg name="path" type="s" direction="in"/>
      <arg name="value" type="as" direction="out"/>
    </method>
    <method name="read">
      <arg name="path" type="s" direction="in"/>
      <arg name="value" type="s" direction="out"/>
    </method>
    <method name="read_binary">
      <arg name="path" type="s" direction="in"/>
      <arg name="value" type="ay" direction="out"/>
    </method>
    <method name="rm">
      <arg name="path" type="s" direction="in"/>
    </method>
    <method name="write">
      <arg name="path" type="s" direction="in"/>
      <arg name="value" type="s" direction="in"/>
    </method>
  </interface>
` + introspect.IntrospectDataString + `
</node>`
