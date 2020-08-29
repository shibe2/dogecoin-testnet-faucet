// SPDX-License-Identifier: AGPL-3.0-or-later

package core_test

import (
	"testing"
	"time"
)

import (
	"faucet/core"
)

func TestTokens(t *testing.T) {
	key, err := core.GenTokenKey()
	if err != nil {
		t.Fatal("GenTokenKey failed:", err)
	}
	c, err := core.NewTokenCipher(key)
	if err != nil {
		t.Fatal("NewTokenCipher failed:", err)
	}
	ip1, err := core.ParseClientAddr("1.2.3.4")
	if err != nil {
		t.Fatal("ParseClientAddr failed:", err)
	}
	ip2, err := core.ParseClientAddr("2.3.4.5")
	if err != nil {
		t.Fatal("ParseClientAddr failed:", err)
	}
	tm := new(timeMock)
	tm.set(time.Now())
	core.Now = tm.get
	defer resetNow()
	t1 := core.GenToken(ip1, c)
	if !core.CheckToken(ip1, t1, c) {
		t.Error("token not accepted at the same instant")
	}
	t2 := core.GenToken(ip2, c)
	if core.CheckToken(ip1, t2, c) {
		t.Error("token for different IP address accepted")
	}
	for i := 0; i < core.TokenIntervalSec; i++ {
		tm.add(time.Second)
		t2 = core.GenToken(ip1, c)
		if t1 != t2 {
			break
		}
	}
	if t1 == t2 {
		t.Error("token did not change after interval")
	}
	if !core.CheckToken(ip1, t1, c) {
		t.Error("previous token not accepted")
	}
	tm.add(time.Hour)
	if core.CheckToken(ip1, t1, c) {
		t.Error("old token accepted")
	}
}
