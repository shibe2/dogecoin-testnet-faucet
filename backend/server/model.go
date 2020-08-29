// SPDX-License-Identifier: AGPL-3.0-or-later

// API types

package server

import (
	"time"
)

// ClaimRejected defines model for ClaimRejected.
type ClaimRejected struct {
	RejectReason string `json:"rejectReason"`

	// The client with this IP address cannot claim coins before the given time.
	Wait *time.Time `json:"wait,omitempty"`
}

// ClaimRequest defines model for ClaimRequest.
type ClaimRequest struct {

	// Cryptocurrency recipient address.
	Recipient string `json:"recipient"`

	// The token obtained from earlier API call.
	Token string `json:"token,omitempty"`
}

// ClaimSucceeded defines model for ClaimSucceeded.
type ClaimSucceeded struct {

	// Actual amount of coins sent.
	Amount float64 `json:"amount"`

	// Cryptocurrency transaction identifier (hash).
	TXID string `json:"txid"`
}

// Info defines model for Info.
type Info struct {

	// Accepted recipient address versions. If this parameter is absent, front-end should accept any address version.
	AddressVersions []uint `json:"addressVersions,omitempty"`

	// Expected giveaway amount. Actual amount may differ. Zero means dry faucet.
	Amount float64 `json:"amount"`

	// A token that must be passed to other API calls where specified. It is valid for at least 1 hour.
	Token string `json:"token,omitempty"`

	// The client with this IP address cannot claim coins before the given time.
	Wait *time.Time `json:"wait,omitempty"`
}

// InvalidRequest defines model for InvalidRequest.
type InvalidRequest struct {
	RequestErrors []RequestError `json:"requestErrors"`
}

// RequestError defines model for RequestError.
type RequestError struct {

	// The problem.
	Error string `json:"error"`

	// Which request parameter has the problem. This is absent if overall request is invalid.
	Parameter string `json:"parameter,omitempty"`
}

// RequestFailed defines model for RequestFailed.
type RequestFailed struct {
	Error string `json:"error"`
}

// ServiceUnavailable defines model for ServiceUnavailable.
type ServiceUnavailable struct {
	Error string `json:"error"`
}
