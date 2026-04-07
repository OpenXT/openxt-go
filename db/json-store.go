// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Soutions, LLC
//

package db

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

func StringToNodePath(s string) NodePath {
	s = strings.Trim(s, PathDelimiter)
	if len(s) == 0 {
		return NodePath{}
	}
	elems := strings.Split(s, PathDelimiter)

	return NodePath(elems)
}

func NodePathToString(np NodePath) string {
	path := PathDelimiter

	for _, p := range np {
		path += PathDelimiter + p
	}

	return path
}

type JsonStore struct {
	mutex sync.RWMutex
	root  *Node
}

func NewJsonStore(base []byte) (*JsonStore, error) {
	node, err := NewNode(base)
	if err != nil {
		return nil, err
	}

	return &JsonStore{root: node}, nil
}

func (js *JsonStore) Read(path string) (string, error) {
	np := StringToNodePath(path)

	js.mutex.RLock()
	defer js.mutex.RUnlock()

	if !js.root.Exists(np) {
		return "", nil
	}

	node, err := js.root.Child(np)
	if err != nil {
		return "", err
	}

	return node.String(), nil
}

func (js *JsonStore) Write(path, value string) error {
	np := StringToNodePath(path)
	pp := np[:len(np)-1]
	key := np[len(np)-1]

	js.mutex.Lock()
	defer js.mutex.Unlock()

	if !js.root.Exists(pp) {
		node := Node{data: map[string]interface{}{key: value}}
		return js.root.AddNode(pp, &node, false)
	}

	node, err := js.root.Child(pp)
	if err != nil {
		return err
	}

	if !node.IsMap() {
		return fmt.Errorf("%s is not a path element", NodePathToString(pp))
	}

	m := (node.data).(map[string]interface{})
	m[key] = value

	return nil
}

func (js *JsonStore) Dump(path string) ([]byte, error) {
	np := StringToNodePath(path)

	js.mutex.RLock()
	defer js.mutex.RUnlock()

	node, err := js.root.Child(np)
	if err != nil {
		return nil, err
	}

	return json.Marshal(node.data)
}

func (js *JsonStore) Inject(path string, contents []byte) error {
	np := StringToNodePath(path)

	js.mutex.Lock()
	defer js.mutex.Unlock()

	node, err := NewNode(contents)
	if err != nil {
		return err
	}

	if err := js.root.AddNode(np, node, false); err != nil {
		return err
	}

	return nil
}

func (js *JsonStore) RList(path string) ([]string, error) {
	np := StringToNodePath(path)

	js.mutex.RLock()
	defer js.mutex.RUnlock()

	node, err := js.root.Child(np)
	if err != nil {
		return nil, err
	}

	paths := [][]string{
		[]string{"", np[len(np)-1], node.String()},
	}

	if node.IsMap() {
		childPaths := node.List(" ")
		paths = append(paths, childPaths...)
	}

	var entries []string
	for _, e := range paths {
		if len(e) != 3 {
			return nil, fmt.Errorf("invalid path entry")
		}
		entry := e[0] + e[1] + " = " + e[2]
		entries = append(entries, entry)
	}

	return entries, nil
}

func (js *JsonStore) List(path string) ([]string, error) {
	np := StringToNodePath(path)

	js.mutex.RLock()
	defer js.mutex.RUnlock()

	node, err := js.root.Child(np)
	if err != nil {
		return nil, err
	}

	return node.Children(), nil
}

func (js *JsonStore) Remove(path string) error {
	np := StringToNodePath(path)

	js.mutex.Lock()
	defer js.mutex.Unlock()

	return js.root.DelNode(np)
}

func (js *JsonStore) Exist(path string) bool {
	np := StringToNodePath(path)

	js.mutex.RLock()
	defer js.mutex.RUnlock()

	return js.root.Exists(np)
}
