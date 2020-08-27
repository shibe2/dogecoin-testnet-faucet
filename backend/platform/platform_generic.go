// +build !linux,!windows

// SPDX-License-Identifier: AGPL-3.0-or-later

// Package platform contains OS-specific code.
package platform

func OSProgName() string { return "" }

func DefaultCookieFile() string { return "" }
