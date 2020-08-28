// SPDX-License-Identifier: AGPL-3.0-or-later

// Package core implements core faucet logic.
package core

import (
	"context"
	"encoding/hex"
	"log"
	"net"
	"sync"
	"time"
)

import (
	"faucet"
	"faucet/base58"
)

type FaucetConfig struct {
	Amount, Fee, MinAmount, StingyAmount, LowBalance float64
	IPClaimInterval                                  time.Duration
	RateLimit                                        struct {
		Amount float64
		Period time.Duration
	}
	TokenKey        faucet.Bytes
	AddressVersions []uint
}

type Faucet struct {
	m             sync.Mutex
	balOK, rateOK bool
	alerter       faucet.Alerter
	bank          faucet.Bank
	cfg           FaucetConfig
	fdb           faucet.FaucetDB
	rcdb          RCDB
	tc            TokenCipher
}

func (self *Faucet) alert(bal, ramt float64) {
	self.m.Lock()
	defer self.m.Unlock()
	if bal <= self.cfg.LowBalance {
		if self.balOK {
			self.balOK = false
			go self.alerter.BalanceAlert(bal)
		}
	} else {
		self.balOK = true
	}
	if ramt > self.cfg.RateLimit.Amount {
		if self.rateOK {
			self.rateOK = false
			go self.alerter.RateAlert(ramt, self.cfg.RateLimit.Period)
		}
	} else {
		self.rateOK = true
	}
}

func (self *Faucet) amountAndBalance(ctx context.Context) (amount, balance float64, err error) {
	balance, err = self.bank.Balance(ctx)
	if err != nil {
		return
	}
	rl := self.cfg.RateLimit.Amount > 0 && self.cfg.RateLimit.Period >= time.Second
	var ramt float64
	if rl {
		ramt = self.rcdb.PeriodTotal()
	}
	amount = balance - self.cfg.Fee
	if amount > self.cfg.StingyAmount && self.cfg.StingyAmount >= self.cfg.MinAmount && amount < self.cfg.LowBalance {
		amount = self.cfg.StingyAmount
	}
	if amount > self.cfg.StingyAmount && ramt > self.cfg.RateLimit.Amount {
		amount = self.cfg.StingyAmount
	}
	if amount > self.cfg.Amount {
		amount = self.cfg.Amount
	}
	if amount < self.cfg.MinAmount {
		amount = 0
	}
	if self.alerter != nil && (self.cfg.LowBalance > 0 || rl) {
		self.alert(balance, ramt)
	}
	return
}

func (self *Faucet) validRecipient(recipient string) bool {
	if len(self.cfg.AddressVersions) == 0 {
		return true
	}
	rv := base58.AddressVersion(recipient)
	if rv < 0 {
		return false
	}
	for _, av := range self.cfg.AddressVersions {
		if uint(rv) == av {
			return true
		}
	}
	return false
}

func (self *Faucet) AddressVersions() []uint { return self.cfg.AddressVersions }

func (self *Faucet) Amount(ctx context.Context) (float64, error) {
	amt, _, err := self.amountAndBalance(ctx)
	if err != nil {
		err = faucet.ServiceUnavailableError{Err: err}
	}
	return amt, err
}

func (self *Faucet) Claim(ctx context.Context, client, recipient, token string) (amount float64, tx string, err error) {
	if !self.validRecipient(recipient) {
		err = faucet.ErrInvalidRecipient
		return
	}
	var a1 net.IP
	a1, err = ParseClientAddr(client)
	if err != nil {
		return
	}
	if self.tc != nil && (len(token) == 0 || !CheckToken(a1, token, self.tc)) {
		err = faucet.ErrInvalidToken
		return
	}
	var ts []time.Time
	if self.cfg.IPClaimInterval >= time.Second {
		a2 := ClientRLAddr(a1)
		ts = self.rcdb.CheckAddIntervals(a2)
		if len(ts) == 0 {
			t := self.rcdb.CheckInterval(a2)
			err = faucet.MustWait{Until: t}
			return
		}
		defer func() {
			if len(ts) > 0 {
				self.rcdb.DelIntervals(a2, ts)
			}
		}()
	}
	var bal float64
	amount, bal, err = self.amountAndBalance(ctx)
	if err != nil {
		err = faucet.ServiceUnavailableError{Err: err}
		return
	}
	if amount == 0 {
		if bal < self.cfg.Fee+self.cfg.MinAmount {
			err = faucet.ErrNoFunds
		} else {
			err = faucet.ErrPaused
		}
		return
	}
	t1 := Now()
	tx, err = self.bank.Send(nil, recipient, amount)
	t2 := Now()
	if err != nil && err != faucet.ErrInvalidRecipient {
		err = faucet.SendError{Err: err}
	}
	if len(tx) > 0 {
		ts = nil
		t := t1.Add(t2.Sub(t1) / 2)
		self.rcdb.AddClaim(t, amount)
		if self.fdb != nil {
			btx, err := hex.DecodeString(tx)
			if err != nil {
				log.Printf("failed to decode transactin identifier %q: %v", tx, err)
			}
			err = self.fdb.LogClaim(t, a1, recipient, amount, btx)
			if err != nil {
				log.Println("failed to log claim", t, a1, recipient, amount, tx, err)
			}
		}
	}
	return
}

func (self *Faucet) Token(ctx context.Context, client string) (string, error) {
	if self.tc == nil {
		return "", nil
	}
	a, err := ParseClientAddr(client)
	if err != nil {
		return "", err
	}
	return GenToken(a, self.tc), nil
}

func (self *Faucet) WaitTime(ctx context.Context, client string) (time.Time, error) {
	if self.cfg.IPClaimInterval < time.Second {
		return time.Time{}, nil
	}
	a, err := ParseClientAddr(client)
	if err != nil {
		return time.Time{}, err
	}
	return self.rcdb.CheckInterval(ClientRLAddr(a)), nil
}

// NewFaucet creates faucet core object. If alerter or db is nil, it will not be used.
func NewFaucet(cfg *FaucetConfig, alerter faucet.Alerter, bank faucet.Bank, db faucet.FaucetDB) (*Faucet, error) {
	self := &Faucet{
		alerter: alerter,
		bank:    bank,
		cfg:     *cfg,
		fdb:     db,
	}
	self.rcdb.IPClaimInterval = cfg.IPClaimInterval
	self.rcdb.RatePeriod = cfg.RateLimit.Period
	if len(cfg.TokenKey) > 0 {
		c, err := NewTokenCipher(cfg.TokenKey)
		if err != nil {
			return nil, err
		}
		self.tc = c
	}
	if db != nil {
		var rld time.Duration
		if rld < cfg.IPClaimInterval {
			rld = cfg.IPClaimInterval
		}
		if rld < cfg.RateLimit.Period {
			rld = cfg.RateLimit.Period
		}
		if rld >= time.Second {
			cli, err := db.ClaimsSince(Now().Add(-rld))
			if err != nil {
				return nil, err
			}
			defer func() {
				if cli != nil {
					cli.Close()
				}
			}()
			err = self.rcdb.AddFromLog(cli)
			if err != nil {
				return nil, err
			}
			err = cli.Close()
			cli = nil
			if err != nil {
				return nil, err
			}
		}
	}
	return self, nil
}
