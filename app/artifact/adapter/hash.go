// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package adapter

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
)

type HashWriter struct {
	md5    hash.Hash
	sha1   hash.Hash
	sha256 hash.Hash
	sha512 hash.Hash

	w io.Writer
}

func NewHashWriter() *HashWriter {
	h := HashWriter{
		md5:    md5.New(),
		sha1:   sha1.New(),
		sha256: sha256.New(),
		sha512: sha512.New(),
	}

	h.w = io.MultiWriter(h.md5, h.sha1, h.sha256, h.sha512)
	return &h
}

func (h *HashWriter) Write(p []byte) (int, error) {
	return h.w.Write(p)
}

func (h *HashWriter) Sum() *Hash {
	return &Hash{
		Md5:    hex.EncodeToString(h.md5.Sum(nil)),
		Sha1:   hex.EncodeToString(h.sha1.Sum(nil)),
		Sha256: hex.EncodeToString(h.sha256.Sum(nil)),
		Sha512: hex.EncodeToString(h.sha512.Sum(nil)),
	}
}

type Hash struct {
	Md5    string `json:"md5"`
	Sha1   string `json:"sha1"`
	Sha256 string `json:"sha256"`
	Sha512 string `json:"sha512"`
}

func (h *Hash) String() string {
	return fmt.Sprintf(`{"md5":"%s", "sha1":"%s", "sha256":"%s", "sha512":"%s"}`, h.Md5, h.Sha1, h.Sha256, h.Sha512)
}
