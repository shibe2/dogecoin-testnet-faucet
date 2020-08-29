// SPDX-License-Identifier: AGPL-3.0-or-later

package core_test

import (
	"time"
)

import (
	"faucet/core"
)

type timeMock struct{ t time.Time }

func (self *timeMock) add(d time.Duration) { self.t = self.t.Add(d) }
func (self *timeMock) get() time.Time      { return self.t }
func (self *timeMock) set(t time.Time)     { self.t = t }

func resetNow() { core.Now = time.Now }
