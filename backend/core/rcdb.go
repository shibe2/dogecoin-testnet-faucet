// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"container/heap"
	"net"
	"sort"
	"sync"
	"time"
)

import (
	"faucet"
)

// cRecord contains information about a claim.
type cRecord struct {
	a float64
	t time.Time
}

// cRecords is an array of cRecord that can be sorted by time.
type cRecords []cRecord

func (self cRecords) Len() int           { return len(self) }
func (self cRecords) Less(i, j int) bool { return self[i].t.Before(self[j].t) }
func (self cRecords) Swap(i, j int)      { self[i], self[j] = self[j], self[i] }

// iRecord containst information about interval for a particular address prefix.
type iRecord struct {
	a string
	t time.Time
}

// iRecords is a heap of iRecord ordered by time.
type iRecords []iRecord

func (self iRecords) Len() int            { return len(self) }
func (self iRecords) Less(i, j int) bool  { return self[i].t.Before(self[j].t) }
func (self iRecords) Swap(i, j int)       { self[i], self[j] = self[j], self[i] }
func (self *iRecords) Push(x interface{}) { *self = append(*self, x.(iRecord)) }
func (self *iRecords) Pop() interface{} {
	nl := len(*self) - 1
	x := (*self)[nl]
	*self = (*self)[:nl]
	return x
}

// RCDB keeps track of recent claims for purposes of rate limiting.
// Old records are automatically removed.
type RCDB struct {
	IPClaimInterval time.Duration // Minimum interval between claims from the same IP address or prefix.
	RatePeriod      time.Duration // Period over which total amount is computed.
	cs              cRecords
	csp             int
	ish             iRecords
	ism             map[string]time.Time
	m               sync.Mutex
	ta              float64
}

func (self *RCDB) compactClaims() {
	copy(self.cs, self.cs[self.csp:])
	self.cs = self.cs[:len(self.cs)-self.csp]
	self.csp = 0
}

func (self *RCDB) purgeClaims(before time.Time) {
	for self.csp < len(self.cs) && before.After(self.cs[self.csp].t) {
		self.ta -= self.cs[self.csp].a
		self.csp++
	}
	if self.csp == len(self.cs) {
		self.cs = self.cs[:0]
		self.csp = 0
		self.ta = 0
	}
}

func (self *RCDB) purgeIntervals(ct time.Time) {
	for len(self.ish) > 0 && ct.After(self.ish[0].t) {
		delete(self.ism, heap.Pop(&self.ish).(iRecord).a)
	}
}

// AddClaim adds a claim record.
func (self *RCDB) AddClaim(t time.Time, amount float64) {
	self.m.Lock()
	defer self.m.Unlock()
	self.purgeClaims(Now().Add(-self.RatePeriod))
	if self.csp > cap(self.cs)/2 {
		self.compactClaims()
	}
	self.cs = append(self.cs, cRecord{
		a: amount,
		t: t,
	})
	self.ta += amount
}

// AddFromLog adds records from claim log.
func (self *RCDB) AddFromLog(cli faucet.ClaimLogIter) error {
	self.m.Lock()
	defer self.m.Unlock()
	ct := Now()
	rt := ct.Add(-self.RatePeriod)
	self.purgeClaims(rt)
	self.compactClaims()
	if self.ism == nil {
		self.ism = make(map[string]time.Time)
	}
	self.purgeIntervals(ct)
	var (
		lt     time.Time
		client net.IP
		amt    float64
	)
	for cli.Next() {
		err := cli.Get(&lt, &client, &amt)
		if err != nil {
			return err
		}
		if rt.Before(lt) {
			self.cs = append(self.cs, cRecord{
				a: amt,
				t: lt,
			})
			self.ta += amt
		}
		a1 := ClientRLAddr(client)
		a2 := string(a1[:])
		d := self.IPClaimInterval
		for l := len(a2); l > 0 && d > time.Second; l-- {
			it := lt.Add(d)
			if ct.After(it) {
				break
			}
			if it.Before(self.ism[a2[:l]]) {
				continue
			}
			self.ism[a2[:l]] = it
			d /= 16
		}
	}
	sort.Sort(self.cs)
	self.ish = self.ish[:0]
	for a, t := range self.ism {
		self.ish = append(self.ish, iRecord{
			a: a,
			t: t,
		})
	}
	heap.Init(&self.ish)
	return nil
}

// CheckAddIntervals atomically checks if the claim should be allowed now and if yes, adds corresponding interval records.
// Returns time points of added records.
// Returns nil if claiming should not be allowed.
func (self *RCDB) CheckAddIntervals(a [8]byte) []time.Time {
	self.m.Lock()
	defer self.m.Unlock()
	ct := Now()
	self.purgeIntervals(ct)
	as := string(a[:])
	for l := len(as); l > 0; l-- {
		if ct.Before(self.ism[as[:l]]) {
			return nil
		}
	}
	if self.ism == nil {
		self.ism = make(map[string]time.Time)
	}
	var ts []time.Time
	d := self.IPClaimInterval
	for l := len(as); l > 0 && d > time.Second; l-- {
		it := ct.Add(d)
		heap.Push(&self.ish, iRecord{
			a: as[:l],
			t: it,
		})
		self.ism[as[:l]] = it
		ts = append(ts, it)
		d /= 16
	}
	return ts
}

// CheckInterval checks if claiming from this client address should be allowed now.
// Returns zero if claiming should be allowed, otherwise returns time of next claim.
func (self *RCDB) CheckInterval(a [8]byte) time.Time {
	self.m.Lock()
	defer self.m.Unlock()
	ct := Now()
	self.purgeIntervals(ct)
	as := string(a[:])
	var nt time.Time
	for l := len(as); l > 0; l-- {
		it := self.ism[as[:l]]
		if nt.Before(it) {
			nt = it
		}
	}
	return nt
}

// DelIntervals removes records added by CheckAddIntervals.
func (self *RCDB) DelIntervals(a [8]byte, ts []time.Time) {
	self.m.Lock()
	defer self.m.Unlock()
	a1 := string(a[:])
	for i, t1 := range ts {
		a2 := a1[:len(a)-i]
		t2 := self.ism[a2]
		if t2.IsZero() {
			continue
		}
		if t1.Before(t2) {
			continue
		}
		delete(self.ism, a2)
		for j, r := range self.ish {
			if r.a == a2 {
				heap.Remove(&self.ish, j)
				break
			}
		}
	}
}

// PeriodTotal returns total amount of claims during the set period.
func (self *RCDB) PeriodTotal() float64 {
	self.m.Lock()
	defer self.m.Unlock()
	self.purgeClaims(Now().Add(-self.RatePeriod))
	return self.ta
}
