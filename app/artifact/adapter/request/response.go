// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package request

import "net/http"

const (
	contentType    = "Content-Type"
	contentLength  = "Content-Length"
	disposition    = "Content-Disposition"
	etag           = "ETag"
	lastModify     = "Last-Modified"
	location       = "Location"
	headerRange    = "Range"
	HeadDockerUUID = "Docker-Upload-UUID"
)

type ResponseHeaderWriter interface {
	Write(w http.ResponseWriter)
}

type commonHttpResponseWriter struct {
	writeFn func(w http.ResponseWriter)
}

func (c *commonHttpResponseWriter) Write(w http.ResponseWriter) {
	c.writeFn(w)
}

func NewResponseWriter(fn func(w http.ResponseWriter)) ResponseHeaderWriter {
	return &commonHttpResponseWriter{writeFn: fn}
}
