// SPDX-License-Identifier: AGPL-3.0-or-later

// Package base58 implements Base58 decoding.
package base58

import (
	"crypto/sha256"
	"math/big"
)

// Alphabet used in the encoding.
const Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var (
	base   big.Int
	digits ['z' - '1' + 1]big.Int
)

func init() {
	base.SetInt64(int64(len(Alphabet)))
	for i := range digits {
		digits[i].SetInt64(-1)
	}
	for i, c := range Alphabet {
		digits[c-'1'].SetInt64(int64(i))
	}
}

// DecodeAppend decodes src and returns it appended to dst.
// If the format is invalid, it returns nil.
func DecodeAppend(dst []byte, src string) []byte {
	for len(src) > 0 && src[0] == '1' {
		dst = append(dst, 0)
		src = src[1:]
	}
	var v big.Int
	for i, c := range src {
		if c < '1' || c > 'z' {
			return nil
		}
		d := &digits[c-'1']
		if d.Sign() < 0 {
			return nil
		}
		if i > 0 {
			v.Mul(&v, &base)
		}
		v.Add(&v, d)
	}
	return append(dst, v.Bytes()...)
}

// DecodeAppendCheck performs Base58Check decode of src, and returns it appended to dst.
// If the format is invalid or check fails, it returns nil.
func DecodeAppendCheck(dst []byte, src string) []byte {
	dst = DecodeAppend(dst, src)
	if len(dst) < 4 {
		return nil
	}
	d := dst[:len(dst)-4]
	c := sha256.Sum256(d)
	c = sha256.Sum256(c[:])
	for i, b := range dst[len(dst)-4:] {
		if b != c[i] {
			return nil
		}
	}
	return d
}

// AddressVersion returns version value of cryptocurrency address a.
// If the address is invalid, it returns -1.
func AddressVersion(a string) int {
	if len(a) < 27 || len(a) > 35 {
		return -1
	}
	d := DecodeAppendCheck(make([]byte, 0, 26), a)
	if len(d) != 21 {
		return -1
	}
	return int(d[0])
}
