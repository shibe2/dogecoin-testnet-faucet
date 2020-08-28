// SPDX-License-Identifier: AGPL-3.0-or-later

// Package server implements HTTP API and file server.
package server

import (
	"context"
	"net/http"
	"strings"
	"time"
)

import (
	"faucet"
)

type ServerConfig struct {
	Listen, CertFile, KeyFile string
	APIPrefix, PubDir         string
	AllowOrigin               string
	UseFwdAddr                bool
}

type mHandler struct {
	h           http.Handler
	allowOrigin string
	useFwdAddr  bool
}

func (self *mHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(self.allowOrigin) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Origin", self.allowOrigin)
	}
	if self.useFwdAddr {
		for _, a := range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
			a = strings.TrimSpace(a)
			if len(a) > 0 {
				r.RemoteAddr = a
			}
		}
	}
	self.h.ServeHTTP(w, r)
}

type Server struct {
	certFile, keyFile string
	m                 *http.ServeMux
	s                 *http.Server
	sc                chan error
}

// Handle registers HTTP request handler for the given pattern.
func (self *Server) Handle(pattern string, handler http.Handler) { self.m.Handle(pattern, handler) }

// Serve listens on configured TCP address and serves HTTP requests. It returns when the server is stopped.
func (self *Server) Serve() error {
	var err error
	if len(self.certFile) > 0 || len(self.keyFile) > 0 {
		err = self.s.ListenAndServeTLS(self.certFile, self.keyFile)
	} else {
		err = self.s.ListenAndServe()
	}
	if err == http.ErrServerClosed {
		e, ok := <-self.sc
		if ok {
			return e
		}
	}
	return err
}

func (self *Server) Stop() {
	c := self.sc
	if c == nil {
		return
	}
	defer close(c)
	s := self.s
	if s == nil {
		return
	}
	ctx, cf := context.WithTimeout(context.Background(), time.Second)
	c <- s.Shutdown(ctx)
	cf()
	c <- s.Close()
}

func NewServer(cfg *ServerConfig, f faucet.Faucet) *Server {
	self := &Server{
		certFile: cfg.CertFile,
		keyFile:  cfg.KeyFile,
		m:        http.NewServeMux(),
		s: &http.Server{
			Addr: cfg.Listen,
		},
		sc: make(chan error, 2),
	}
	self.s.Handler = &mHandler{
		h:           self.m,
		allowOrigin: cfg.AllowOrigin,
		useFwdAddr:  cfg.UseFwdAddr,
	}
	if len(cfg.PubDir) > 0 {
		self.m.Handle("/", http.FileServer(http.Dir(cfg.PubDir)))
	}
	registerAPIServer(self.m, apiServer{faucet: f}, cfg.APIPrefix)
	return self
}
