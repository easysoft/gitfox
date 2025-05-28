// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package ocischema

import (
	"encoding/json"

	"github.com/easysoft/gitfox/app/artifact/adapter/container/schema"

	"github.com/opencontainers/image-spec/specs-go"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func init() {
	unmarshalFn := func(b []byte) (schema.Manifest, error) {
		i := new(DeserializedIndex)
		if err := i.UnmarshalJSON(b); err != nil {
			return Manifest{}, err
		}
		return i, nil
	}
	err := schema.Register(v1.MediaTypeImageIndex, unmarshalFn)
	if err != nil {
		panic(err)
	}
}

var _ schema.Manifest = (*Index)(nil)

type Index struct {
	specs.Versioned
	MediaType string `json:"mediaType,omitempty"`
	Manifests []schema.Descriptor
}

func (i Index) GetMediaType() string {
	return i.MediaType
}

func (i Index) GetKind() schema.Kind {
	return schema.KindImageIndex
}

func (i Index) References() []schema.Descriptor { return i.Manifests }

type DeserializedIndex struct {
	Index
}

func (d *DeserializedIndex) UnmarshalJSON(b []byte) error {
	var obj Index
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}
	d.Index = obj
	return nil
}
