// SPDX-License-Identifier: AGPL-3.0-or-later

package platform

import (
	"os"
	"path/filepath"
)

func DefaultCookieFile() string {
	d, err := os.UserHomeDir()
	if len(d) == 0 {
		return ""
	}
	if err != nil {
		return ""
	}
	return filepath.Join(d, ".dogecoin/testnet3/.cookie")
}
