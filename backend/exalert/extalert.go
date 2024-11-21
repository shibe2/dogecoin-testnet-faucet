// SPDX-License-Identifier: AGPL-3.0-or-later

// Package exalert sends alerts by executing an external program.
package exalert

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

type ExAlerterConfig struct{ AlertProgram string }

func (self *ExAlerterConfig) Configured() bool { return len(self.AlertProgram) > 0 }

// ExAlerter implements Alerter interface.
type ExAlerter struct {
	m  sync.Mutex
	p  string
	rt time.Time
}

// BalanceAlert executes the program with arguments "balance" and the balance.
// For example, given balance 1000:
//
//	program balance 1000
func (self *ExAlerter) BalanceAlert(balance float64) {
	self.m.Lock()
	defer self.m.Unlock()
	c := exec.Command(self.p, "balance", strconv.FormatFloat(balance, 'f', -1, 64))
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		log.Println("failed to send balance", balance, "alert:", err)
	}
}

// RateAlert executes the program with arguments "rate", the amount and period in seconds.
// For example, given amount 1000 and period 1 hour:
//
//	program rate 1000 3600
//
// After successful execution, further alerts of this type will be ignored until there will be a full period with no alerts.
func (self *ExAlerter) RateAlert(amount float64, period time.Duration) {
	self.m.Lock()
	defer self.m.Unlock()
	ct := time.Now()
	nt := ct.Add(period)
	if ct.Before(self.rt) {
		if nt.After(self.rt) {
			self.rt = nt
		}
		return
	}
	c := exec.Command(self.p, "rate", strconv.FormatFloat(amount, 'f', -1, 64), strconv.FormatInt(int64(period/time.Second), 10))
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		log.Println("failed to send rate", amount, period, "alert:", err)
	} else {
		self.rt = nt
	}
}

func NewExAlerter(cfg *ExAlerterConfig) *ExAlerter { return &ExAlerter{p: cfg.AlertProgram} }
