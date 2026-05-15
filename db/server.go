// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Soutions, LLC
//

package db

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/openxt/openxt-go/logging"
)

var (
	ConfigPath = "/config"

	methodMap = map[string]string{
		"Dump":        "dump",
		"Exists":      "exists",
		"Inject":      "inject",
		"List":        "list",
		"Read":        "read",
		"ReadBinary":  "read_binary",
		"Remove":      "rm",
		"Write":       "write",
		"WriteBinary": "write_binary",
	}

	logger *logging.SystemLogger
)

const (
	coreDbFile  = "db"
	vmDir       = "vms"
	domstoreDir = "dom-store"

	dbdInterface        = "com.citrix.xenclient.db"
	introspectInterface = "org.freedesktop.DBus.Introspectable"
)

type Server struct {
	js    *JsonStore
	conn  *dbus.Conn
	mutex sync.Mutex
	dirty bool
}

func NewServer() (*Server, error) {
	if logger == nil {
		logger = logging.NewSystemLogger("dbd")
	}
	logger.Info("starting new dbd server using config directory %s", ConfigPath)

	s := &Server{}

	if err := s.initJsonStore(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) initJsonStore() error {
	jsonBytes, err := ioutil.ReadFile(filepath.Join(ConfigPath, coreDbFile))
	if err != nil {
		logger.Crit("Failed reading db file (%s): %s", coreDbFile, err)
		return err
	}

	js, err := NewJsonStore(jsonBytes)
	if err != nil {
		logger.Crit("Failed parsing db file (%s): %s", coreDbFile, err)
		return err
	}

	s.js = js

	if err := s.loadConfigDir(vmDir, "/vm/"); err != nil {
		logger.Crit("Failed parsing VM db file(s): %s", err)
		return err
	}
	if err := s.loadConfigDir(domstoreDir, "/dom-store/"); err != nil {
		logger.Crit("Failed parsing domstore db file(s): %s", err)
		return err
	}

	return nil
}

func (s *Server) flushJsonStore() error {
	if err := s.storeConfigDir(vmDir, "/vm"); err != nil {
		logger.Crit("Failed to persist VM configs to disk: %s", err)
		return err
	}

	if err := s.js.Remove("/vm"); err != nil {
		logger.Crit("Failed to clear VM configs from store: %s", err)
		return err
	}

	if err := s.storeConfigDir(domstoreDir, "/dom-store"); err != nil {
		logger.Crit("Failed to persist domstore configs to disk: %s", err)
		return err
	}

	if err := s.js.Remove("/dom-store"); err != nil {
		logger.Crit("Failed to clear dom-store configs from store: %s", err)
		return err
	}

	f, err := os.OpenFile(filepath.Join(ConfigPath, coreDbFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logger.Crit("Unable to open db file, unable to write to disk: %s", err)
		return err
	}
	defer f.Close()

	contents, err := s.js.Dump("/")
	if err != nil {
		logger.Crit("Unable to export db contents, unable to write to disk: %s", err)
		return err
	}

	var buf bytes.Buffer

	if err := json.Indent(&buf, contents, "", " "); err != nil {
		logger.Crit("Failed to format db contents, unable to write to disk: %s", err)
		return err
	}

	if _, err := f.Write(buf.Bytes()); err != nil {
		logger.Crit("Failed to write db contents to disk: %s", err)
		return err
	}

	return nil
}

func (s *Server) loadConfigDir(dir, path string) error {
	baseDir := filepath.Join(ConfigPath, dir)
	entries, err := ioutil.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		node := strings.TrimSuffix(entry.Name(), ".db")
		if node == entry.Name() {
			continue
		}

		contents, err := ioutil.ReadFile(baseDir + "/" + entry.Name())
		if err != nil {
			logger.Err("Failed reading VM db file (%s), skipping: %s", entry.Name(), err)
		}

		if err := s.js.Inject(path+node, contents); err != nil {
			logger.Err("Failed inserting VM (%s) config, skipping: %s", node, err)
		}
	}

	return nil
}

func checkDir(dirName string) bool {
	dir, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dirName, 0755); err != nil {
			return false
		}
		return true
	}

	return dir.Mode().IsDir()
}

func (s *Server) storeConfigDir(dir, path string) error {
	dir = filepath.Join(ConfigPath, dir)
	path = strings.TrimRight(path, PathDelimiter)

	if !s.js.Exist(path) {
		logger.Info("Store config: No entries for path %s", path)
		return nil
	}

	if !checkDir(dir) {
		return fmt.Errorf("Unable to create directory: %s", dir)
	}

	entries, err := s.js.List(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fpath := fmt.Sprintf("%s/%s.db", dir, entry)
		f, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			logger.Crit("Unable to open db file for %s, skipping writing to disks: %s", entry, err)
			continue
		}

		contents, err := s.js.Dump(path + PathDelimiter + entry)
		if err != nil {
			logger.Crit("Unable to export contents for %s, skipping writing to disk: %s", entry, err)
			continue
		}

		var buf bytes.Buffer

		if err := json.Indent(&buf, contents, "", " "); err != nil {
			logger.Crit("Failed to format contents for %s, skipping writing to disk: %s", entry, err)
			continue
		}

		f.Write(buf.Bytes())
		f.Close()
	}

	return nil
}

