// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"time"
)

import (
	"faucet"
)

type mockFaucet struct {
	amt  float64
	avs  []uint
	err  error
	wait time.Time
}

func (self *mockFaucet) AddressVersions() []uint { return self.avs }

func (self *mockFaucet) Amount(ctx context.Context) (float64, error) {
	return self.amt, self.err
}

func (self *mockFaucet) Claim(ctx context.Context, client, recipient, token string) (amount float64, tx string, err error) {
	if self.err != nil {
		err = self.err
		return
	}
	if time.Now().Before(self.wait) {
		err = faucet.MustWait{Until: self.wait}
		return
	}
	if self.amt <= 0 {
		err = faucet.ErrNoFunds
		return
	}
	btx := make([]byte, 32)
	_, err = rand.Read(btx)
	if err != nil {
		return
	}
	amount = self.amt
	tx = hex.EncodeToString(btx)
	return
}

func (self *mockFaucet) Token(ctx context.Context, client string) (string, error) {
	if self.err != nil {
		return "", self.err
	}
	t := make([]byte, 12)
	_, err := rand.Read(t)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(t), nil
}

func (self *mockFaucet) WaitTime(ctx context.Context, client string) (time.Time, error) {
	if self.err != nil {
		return time.Time{}, self.err
	}
	if time.Now().Before(self.wait) {
		return self.wait, nil
	}
	return time.Time{}, nil
}
