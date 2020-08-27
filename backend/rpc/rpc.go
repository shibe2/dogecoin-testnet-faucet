// SPDX-License-Identifier: AGPL-3.0-or-later

// Package rpc contains Dogecoin Core RPC client.
package rpc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

import (
	"faucet"
)

type RPCConfig struct{ URL, Username, Password, CookieFile string }

var errInvalidCookie = errors.New("invalid RPC cookie")

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (self RPCError) Error() string { return fmt.Sprintf("RPC error %v %q", self.Code, self.Message) }

type rpcReply struct {
	Result interface{} `json:"result"`
	Error  RPCError    `json:"error"`
	ID     uint32      `json:"id"`
}

func pipeJSON(v interface{}, w *io.PipeWriter) {
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		w.CloseWithError(err)
	} else {
		w.Close()
	}
}

// RPCClient implements Bank interface using Dogecoin Core wallet.
type RPCClient struct {
	bal  float64
	balx time.Time
	cfg  RPCConfig
	id   uint32
	m    sync.Mutex
}

func (self *RPCClient) cacheBalance(b float64) {
	var x time.Time
	if !math.IsNaN(b) {
		x = time.Now().Add(time.Minute)
	}
	self.m.Lock()
	defer self.m.Unlock()
	self.bal = b
	self.balx = x
}

func (self *RPCClient) cachedBalance() float64 {
	self.m.Lock()
	defer self.m.Unlock()
	if time.Now().After(self.balx) {
		return math.NaN()
	}
	return self.bal
}

func (self *RPCClient) readCookie() (un, pw string, err error) {
	if len(self.cfg.CookieFile) == 0 {
		return
	}
	f, err := os.Open(self.cfg.CookieFile)
	if err != nil {
		return
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()
	s := bufio.NewScanner(f)
	if s.Scan() {
		c := strings.Split(s.Text(), ":")
		switch len(c) {
		case 0:
		case 2:
			un = c[0]
			pw = c[1]
		default:
			err = errInvalidCookie
			return
		}
	}
	err = s.Err()
	if err != nil {
		return
	}
	err = f.Close()
	f = nil
	return
}

func (self *RPCClient) rpc(ctx context.Context, method string, params ...interface{}) (*rpcReply, error) {
	rr, rw := io.Pipe()
	var hreq *http.Request
	var err error
	if ctx != nil {
		hreq, err = http.NewRequestWithContext(ctx, "POST", self.cfg.URL, rr)
	} else {
		hreq, err = http.NewRequest("POST", self.cfg.URL, rr)
	}
	if err != nil {
		return nil, err
	}
	un, pw, err := self.readCookie()
	if err != nil {
		log.Println("failed to read RPC cookie:", err)
	}
	if len(un) == 0 && len(pw) == 0 {
		un = self.cfg.Username
		pw = self.cfg.Password
	}
	if len(un) > 0 || len(pw) > 0 {
		hreq.SetBasicAuth(un, pw)
	}
	jreq := &struct {
		Method string        `json:"method"`
		Params []interface{} `json:"params"`
		ID     uint32        `json:"id"`
	}{
		Method: method,
		Params: params,
		ID:     atomic.AddUint32(&self.id, 1),
	}
	go pipeJSON(jreq, rw)
	hres, err := http.DefaultClient.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer func() {
		if hres != nil {
			hres.Body.Close()
		}
	}()
	jres := new(rpcReply)
	err = json.NewDecoder(hres.Body).Decode(jres)
	if jres.Error.Code == 0 && hres.StatusCode != http.StatusOK {
		return jres, fmt.Errorf("RPC HTTP status %v", hres.StatusCode)
	}
	if err != nil {
		return nil, err
	}
	err = hres.Body.Close()
	hres = nil
	if err != nil {
		return jres, err
	}
	if jreq.ID != jres.ID {
		return jres, fmt.Errorf("RPC request identifier mismatch: request %v reply %v", jreq.ID, jres.ID)
	}
	return jres, nil
}

func (self *RPCClient) Balance(ctx context.Context) (float64, error) {
	bal := self.cachedBalance()
	if !math.IsNaN(bal) {
		return bal, nil
	}
	res, err := self.rpc(ctx, "getbalance")
	if err != nil {
		return 0, err
	}
	if res.Error.Code != 0 {
		return 0, res.Error
	}
	bal, ok := res.Result.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected RPC result type: %T", res.Result)
	}
	self.cacheBalance(bal)
	return bal, nil
}

func (self *RPCClient) Send(ctx context.Context, recipient string, amount float64) (string, error) {
	res, err := self.rpc(ctx, "sendtoaddress", recipient, amount)
	self.cacheBalance(math.NaN())
	if err != nil {
		return "", err
	}
	switch res.Error.Code {
	case 0:
	case -5:
		return "", faucet.ErrInvalidRecipient
	case -6:
		return "", faucet.ErrNoFunds
	default:
		return "", res.Error
	}
	tx, ok := res.Result.(string)
	if !ok {
		return "", fmt.Errorf("unexpected RPC result type: %T", res.Result)
	}
	return tx, nil
}

func NewRPCClient(cfg *RPCConfig) (*RPCClient, error) {
	_, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}
	return &RPCClient{cfg: *cfg}, nil
}
