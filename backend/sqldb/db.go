// SPDX-License-Identifier: AGPL-3.0-or-later

// Package sqldb implements faucet database using SQL back-end.
package sqldb

import (
	"database/sql"
	"net"
	"time"
)

import (
	"faucet"
)

// Driver-specific SQL code to create needed tables.
var CreateSQL = make(map[string][]string)

type ErrUnsupportedDriver struct{ Driver string }

func (self ErrUnsupportedDriver) Error() string { return "unsupported SQL driver " + self.Driver }

type cli struct{ rs *sql.Rows }

func (self cli) Close() error {
	err := self.rs.Close()
	if err != nil {
		return err
	}
	return self.rs.Err()
}

func (self cli) Get(t *time.Time, client *net.IP, amount *float64) error {
	return self.rs.Scan(t, client, amount)
}

func (self cli) Next() bool { return self.rs.Next() }

type DBConfig struct{ Driver, Source string }

func (self *DBConfig) Configured() bool { return len(self.Driver) > 0 }

type DB struct {
	db *sql.DB
	dn string
}

func (self *DB) ClaimsSince(t time.Time) (faucet.ClaimLogIter, error) {
	rs, err := self.db.Query(`SELECT"time","client","amount"FROM"claims"WHERE"time">=?`, t.UTC())
	if err != nil {
		return nil, err
	}
	return cli{rs}, nil
}

func (self *DB) Close() error { return self.db.Close() }

func (self *DB) CreateTables() error {
	csql := CreateSQL[self.dn]
	if len(csql) == 0 {
		return ErrUnsupportedDriver{Driver: self.dn}
	}
	for _, s := range csql {
		_, err := self.db.Exec(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *DB) LogClaim(t time.Time, client net.IP, recipient string, amount float64, tx []byte) error {
	_, err := self.db.Exec(`INSERT INTO"claims"("time","client","recipient","amount","txid")VALUES(?,?,?,?,?)`, t.UTC(), client, recipient, amount, tx)
	return err
}

func NewDB(cfg *DBConfig) (*DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.Source)
	if err != nil {
		return nil, err
	}
	return &DB{
		db: db,
		dn: cfg.Driver,
	}, nil
}
