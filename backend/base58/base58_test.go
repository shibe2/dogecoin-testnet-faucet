// SPDX-License-Identifier: AGPL-3.0-or-later

package base58_test

import (
	"testing"
)

import (
	"faucet/base58"
)

func TestVersion(t *testing.T) {
	for _, c := range [...]struct {
		s string
		v int
	}{
		{"", -1},
		{"ShortData8gn4BZ3oEasFafsEyBg97At", -1},
		{"ShortStringLF9AUv3BwWx4qx3", -1},
		{"1111111111111111111114oLvT2", 0},
		{"Vanity28Chh8vK8p8p2qYtK3KgCDLoVdaJ", 71},
		{"2n1XR4oJkmBdJMxhBGQGb96gQ88xUyGML1i", 255},
		{"LongData24ezZw7Dx4AF1n4fBM8it5RTN1i", -1},
		{"LongStringDmXZEK54JycjXfLdciqcJ1AKts", -1},
		{"CheckFaiL6vnwRczcqLGsb1gF6eMxQM7jm", -1},
		{"InvaLidCharacter5vYm16DKXtJEp2WazB", -1},
		{"invaLidCharacter0vYm16DKXtJEp2WazB", -1},
		{"invaLidCharacter{}Ym16DKXtJEp2WazB", -1},
	} {
		v := base58.AddressVersion(c.s)
		if v != c.v {
			t.Error(c.s, "got", v, "want", c.v)
		}
	}
}
