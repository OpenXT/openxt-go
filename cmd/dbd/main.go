// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Soutions, LLC
//

package main

import (
	"fmt"
	"log/syslog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openxt/openxt-go/db"
	"github.com/openxt/openxt-go/logging"
	flag "github.com/spf13/pflag"
)

const (
	RefreshUnit     = time.Minute
	RefreshInterval = 30
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(0)
}

func main() {
	helpFlag := flag.Bool("help", false, "Print help")
	debugFlag := flag.Bool("debug", false, "Enable debug")
	configDirFlag := flag.String("config", "", "Set the config directory")
	flag.Parse()

	if *helpFlag {
		usage()
	}

	if *debugFlag {
		logging.DefaultLogLevel = syslog.LOG_DEBUG
	}

	if *configDirFlag != "" {
		db.ConfigPath = *configDirFlag
	}

	sigs := make(chan os.Signal, 1)
	exit := make(chan int, 1)
	signal.Notify(sigs,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)

	s, err := db.NewServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = s.DBusListen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	go func() {
		for {
			select {
			case sig := <-sigs:
				switch sig {
				case syscall.SIGHUP:
					if err := s.Reload(); err != nil {
						exit <- 1
						return
					}
				default:
					s.Shutdown(true)
					exit <- 0
					return
				}
			case <-time.After(RefreshInterval * RefreshUnit):
				if err := s.Sync(); err != nil {
					exit <- 1
					return
				}
			}
		}
	}()

	code := <-exit
	os.Exit(code)
}
