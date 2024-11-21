// SPDX-License-Identifier: AGPL-3.0-or-later

// Package mock-faucetd is mock API server for testing faucet front-end.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

import (
	"faucet/server"
)

type config struct {
	Server      server.ServerConfig `yaml:",inline"`
	ControlPage string
}

var defCfg = config{
	Server: server.ServerConfig{
		APIPrefix:   "/api",
		AllowOrigin: "*",
	},
	ControlPage: "/mock.html",
}

func progFile() string {
	fn, err := os.Executable()
	if len(fn) > 0 && err == nil {
		return fn
	}
	if len(os.Args) > 0 {
		return os.Args[0]
	}
	return ""
}

func progName() string {
	fn := progFile()
	if len(fn) > 0 {
		fn = filepath.Base(fn)
		if len(fn) > 0 && fn != "." && fn != string(filepath.Separator) {
			return fn
		}
	}
	return "mock-faucetd"
}

func usage() {
	pn := progName()
	fmt.Println("usage:")
	fmt.Println(pn, "config create configout.yaml")
	fmt.Println(pn, "config dump config.yaml")
	fmt.Println(pn, "config process config.yaml configout.yaml")
	fmt.Println(pn, "serve config.yaml")
	os.Exit(1)
}

func cmdConfig(args []string) error {
	if len(args) < 1 {
		usage()
	}
	switch args[0] {
	case "create":
		if len(args) != 2 {
			usage()
		}
		cfg := defCfg
		err := storeYAML(args[1], &cfg)
		if err != nil {
			return err
		}
	case "dump":
		if len(args) != 2 {
			usage()
		}
		cfg := defCfg
		err := loadYAML(args[1], &cfg)
		if err != nil {
			return err
		}
		e := yaml.NewEncoder(os.Stdout)
		err = e.Encode(&cfg)
		if err != nil {
			return err
		}
		err = e.Close()
		if err != nil {
			return err
		}
	case "process":
		if len(args) != 3 {
			usage()
		}
		cfg := defCfg
		err := loadYAML(args[1], &cfg)
		if err != nil {
			return err
		}
		err = storeYAML(args[2], &cfg)
		if err != nil {
			return err
		}
	default:
		usage()
	}
	return nil
}

func shutdownOnSignal(s *server.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for s := range c {
		if s == os.Interrupt {
			break
		}
	}
	signal.Stop(c)
	s.Stop()
}

func cmdServe(args []string) error {
	if len(args) != 1 {
		usage()
	}
	cfg := defCfg
	err := loadYAML(args[0], &cfg)
	if err != nil {
		return err
	}
	f := &mockFaucet{
		amt: 100,
		avs: []uint{113, 196},
	}
	s := server.NewServer(&cfg.Server, f)
	if len(cfg.ControlPage) > 0 {
		var h controlHandler
		h, err = newControlHandler(f)
		if err != nil {
			return err
		}
		s.Handle(cfg.ControlPage, h)
	}
	go shutdownOnSignal(s)
	return s.Serve()
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var c func(args []string) error
	switch os.Args[1] {
	case "config":
		c = cmdConfig
	case "serve":
		c = cmdServe
	default:
		usage()
	}
	err := c(os.Args[2:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadYAML(fn string, v interface{}) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()
	err = yaml.NewDecoder(f).Decode(v)
	if err != nil {
		return err
	}
	err = f.Close()
	f = nil
	return err
}

func storeYAML(fn string, v interface{}) error {
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()
	y := yaml.NewEncoder(f)
	err = y.Encode(v)
	if err != nil {
		return err
	}
	err = y.Close()
	if err != nil {
		return err
	}
	err = f.Close()
	f = nil
	return err
}
