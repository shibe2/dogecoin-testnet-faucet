// SPDX-License-Identifier: AGPL-3.0-or-later

package faucet

import (
	"context"
	"errors"
	"net"
	"time"
)

var (
	ErrInvalidClientAddress = errors.New("invalid client IP address")
	ErrInvalidRecipient     = errors.New("invalid recipient address")
	ErrInvalidToken         = errors.New("invalid or missing token")
	ErrNoFunds              = errors.New("no funds in the bank")
	ErrPaused               = errors.New("service paused")
)

type MustWait struct{ Until time.Time }

func (self MustWait) Error() string { return "this client must wait until " + self.Until.String() }

type SendError struct{ Err error }

func (self SendError) Error() string { return "failed to send coins: " + self.Err.Error() }
func (self SendError) Unwrap() error { return self.Err }

type ServiceUnavailableError struct{ Err error }

func (self ServiceUnavailableError) Error() string { return "service unavailable: " + self.Err.Error() }
func (self ServiceUnavailableError) Unwrap() error { return self.Err }

// Bank provides funds for the faucet.
type Bank interface {
	// Balance available for giveaway.
	Balance(ctx context.Context) (float64, error)

	// Send coins. Returns cryptocurrency transaction identifier.
	Send(ctx context.Context, recipient string, amount float64) (string, error)
}

// Faucet implements core logic.
// Argument client is client IP address with optional TCP port number.
type Faucet interface {
	// AddressVersions returns accepted recipient address versions.
	// Empty result means unspecified versions are accepted.
	AddressVersions() []uint

	// Amount returns expected giveaway amount.
	Amount(ctx context.Context) (float64, error)

	// Claim checks validity of claim request and sends coins.
	// Returns actual amount of coins sent and cryptocurrency transaction identifier.
	Claim(ctx context.Context, client, recipient, token string) (amount float64, tx string, err error)

	// Token that must be supplied when claiming.
	// If empty then token is not required.
	Token(ctx context.Context, client string) (string, error)

	// WaitTime returns time point after which this client can claim again.
	// Returns zero if this client can claim now.
	WaitTime(ctx context.Context, client string) (time.Time, error)
}

// ClaimLogIter enumerates claim log records.
type ClaimLogIter interface {
	// Close should be called after using the iterator.
	Close() error

	// Get reads time and client IP address of current record.
	Get(t *time.Time, client *net.IP, amount *float64) error

	// Next advances to the next record. It should also be called before reading the first record.
	// Returns whether there is a record.
	Next() bool
}

// FaucetDB stores persistent data for the faucet.
type FaucetDB interface {
	// ClaimsSince returns all claim records since given time.
	ClaimsSince(t time.Time) (ClaimLogIter, error)

	// LogClaim adds log record about successful claim.
	LogClaim(t time.Time, client net.IP, recipient string, amount float64, tx []byte) error
}

// Alerter sends notifications about important events.
// Alert methods should be called once when the condition changes from false to true.
type Alerter interface {
	// BalanceAlert sends a notification about low balance.
	BalanceAlert(balance float64)

	// RateAlert sends a notification about excessive total giveaway rate
	RateAlert(amount float64, period time.Duration)
}
