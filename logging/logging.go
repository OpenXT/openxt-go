// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Soutions, LLC
//

package logging

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
)

var (
	DefaultLogLevel syslog.Priority = syslog.LOG_INFO
)

type SystemLogger struct {
	handle *syslog.Writer
}

func NewSystemLogger(name string) *SystemLogger {
	l := SystemLogger{}

	w, e := syslog.New(syslog.LOG_DAEMON|DefaultLogLevel, name)
	if e != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to connect to syslog!\n")
		return &l
	}

	// Any library that uses Go's log library will also get logged to
	// syslog at the debug level
	log.SetOutput(w)

	l.handle = w

	return &l
}

func (l *SystemLogger) Close() {
	if l.handle != nil {
		l.handle.Close()
	}
}

func (l *SystemLogger) Emerg(format string, a ...interface{}) {
	if l.handle != nil {
		m := fmt.Sprintf(format, a...)
		l.handle.Emerg(m)
	}
}

func (l *SystemLogger) Alert(format string, a ...interface{}) {
	if l.handle != nil {
		m := fmt.Sprintf(format, a...)
		l.handle.Alert(m)
	}
}

func (l *SystemLogger) Crit(format string, a ...interface{}) {
	if l.handle != nil {
		m := fmt.Sprintf(format, a...)
		l.handle.Crit(m)
	}
}

func (l *SystemLogger) Err(format string, a ...interface{}) {
	if l.handle != nil {
		m := fmt.Sprintf(format, a...)
		l.handle.Err(m)
	}
}

func (l *SystemLogger) Warning(format string, a ...interface{}) {
	if l.handle != nil {
		m := fmt.Sprintf(format, a...)
		l.handle.Warning(m)
	}
}

func (l *SystemLogger) Notice(format string, a ...interface{}) {
	if l.handle != nil {
		m := fmt.Sprintf(format, a...)
		l.handle.Notice(m)
	}
}

func (l *SystemLogger) Info(format string, a ...interface{}) {
	if l.handle != nil {
		m := fmt.Sprintf(format, a...)
		l.handle.Info(m)
	}
}

func (l *SystemLogger) Debug(format string, a ...interface{}) {
	if l.handle != nil {
		m := fmt.Sprintf(format, a...)
		l.handle.Debug(m)
	}
}
