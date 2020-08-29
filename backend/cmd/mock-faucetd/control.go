// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

import (
	"faucet"
)

type controlData struct {
	AddressVersions  []uint
	Amount           float64
	Error, Msg, Wait string
	Errors           []string
}

var controlTemplate = `
<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml">
	<head>
		<meta charset="UTF-8"/>
		<title>Mock Faucet Back-end</title>
	</head>
	<body><div class="container">
		<h1>Mock Faucet Back-end</h1>
{{- if .Msg}}
		<div class="alert alert-warning" role="alert">{{.Msg}}</div>
{{- end}}
		<form method="post">
			<div class="form-group">
				<label for="addressVersions">Address Versions</label>:
				<input class="form-control" id="addressVersions" name="addressVersions" value="{{.AddressVersions}}"/>
			</div>
			<div class="form-group">
				<label for="amount">Amount</label>:
				<input class="form-control" id="amount" name="amount" value="{{.Amount}}"/>
			</div>
			<div class="form-group">
				<label for="error">Error</label>:
				<select class="form-control" id="error" name="error">
					<option {{- if not .Error}} selected="selected"{{end}} value="">No error</option>
{{- $e := .Error}}{{range .Errors}}
					<option {{- if eq . $e}} selected="selected"{{end}} value="{{.}}">{{.}}</option>
{{- end}}
				</select>
			</div>
			<div class="form-group">
				<label for="wait">Wait until</label>:
				<input class="form-control" id="wait" name="wait" value="{{.Wait}}"/>
			</div>
			<button class="btn btn-primary" type="submit">Set</button>
		</form>
	</div></body>
</html>
`[1:]

type controlHandler struct {
	f *mockFaucet
	t *template.Template
}

var errMock = errors.New("test error")

func (self controlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xhtml+xml;charset=UTF-8")
	d := &controlData{
		AddressVersions: self.f.avs,
		Amount:          self.f.amt,
		Errors: []string{
			"FailedToSend",
			"InternalError",
			"InvalidToken",
			"InvalidValue",
			"ServicePaused",
			"ServiceUnavailable",
		},
	}
	v := r.FormValue("addressVersions")
	if len(v) > 0 {
		v = strings.TrimSpace(v)
		if len(v) > 0 && v[0] == '[' {
			v = v[1:]
		}
		if len(v) > 0 && v[len(v)-1] == ']' {
			v = v[:len(v)-1]
		}
		avs1 := strings.Fields(v)
		avs2 := make([]uint, 0, len(avs1))
		var err error
		for _, v1 := range avs1 {
			var v2 uint64
			v2, err = strconv.ParseUint(v1, 0, 8)
			if err != nil {
				if len(d.Msg) > 0 {
					d.Msg += " "
				}
				d.Msg += "Invalid address version: " + err.Error() + "."
				break
			}
			avs2 = append(avs2, uint(v2))
		}
		if err == nil {
			self.f.avs = avs2
			d.AddressVersions = avs2
		}
	}
	v = r.FormValue("amount")
	if len(v) > 0 {
		a, err := strconv.ParseFloat(v, 64)
		if err != nil {
			if len(d.Msg) > 0 {
				d.Msg += " "
			}
			d.Msg += "Invalid amount: " + err.Error() + "."
		} else {
			self.f.amt = a
			d.Amount = a
		}
	}
	d.Error = r.FormValue("error")
	switch d.Error {
	case "":
		self.f.err = nil
	case "FailedToSend":
		self.f.err = faucet.SendError{Err: errMock}
	case "InternalError":
		self.f.err = errMock
	case "InvalidToken":
		self.f.err = faucet.ErrInvalidToken
	case "InvalidValue":
		self.f.err = faucet.ErrInvalidRecipient
	case "ServicePaused":
		self.f.err = faucet.ErrPaused
	case "ServiceUnavailable":
		self.f.err = faucet.ServiceUnavailableError{Err: errMock}
	default:
		if len(d.Msg) > 0 {
			d.Msg += " "
		}
		d.Msg = "Unknown error: " + d.Error + "."
		d.Error = ""
	}
	v = r.FormValue("wait")
	if len(v) > 0 {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			if len(d.Msg) > 0 {
				d.Msg += " "
			}
			d.Msg = "Invalid wait time: " + err.Error() + "."
		} else {
			self.f.wait = t
		}
	} else {
		self.f.wait = time.Time{}
	}
	if self.f.wait.IsZero() {
		d.Wait = time.Now().Format(time.RFC3339)
	} else {
		d.Wait = self.f.wait.Format(time.RFC3339)
	}
	err := self.t.Execute(w, d)
	if err != nil {
		log.Println("failed to send control page:", err)
	}
}

func newControlHandler(f *mockFaucet) (controlHandler, error) {
	t, err := template.New("control").Parse(controlTemplate)
	if err != nil {
		return controlHandler{}, err
	}
	return controlHandler{
		f: f,
		t: t,
	}, nil
}
