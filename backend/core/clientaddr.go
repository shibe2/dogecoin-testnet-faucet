// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"errors"
	"net"
)

var ErrInvalidClientAddress = errors.New("invalid client IP address")

// ClientRLAddr transforms IP address into a form used for claim rate limiting.
// For IPv4 it is 4 zero bytes followed by IP address bytes.
// For IPv6 it is first 8 bytes of IP address.
func ClientRLAddr(ip net.IP) [8]byte {
	var a [8]byte
	ip = ip.To16()
	if len(ip) != net.IPv6len {
		return a
	}
	ip4 := ip.To4()
	var xor byte
	if len(ip4) == 0 && ip[0] == 0x20 {
		switch ip[1] {
		case 0x00:
			if ip[2] == 0x00 && ip[3] == 0x00 { // Teredo
				ip4 = ip[12:]
				xor = 0xFF
			}
		case 0x02:
			ip4 = ip[2:6] // 6to4
		}
	}
	if len(ip4) == net.IPv4len {
		a[4] = ip4[0] ^ xor
		a[5] = ip4[1] ^ xor
		a[6] = ip4[2] ^ xor
		a[7] = ip4[3] ^ xor
	} else {
		copy(a[:], ip)
	}
	return a
}

// ParseClientAddr parses client IP address with optional port number.
func ParseClientAddr(client string) (net.IP, error) {
	h, _, err := net.SplitHostPort(client)
	if err == nil {
		client = h
	}
	ip := net.ParseIP(client)
	if len(ip) == 0 {
		if err == nil {
			err = ErrInvalidClientAddress
		}
		return nil, err
	}
	return ip.To16(), nil
}
