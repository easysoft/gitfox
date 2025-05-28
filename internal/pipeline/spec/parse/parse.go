// Copyright 2022 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parse

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	schema "github.com/easysoft/gitfox/internal/pipeline/spec"

	"github.com/ghodss/yaml"
)

// Parse parses the configuration from io.Reader r.
func Parse(r io.Reader) (*schema.Config, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	b, err = yaml.YAMLToJSON(b)
	if err != nil {
		return nil, err
	}
	out := new(schema.Config)
	err = json.Unmarshal(b, out)
	return out, err
}

// ParseBytes parses the configuration from bytes b.
func ParseBytes(b []byte) (*schema.Config, error) {
	return Parse(
		bytes.NewBuffer(b),
	)
}

// ParseString parses the configuration from string s.
func ParseString(s string) (*schema.Config, error) {
	return ParseBytes(
		[]byte(s),
	)
}

// ParseFile parses the configuration from path p.
func ParseFile(p string) (*schema.Config, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// ParseMulti parses a multi-document configuration
// from io.Reader r.
func ParseMulti(r io.Reader) ([]*schema.Config, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseMultiBytes(b)
}

// ParseMultiBytes parses a multi-document configuration
// from bytes b.
func ParseMultiBytes(b []byte) ([]*schema.Config, error) {
	var out []*schema.Config
	parts := bytes.Split(b, []byte("\n---\n"))
	for _, part := range parts {
		resource, err := ParseBytes(part)
		if err != nil {
			return nil, err
		}
		out = append(out, resource)
	}
	return out, nil
}

// ParseMultiFile parses a multi-document configuration
// from path p.
func ParseMultiFile(p string) ([]*schema.Config, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseMulti(f)
}
