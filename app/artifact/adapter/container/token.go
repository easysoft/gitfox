// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package container

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type hmacKey string

var errInvalidSecret = fmt.Errorf("invalid secret")

// unpackUploadState unpacks and validates the blob upload state from the
// token, using the hmacKey secret.
func (secret hmacKey) unpackUploadState(token string) (blobUploadState, error) {
	var state blobUploadState

	tokenBytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return state, err
	}
	mac := hmac.New(sha256.New, []byte(secret))

	if len(tokenBytes) < mac.Size() {
		return state, errInvalidSecret
	}

	macBytes := tokenBytes[:mac.Size()]
	messageBytes := tokenBytes[mac.Size():]

	mac.Write(messageBytes)
	if !hmac.Equal(mac.Sum(nil), macBytes) {
		return state, errInvalidSecret
	}

	if err := json.Unmarshal(messageBytes, &state); err != nil {
		return state, err
	}

	return state, nil
}

// packUploadState packs the upload state signed with and hmac digest using
// the hmacKey secret, encoding to url safe base64. The resulting token can be
// used to share data with minimized risk of external tampering.
func (secret hmacKey) packUploadState(lus blobUploadState) (string, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	p, err := json.Marshal(lus)
	if err != nil {
		return "", err
	}

	mac.Write(p)

	return base64.URLEncoding.EncodeToString(append(mac.Sum(nil), p...)), nil
}
