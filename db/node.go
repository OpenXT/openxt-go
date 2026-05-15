// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Soutions, LLC
//

package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/openxt/openxt-go/utils"
)

type NodeType int

const (
	UnknownNode NodeType = iota
	MapNode
	ListNode
	StringNode
	NumberNode
	BooleanNode
)

func (nt *NodeType) String() string {
	switch *nt {
	case UnknownNode:
		return "Unknown"
	case MapNode:
		return "Map"
	case ListNode:
		return "List"
	case StringNode:
		return "String"
	case NumberNode:
		return "Number"
	case BooleanNode:
		return "Boolean"
	}
	return "Unknown"
}

type NodePath []string

type Node struct {
	data interface{}
}

func NewNode(jsonBytes []byte) (*Node, error) {
	var node Node

	if err := json.Unmarshal(jsonBytes, &node.data); err != nil {
		return nil, utils.FormatJsonError(jsonBytes, err)
	}

	return &node, nil
}

func (n *Node) IsMap() bool {
	if _, ok := (n.data).(map[string]interface{}); ok {
		return true
	}
	return false
}

func (n *Node) IsList() bool {
	if _, ok := (n.data).([]interface{}); ok {
		return true
	}
	return false
}

func (n *Node) IsString() bool {
	if _, ok := (n.data).(string); ok {
		return true
	}
	return false
}

func (n *Node) IsNumber() bool {
	switch n.data.(type) {
	case float32, float64:
		return true
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

func (n *Node) IsBoolean() bool {
	if _, ok := (n.data).(bool); ok {
		return true
	}
	return false
}

func (n *Node) String() string {

	switch n.data.(type) {
	case string:
		return n.data.(string)
	case bool:
	case float32, float64:
		return strconv.FormatFloat(n.data.(float64), 'e', -1, 64)
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(n.data.(int64), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(n.data.(uint64), 10)
	}

	return ""
}

func (n *Node) Type() NodeType {
	switch {
	case n.IsMap():
		return MapNode
	case n.IsList():
		return ListNode
	case n.IsString():
		return StringNode
	case n.IsNumber():
		return NumberNode
	case n.IsBoolean():
		return BooleanNode
	}
	return UnknownNode
}

var (
	ErrNoSuchChild = errors.New("no such child")
	ErrNotAMap     = errors.New("node is not a map")
)

func (n *Node) Children() []string {
	if n.IsMap() {
		m := (n.data).(map[string]interface{})

		children := []string{}
		for k := range m {
			children = append(children, k)
		}
		return children
	}

	return []string{}
}

func (n *Node) nextChild(key string) (*Node, error) {
	if n.IsMap() {
		m := (n.data).(map[string]interface{})
		if _, ok := m[key]; !ok {
			return nil, fmt.Errorf("%w: invalid map key (%s)",
				ErrNoSuchChild, key)
		}

		return &Node{data: m[key]}, nil
	}

	return nil, ErrNotAMap
}

func (n *Node) Child(path NodePath) (*Node, error) {
	if len(path) == 0 {
		return n, nil
	}

	next, err := n.nextChild(path[0])
	if err != nil {
		return nil, err
	}

	return next.Child(path[1:])
}

func (n *Node) NewChild(key string) (*Node, error) {
	if n.IsMap() {
		node := Node{data: map[string]interface{}{}}
		m := (n.data).(map[string]interface{})
		m[key] = node.data

		return &node, nil
	}

	return nil, ErrNotAMap
}

func (n *Node) DelChild(key string) error {
	if n.IsMap() {
		m := (n.data).(map[string]interface{})
		delete(m, key)

		return nil
	}

	return ErrNotAMap
}

func (n *Node) AddNode(path NodePath, node *Node, replace bool) error {
	if len(path) == 0 {
		return fmt.Errorf("path must have at least one element")
	}

	if !n.IsMap() {
		return ErrNotAMap
	}

	next, err := n.nextChild(path[0])
	if !errors.Is(err, ErrNoSuchChild) {
		return err
	}

	if len(path) == 1 {
		if next != nil {
			if replace {
				next.data = node.data
				return nil
			}
			return fmt.Errorf("node exists")
		}

		m := (n.data).(map[string]interface{})
		m[path[0]] = node.data

		return nil
	}

	if next != nil {
		return next.AddNode(path[1:], node, replace)
	}

	child, err := n.NewChild(path[0])
	if err != nil {
		return err
	}
	if err := child.AddNode(path[1:], node, replace); err != nil {
		n.DelChild(path[0])
		return err
	}

	return nil
}

func (n *Node) DelNode(path NodePath) error {
	plen := len(path)
	key := path[plen-1]
	parent, err := n.Child(path[:plen-1])
	if err != nil {
		return err
	}

	m := (parent.data).(map[string]interface{})

	if _, ok := m[key]; ok {
		delete(m, key)
		return nil
	}

	return ErrNoSuchChild
}

func (n *Node) Exists(path NodePath) bool {
	_, err := n.Child(path)
	if err == nil {
		return true
	}

	return false
}

func (n *Node) List(indent string) [][]string {
	paths := [][]string{}

	if n.IsMap() {
		m := (n.data).(map[string]interface{})

		for k, v := range m {
			child := Node{data: v}
			paths = append(paths, []string{indent, k, child.String()})

			if child.IsMap() {
				childPaths := child.List(indent + " ")
				paths = append(paths, childPaths...)
			}
		}
	}

	return paths
}
