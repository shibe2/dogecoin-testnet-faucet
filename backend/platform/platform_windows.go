// SPDX-License-Identifier: AGPL-3.0-or-later

package platform

import (
	"os"
	"path/filepath"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	modkernel32            = syscall.NewLazyDLL("kernel32.dll")
	procGetModuleFileNameW = modkernel32.NewProc("GetModuleFileNameW")
)

func getModuleFileName(module syscall.Handle, fn *uint16, len uint32) (n uint32, err error) {
	r0, _, e1 := syscall.Syscall(procGetModuleFileNameW.Addr(), 3, uintptr(module), uintptr(unsafe.Pointer(fn)), uintptr(len))
	n = uint32(r0)
	if n == 0 {
		if e1 != 0 {
			err = e1
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func OSProgName() string {
	fn1 := make([]uint16, syscall.MAX_PATH)
	n, err := getModuleFileName(0, &fn1[0], uint32(len(fn1)))
	if n == 0 || err != nil {
		return ""
	}
	for i, c := range fn1[:n] {
		if c != 0 {
			continue
		}
		fn2 := filepath.Base(string(utf16.Decode(fn1[:i])))
		switch fn2 {
		case "", ".", string(filepath.Separator):
			return ""
		}
		return fn2
	}
	return ""
}

func DefaultCookieFile() string {
	d, err := os.UserConfigDir()
	if len(d) == 0 {
		return ""
	}
	if err != nil {
		return ""
	}
	return filepath.Join(d, "Dogecoin\\testnet3\\.cookie")
}
