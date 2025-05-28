// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package openai

import (
	"net/http"
	"strings"
)

// DefaultHeaderTransport is an http.RoundTripper that adds the given headers to
type DefaultHeaderTransport struct {
	Origin http.RoundTripper
	Header http.Header
}

// RoundTrip implements the http.RoundTripper interface.
func (t *DefaultHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, values := range t.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return t.Origin.RoundTrip(req)
}

// NewHeaders creates a new http.Header from the given slice of headers.
func NewHeaders(headers []string) http.Header {
	h := make(http.Header)
	for _, header := range headers {
		// split header into key and value with = as delimiter
		vals := strings.Split(header, "=")
		if len(vals) != 2 {
			continue
		}
		h.Add(vals[0], vals[1])
	}
	return h
}
