// SPDX-License-Identifier: AGPL-3.0-or-later

package core_test

import (
	"math"
	"testing"
	"time"
)

import (
	"faucet/core"
)

func TestIPIntervals(t *testing.T) {
	tm := new(timeMock)
	tm.set(time.Now().Truncate(time.Second))
	core.Now = tm.get
	defer resetNow()
	db := core.RCDB{IPClaimInterval: 256 * time.Minute}

	t1 := tm.get()
	gt := db.CheckInterval([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	if !gt.IsZero() {
		t.Error("check on empty DB:", gt, "want zero")
	}
	ts1 := db.CheckAddIntervals([8]byte{2, 3, 4, 5, 6, 7, 8, 9})
	if len(ts1) == 0 {
		t.Fatal("check-add on empty DB:", ts1, "want non-empty")
	}

	tm.add(time.Second)
	gt = db.CheckInterval([8]byte{3, 4, 5, 6, 7, 8, 9, 10})
	if !gt.IsZero() {
		t.Error("check on unrelated address:", gt, "want zero")
	}
	nt := t1.Add(256 * time.Minute)
	gt = db.CheckInterval([8]byte{2, 3, 4, 5, 6, 7, 8, 9})
	if !nt.Equal(gt) {
		t.Error("check on same address:", gt, "want", nt)
	}
	ts1 = db.CheckAddIntervals([8]byte{2, 3, 4, 5, 6, 7, 8, 9})
	if len(ts1) > 0 {
		t.Fatal("check-add on same address:", ts1, "want empty")
	}
	nt = t1.Add(16 * time.Minute)
	gt = db.CheckInterval([8]byte{2, 3, 4, 5, 6, 7, 8, 7})
	if !nt.Equal(gt) {
		t.Error("check on /56 subnet:", gt, "want", nt)
	}
	ts1 = db.CheckAddIntervals([8]byte{2, 3, 4, 5, 6, 7, 8, 6})
	if len(ts1) > 0 {
		t.Fatal("check-add on /56 subnet:", ts1, "want empty")
	}
	nt = t1.Add(time.Minute)
	gt = db.CheckInterval([8]byte{2, 3, 4, 5, 6, 7, 6, 5})
	if !nt.Equal(gt) {
		t.Error("check on /48 subnet:", gt, "want", nt)
	}
	ts1 = db.CheckAddIntervals([8]byte{2, 3, 4, 5, 6, 7, 6, 4})
	if len(ts1) > 0 {
		t.Fatal("check-add on /48 subnet:", ts1, "want empty")
	}

	tm.add(time.Minute)
	nt = t1.Add(256 * time.Minute)
	gt = db.CheckInterval([8]byte{2, 3, 4, 5, 6, 7, 8, 9})
	if !nt.Equal(gt) {
		t.Error("check on same address after 1/256 interval:", gt, "want", nt)
	}
	ts1 = db.CheckAddIntervals([8]byte{2, 3, 4, 5, 6, 7, 8, 9})
	if len(ts1) > 0 {
		t.Fatal("check-add on same address after 1/256 interval:", ts1, "want empty")
	}
	nt = t1.Add(16 * time.Minute)
	gt = db.CheckInterval([8]byte{2, 3, 4, 5, 6, 7, 8, 7})
	if !nt.Equal(gt) {
		t.Error("check on /56 subnet after 1/256 interval:", gt, "want", nt)
	}
	ts1 = db.CheckAddIntervals([8]byte{2, 3, 4, 5, 6, 7, 8, 6})
	if len(ts1) > 0 {
		t.Fatal("check-add on /56 subnet after 1/256 interval:", ts1, "want empty")
	}
	gt = db.CheckInterval([8]byte{2, 3, 4, 5, 6, 7, 6, 5})
	if !gt.IsZero() {
		t.Error("check on /48 subnet after 1/256 interval:", gt, "want zero")
	}

	tm.add(time.Second)
	t2 := tm.get()
	_ = t2
	ts1 = db.CheckAddIntervals([8]byte{2, 3, 4, 5, 6, 7, 6, 4})
	if len(ts1) == 0 {
		t.Fatal("check-add on /48 subnet after 1/256 interval:", ts1, "want non-empty")
	}

	tm.add(15 * time.Minute)
	nt = t1.Add(256 * time.Minute)
	gt = db.CheckInterval([8]byte{2, 3, 4, 5, 6, 7, 8, 9})
	if !nt.Equal(gt) {
		t.Error("check on same address after 1/16 interval:", gt, "want", nt)
	}
	gt = db.CheckInterval([8]byte{2, 3, 4, 5, 6, 7, 8, 7})
	if !gt.IsZero() {
		t.Error("check on /56 subnet after 1/16 interval:", gt, "want zero")
	}

	tm.add(time.Minute + time.Second)
	t3 := tm.get()
	_ = t3
	ts2 := db.CheckAddIntervals([8]byte{2, 3, 4, 5, 6, 7, 6, 3})
	if len(ts2) == 0 {
		t.Fatal("check-add on /56 subnet after 1/16 interval:", ts2, "want non-empty")
	}

	tm.add(time.Minute + time.Second)
	db.DelIntervals([8]byte{2, 3, 4, 5, 6, 7, 6, 4}, ts1)

	tm.add(14 * time.Minute)
	nt = t3.Add(16 * time.Minute)
	gt = db.CheckInterval([8]byte{2, 3, 4, 5, 6, 7, 6, 2})
	if !nt.Equal(gt) {
		t.Error("check on /56 subnet after 15/256 interval:", gt, "want", nt)
	}

	tm.add(time.Minute)
	gt = db.CheckInterval([8]byte{2, 3, 4, 5, 6, 7, 6, 2})
	if !gt.IsZero() {
		t.Error("check on /56 subnet after 1/16 interval:", gt, "want", nt)
	}
	gt = db.CheckInterval([8]byte{2, 3, 4, 5, 6, 7, 6, 4})
	if !gt.IsZero() {
		t.Error("check on same address after del:", gt, "want zero")
	}

	tm.add(223 * time.Minute)
	gt = db.CheckInterval([8]byte{2, 3, 4, 5, 6, 7, 8, 9})
	if !gt.IsZero() {
		t.Error("check on same address after full interval:", gt, "want zero")
	}

	tm.add(17 * time.Minute)
	gt = db.CheckInterval([8]byte{2, 3, 4, 5, 6, 7, 6, 3})
	if !gt.IsZero() {
		t.Error("check on same address after full interval:", gt, "want zero")
	}
}

func TestRCAmount(t *testing.T) {
	tm := new(timeMock)
	tm.set(time.Now())
	core.Now = tm.get
	defer resetNow()
	db := core.RCDB{RatePeriod: time.Hour}
	ga := db.PeriodTotal()
	if math.Abs(ga) > .25 {
		t.Fatal("initial total:", ga, "want", 0)
	}
	t0 := tm.get()
	var ta float64
	for i := 1; i < 10; i++ {
		db.AddClaim(t0.Add(time.Duration(i)*time.Minute), float64(i))
		ta += float64(i)
	}
	tm.add(time.Hour)
	for i := 0; i < 10; i++ {
		ta -= float64(i)
		ga = db.PeriodTotal()
		if math.Abs(ga-ta) > .25 {
			t.Fatal("full period: total", ga, "want", ta)
		}
		tm.add(time.Minute + time.Second)
	}
	t0 = tm.get()
	ta = 0
	for i := 0; i < 3; i++ {
		for j := i*8 + 1; j < i*8+9; j++ {
			db.AddClaim(t0.Add(time.Duration(j)*time.Minute), float64(j))
			ta += float64(j)
		}
		if i == 0 {
			tm.add(time.Hour)
		}
		for j := i * 5; j < i*5+5; j++ {
			ta -= float64(j)
			ga = db.PeriodTotal()
			if math.Abs(ga-ta) > .25 {
				t.Fatal("partial period: total", ga, "want", ta)
			}
			tm.add(time.Minute + time.Second)
		}
	}
	for i := 5 * 3; i < 8*3+1; i++ {
		ta -= float64(i)
		ga = db.PeriodTotal()
		if math.Abs(ga-ta) > .25 {
			t.Fatal("final purge: total", ga, "want", ta)
		}
		tm.add(time.Minute + time.Second)
	}
	if math.Abs(ta) > .25 {
		t.Fatal("test error: final total:", ta, "want 0")
	}
}
