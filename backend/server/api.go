// SPDX-License-Identifier: AGPL-3.0-or-later

// API handlers

package server

import (
	"context"
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"path"
	"time"
)

import (
	"faucet"
)

func errorResponse(msg string, err error) interface{} {
	switch e := err.(type) {
	case faucet.SendError:
		log.Println(msg, e.Err)
		return &RequestFailed{Error: "FailedToSend"}
	case faucet.MustWait:
		res := &ClaimRejected{
			RejectReason: "MustWait",
			Wait:         new(time.Time),
		}
		*res.Wait = e.Until.UTC().Round(time.Second)
		return res
	case faucet.ServiceUnavailableError:
		log.Println(msg, e.Err)
		return &ServiceUnavailable{Error: "ServiceUnavailable"}
	}
	switch err {
	case faucet.ErrInvalidToken:
		return &ClaimRejected{RejectReason: "InvalidToken"}
	case faucet.ErrInvalidRecipient:
		return &InvalidRequest{RequestErrors: []RequestError{{
			Error:     "InvalidValue",
			Parameter: "recipient",
		}}}
	case faucet.ErrPaused:
		return &ServiceUnavailable{Error: "ServicePaused"}
	case faucet.ErrNoFunds:
		return &ServiceUnavailable{Error: "NoFunds"}
	}
	log.Println(msg, err)
	return &RequestFailed{Error: "InternalError"}
}

type apiServer struct {
	faucet faucet.Faucet
}

func (self apiServer) ClaimPost(ctx context.Context, client string, body *ClaimRequest) interface{} {
	if len(body.Recipient) == 0 {
		return &InvalidRequest{RequestErrors: []RequestError{{
			Error:     "MissingValue",
			Parameter: "recipient",
		}}}
	}
	a, tx, err := self.faucet.Claim(ctx, client, body.Recipient, body.Token)
	if err != nil {
		return errorResponse("failed to send coins:", err)
	}
	return &ClaimSucceeded{
		Amount: a,
		TXID:   tx,
	}
}

func (self apiServer) InfoGet(ctx context.Context, client string) interface{} {
	a, err := self.faucet.Amount(ctx)
	if err != nil {
		return errorResponse("failed to get giveaway amount:", err)
	}
	t, err := self.faucet.Token(ctx, client)
	if err != nil {
		return errorResponse("failed to generate token:", err)
	}
	w, err := self.faucet.WaitTime(ctx, client)
	if err != nil {
		return errorResponse("failed to get wait time:", err)
	}
	res := &Info{
		AddressVersions: self.faucet.AddressVersions(),
		Amount:          a,
		Token:           t,
	}
	if !w.IsZero() {
		res.Wait = new(time.Time)
		*res.Wait = w.UTC().Round(time.Second)
	}
	return res
}

func ensureContentType(w http.ResponseWriter, r *http.Request, ct string) bool {
	mt := r.Header.Get("Content-Type")
	if len(mt) == 0 {
		return true
	}
	mt, _, _ = mime.ParseMediaType(mt)
	if len(mt) == 0 {
		return true
	}
	if mt == ct {
		return true
	}
	http.Error(w, "Request media type must be "+ct, http.StatusUnsupportedMediaType)
	return false
}

type claimHandler struct{ s apiServer }

func (self claimHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "OPTIONS,POST")
	switch r.Method {
	case "OPTIONS":
	case "POST":
		if !ensureContentType(w, r, "application/json") {
			return
		}
		body := new(ClaimRequest)
		err := json.NewDecoder(r.Body).Decode(body)
		var res interface{}
		switch err.(type) {
		case nil:
			res = self.s.ClaimPost(r.Context(), r.RemoteAddr, body)
		case *json.InvalidUTF8Error:
			res = &InvalidRequest{RequestErrors: []RequestError{{Error: "InvalidFormat"}}}
		case *json.InvalidUnmarshalError:
			log.Println("failed to decode JSON to ClaimRequest:", err)
			res = &RequestFailed{Error: "InternalError"}
		case *json.SyntaxError:
			res = &InvalidRequest{RequestErrors: []RequestError{{Error: "InvalidFormat"}}}
		case *json.UnmarshalTypeError:
			res = &InvalidRequest{RequestErrors: []RequestError{{Error: "InvalidValue"}}}
		default:
			log.Println("failed receive /claim POST request body:", err)
			res = &InvalidRequest{RequestErrors: []RequestError{{Error: "InvalidFormat"}}}
		}
		var st int
		switch res.(type) {
		case *ClaimSucceeded:
			st = 200
		case *InvalidRequest:
			st = 400
		case *ClaimRejected:
			st = 403
		case *RequestFailed:
			st = 500
		case *ServiceUnavailable:
			st = 503
		default:
			log.Printf("unexpected /claim POST response type: %T", res)
			res = &RequestFailed{Error: "InternalError"}
			st = 500
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(st)
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Println("failed to send /claim POST response:", err)
		}
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}

type infoHandler struct{ s apiServer }

func (self infoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "GET,OPTIONS")
	switch r.Method {
	case "GET":
		res := self.s.InfoGet(r.Context(), r.RemoteAddr)
		var st int
		switch res.(type) {
		case *Info:
			st = 200
		case *RequestFailed:
			st = 500
		case *ServiceUnavailable:
			st = 503
		default:
			log.Printf("unexpected /info GET response type: %T", res)
			res = &RequestFailed{Error: "InternalError"}
			st = 500
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(st)
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Println("failed to send /info GET response:", err)
		}
	case "OPTIONS":
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}

func registerAPIServer(mux *http.ServeMux, s apiServer, prefix string) {
	{
		p := "/claim"
		if len(prefix) > 0 {
			p = path.Join(prefix, p)
		}
		h := claimHandler{s}
		if mux != nil {
			mux.Handle(p, h)
		} else {
			http.Handle(p, h)
		}
	}
	{
		p := "/info"
		if len(prefix) > 0 {
			p = path.Join(prefix, p)
		}
		h := infoHandler{s}
		if mux != nil {
			mux.Handle(p, h)
		} else {
			http.Handle(p, h)
		}
	}
}
