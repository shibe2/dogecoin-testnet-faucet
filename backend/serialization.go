// SPDX-License-Identifier: AGPL-3.0-or-later

package faucet

import (
	"encoding/base64"

	"gopkg.in/yaml.v3"
)

// Bytes is a byte slice that is encoded in Base64 in YAML.
type Bytes []byte

func (self Bytes) MarshalYAML() (interface{}, error) {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!binary",
		Value: base64.StdEncoding.EncodeToString(self),
	}, nil
}

func (self *Bytes) UnmarshalYAML(value *yaml.Node) error {
	s := new(string)
	err := value.Decode(s)
	if err != nil {
		return err
	}
	*self = []byte(*s)
	return nil
}
