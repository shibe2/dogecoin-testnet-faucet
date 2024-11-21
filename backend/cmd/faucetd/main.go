// SPDX-License-Identifier: AGPL-3.0-or-later

// Package faucetd is faucet back-end service.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

import (
	"faucet"
	"faucet/core"
	"faucet/exalert"
	"faucet/platform"
	"faucet/rpc"
	"faucet/server"
	"faucet/sqldb"
)

type logCfg struct{ Date, Time, Microseconds, UTC bool }

type config struct {
	Faucet core.FaucetConfig       `yaml:",inline"`
	Alerts exalert.ExAlerterConfig `yaml:",inline"`
	Server server.ServerConfig     `yaml:",inline"`
	DB     sqldb.DBConfig
	RPC    rpc.RPCConfig
	Log    logCfg
}

var defCfg = config{
	Faucet: core.FaucetConfig{
		Fee:       1,
		MinAmount: 2,
	},
	Server: server.ServerConfig{
		APIPrefix: "/api",
	},
	RPC: rpc.RPCConfig{
		URL: "http://localhost:44555",
	},
	Log: logCfg{
		Date: true,
		Time: true,
	},
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
	return "faucetd"
}

func usage() {
	pn := progName()
	fmt.Println("usage:")
	fmt.Println(pn, "config create configout.yaml")
	fmt.Println(pn, "config dump config.yaml")
	fmt.Println(pn, "config process config.yaml configout.yaml")
	fmt.Println(pn, "db create config.yaml")
	fmt.Println(pn, "db sql driver_name")
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
		var err error
		cfg.Faucet.TokenKey, err = core.GenTokenKey()
		if err != nil {
			return err
		}
		cfg.Faucet.AddressVersions = []uint{113, 196}
		cfg.RPC.CookieFile = platform.DefaultCookieFile()
		err = storeYAML(args[1], &cfg)
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

var sqlStmtEnd = []byte{';', '\n'}

func cmdDB(args []string) error {
	if len(args) < 1 {
		usage()
	}
	switch args[0] {
	case "create":
		if len(args) != 2 {
			usage()
		}
		cfg := defCfg
		err := loadYAML(args[1], &cfg)
		if err != nil {
			return err
		}
		if !cfg.DB.Configured() {
			return fmt.Errorf("database is not configured")
		}
		db, err := sqldb.NewDB(&cfg.DB)
		if err != nil {
			return err
		}
		err = db.CreateTables()
		if err != nil {
			return err
		}
		err = db.Close()
		if err != nil {
			return err
		}
	case "sql":
		if len(args) != 2 {
			usage()
		}
		sql := sqldb.CreateSQL[args[1]]
		if len(sql) == 0 {
			return fmt.Errorf("don't have table creation SQL code for driver %q", args[1])
		}
		for _, s := range sql {
			n, err := io.WriteString(os.Stdout, s)
			if err != nil {
				return err
			}
			if n < len(s) {
				return io.ErrShortWrite
			}
			n, err = os.Stdout.Write(sqlStmtEnd)
			if err != nil {
				return err
			}
			if n < len(sqlStmtEnd) {
				return io.ErrShortWrite
			}
		}
	default:
		usage()
	}
	return nil
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
	lf := 0
	if cfg.Log.Date {
		lf |= log.Ldate
	}
	if cfg.Log.Time {
		lf |= log.Ltime
	}
	if cfg.Log.Microseconds {
		lf |= log.Lmicroseconds
	}
	if cfg.Log.UTC {
		lf |= log.LUTC
	}
	log.SetFlags(lf)
	var al faucet.Alerter
	if cfg.Alerts.Configured() {
		al = exalert.NewExAlerter(&cfg.Alerts)
	}
	bank, err := rpc.NewRPCClient(&cfg.RPC)
	if err != nil {
		return err
	}
	var sdb *sqldb.DB
	var fdb faucet.FaucetDB
	if len(cfg.DB.Driver) > 0 {
		sdb, err = sqldb.NewDB(&cfg.DB)
		if err != nil {
			return err
		}
		defer func() {
			if sdb != nil {
				sdb.Close()
			}
		}()
		fdb = sdb
	}
	f, err := core.NewFaucet(&cfg.Faucet, al, bank, fdb)
	if err != nil {
		return err
	}
	s := server.NewServer(&cfg.Server, f)
	go shutdownOnSignal(s)
	err = s.Serve(nil)
	if err != nil {
		return err
	}
	if sdb != nil {
		err = sdb.Close()
		sdb = nil
	}
	return err
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var c func(args []string) error
	switch os.Args[1] {
	case "config":
		c = cmdConfig
	case "db":
		c = cmdDB
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
