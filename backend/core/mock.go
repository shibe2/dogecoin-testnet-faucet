// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"time"
)

// Now is a function that this package uses to get current time.
// It can be changed for purposes of testing.
var Now = time.Now