func (s *Server) DBusListen() error {
	conn, err := dbus.SystemBus()
	if err != nil {
		logger.Crit("Failed connecting to dbus system bus: %s", err)
		return err
	}

	err = conn.ExportWithMap(s, methodMap, "/", dbdInterface)
	if err != nil {
		conn.Close()
		logger.Crit("Failed exporting dbus interface: %s", err)
		return err
	}
	err = conn.Export(introspect.Introspectable(DbdIntrospection), "/",
		introspectInterface)
	if err != nil {
		conn.Close()
		logger.Crit("Failed exporting introspection interface: %s", err)
		return err
	}

	reply, err := conn.RequestName(dbdInterface, dbus.NameFlagDoNotQueue)
	if err != nil {
		conn.Close()
		logger.Crit("Failed requesting dbus name: %s", err)
		return err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		conn.Close()
		logger.Crit("Failed requesting dbus primary owner: %s", err)
		return fmt.Errorf("name already taken")
	}

	s.conn = conn
	return nil
}

func (s *Server) Sync() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.dirty {
		if err := s.flushJsonStore(); err != nil {
			return err
		}
		if err := s.initJsonStore(); err != nil {
			return err
		}

		s.dirty = false
	}

	return nil
}

func (s *Server) Reload() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.initJsonStore(); err != nil {
		return err
	}

	s.dirty = false
	return nil
}

func (s *Server) Shutdown(flush bool) {
	s.conn.Close()
	if flush {
		s.flushJsonStore()
	}
}

func (s *Server) Dump(path string) (string, *dbus.Error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	jsb, err := s.js.Dump(path)
	if err != nil {
		logger.Err("dump: failed dumping path %s: %s", path, err)
		return "", dbus.MakeFailedError(err)
	}

	var buf bytes.Buffer

	if err := json.Indent(&buf, jsb, "", " "); err != nil {
		logger.Err("dump: failed formating for path %s: %s", path, err)
		return "", dbus.MakeFailedError(err)
	}

	return buf.String(), nil
}

func (s *Server) Exists(path string) (bool, *dbus.Error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.js.Exist(path), nil
}

func (s *Server) Inject(path, value string) *dbus.Error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := s.js.Inject(path, []byte(value))
	if err != nil {
		logger.Err("inject: failed inserting at path %s: %s", path, err)
		return dbus.MakeFailedError(err)
	}
	s.dirty = true
	return nil
}

func (s *Server) List(path string) ([]string, *dbus.Error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	str, err := s.js.List(path)
	if err != nil {
		logger.Err("list: failed list of path %s: %s", path, err)
		return nil, dbus.MakeFailedError(err)
	}
	return str, nil
}

func (s *Server) Read(path string) (string, *dbus.Error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	str, err := s.js.Read(path)
	if err != nil {
		logger.Err("read: failed reading path %s: %s", path, err)
		return "", dbus.MakeFailedError(err)
	}
	return str, nil
}

func (s *Server) ReadBinary(path string) ([]byte, *dbus.Error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	str, err := s.js.Read(path)
	if err != nil {
		logger.Err("read binary: failed reading path %s: %s", path, err)
		return nil, dbus.MakeFailedError(err)
	}

	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		logger.Err("read binary: failed decoding %s: %s", path, err)
		return nil, dbus.MakeFailedError(err)
	}

	return data, nil
}

func (s *Server) Remove(path string) *dbus.Error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := s.js.Remove(path)
	if err != nil {
		logger.Err("remove: failed deleting %s: %s", path, err)
		return dbus.MakeFailedError(err)
	}
	s.dirty = true
	return nil
}

func (s *Server) Write(path, value string) *dbus.Error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := s.js.Write(path, value)
	if err != nil {
		logger.Err("write: failed writing to %s: %s", path, err)
		return dbus.MakeFailedError(err)
	}
	s.dirty = true
	return nil
}

func (s *Server) WriteBinary(path string, value []byte) *dbus.Error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data := base64.StdEncoding.EncodeToString(value)
	err := s.js.Write(path, data)
	if err != nil {
		logger.Err("write binary: failed writing to %s: %s", path, err)
		return dbus.MakeFailedError(err)
	}
	s.dirty = true
	return nil
}
