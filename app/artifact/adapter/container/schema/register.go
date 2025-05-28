// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package schema

import (
	"fmt"
)

var schemaMap = make(map[string]UnmarshalFunc)

var MediaList = make([]string, 0)

type UnmarshalFunc func([]byte) (Manifest, error)

func Register(mediaType string, fn UnmarshalFunc) error {
	if _, ok := schemaMap[mediaType]; ok {
		return fmt.Errorf("manifest media type '%s' is already registered", mediaType)
	}
	schemaMap[mediaType] = fn
	MediaList = append(MediaList, mediaType)
	return nil
}

func UnmarshalManifest(mediaType string, b []byte) (Manifest, error) {
	fn, ok := schemaMap[mediaType]
	if !ok {
		return nil, fmt.Errorf("unsupported manifest media type '%s'", mediaType)
	}
	return fn(b)
}
