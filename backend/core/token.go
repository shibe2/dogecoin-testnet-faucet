// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"net"
)

// TokenIntervalSec is the interval in seconds of token change for the same client address.
const TokenIntervalSec = 60 * 60

// TokenCipher used to encode time and client address in a token.
type TokenCipher = cipher.Block

func isTokenChr(c rune) bool {
	if c >= '0' && c <= '9' {
		return true
	}
	if c >= 'A' && c <= 'Z' {
		return true
	}
	if c >= 'a' && c <= 'z' {
		return true
	}
	return false
}

func genTokenBytes(client net.IP, c TokenCipher, t uint64, tb, ts []byte) int {
	for i := range tb {
		tb[i] = byte(t)
		t /= 256
	}
	c.Encrypt(tb, tb)
	for len(client) > 0 {
		for i, b := range client {
			if i >= len(tb) {
				break
			}
			tb[i] ^= b
		}
		if len(client) < len(tb) {
			client = nil
		} else {
			client = client[len(tb):]
		}
		c.Encrypt(tb, tb)
	}
	base64.RawStdEncoding.Encode(ts, tb)
	l := 0
	for _, b := range ts {
		if isTokenChr(rune(b)) {
			ts[l] = b
			l++
		}
	}
	if l > 0 && len(tb)%3 != 0 {
		l--
	}
	return l
}

// CheckToken checks if the token is valid for given client address.
// Both tokens for current and previous interval are valid.
func CheckToken(client net.IP, token string, c TokenCipher) bool {
	for _, c := range token {
		if !isTokenChr(c) {
			return false
		}
	}
	t := uint64(Now().Unix() / TokenIntervalSec)
	client = client.To16()
	tb := make([]byte, c.BlockSize()+base64.RawStdEncoding.EncodedLen(c.BlockSize()))
	ts := tb[c.BlockSize():]
	tb = tb[:c.BlockSize()]
	for dt := uint64(0); dt < 2; dt++ {
		l := genTokenBytes(client, c, t-dt, tb, ts)
		if l != len(token) {
			continue
		}
		if string(ts[:l]) == token {
			return true
		}
	}
	return false
}

// GenToken generates current token for given client address.
func GenToken(client net.IP, c TokenCipher) string {
	tb := make([]byte, c.BlockSize()+base64.RawStdEncoding.EncodedLen(c.BlockSize()))
	ts := tb[c.BlockSize():]
	tb = tb[:c.BlockSize()]
	l := genTokenBytes(client.To16(), c, uint64(Now().Unix()/TokenIntervalSec), tb, ts)
	return string(ts[:l])
}

// GenTokenKey generates cryptographic key suitable for NewTokenCipher.
func GenTokenKey() ([]byte, error) {
	k := make([]byte, 16)
	l := 0
	for {
		n, err := rand.Read(k[l:])
		l += n
		if l >= len(k) {
			return k, nil
		}
		if err != nil {
			return nil, err
		}
	}
}

// NewTokenCipher returns cipher instance that can be used to generate tokens.
func NewTokenCipher(key []byte) (TokenCipher, error) { return aes.NewCipher(key) }
