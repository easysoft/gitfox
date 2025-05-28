// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package schema2

import (
	"encoding/json"

	"github.com/easysoft/gitfox/app/artifact/adapter/container/schema"

	"github.com/opencontainers/image-spec/specs-go"
)

const (
	MediaTypeManifest = "application/vnd.docker.distribution.manifest.v2+json"
)

func init() {
	unmarshalFn := func(b []byte) (schema.Manifest, error) {
		i := new(DeserializedManifest)
		if err := i.UnmarshalJSON(b); err != nil {
			return Manifest{}, err
		}
		return i, nil
	}
	err := schema.Register(MediaTypeManifest, unmarshalFn)
	if err != nil {
		panic(err)
	}
}

type Manifest struct {
	specs.Versioned
	MediaType string `json:"mediaType,omitempty"`

	// Config references the image configuration as a blob.
	Config schema.Descriptor `json:"config"`

	// Layers lists descriptors for the layers referenced by the
	// configuration.
	Layers []schema.Descriptor `json:"layers"`
}

// References returns the descriptors of this manifests references.
func (m Manifest) References() []schema.Descriptor {
	references := make([]schema.Descriptor, 0, 1+len(m.Layers))
	references = append(references, m.Config)
	references = append(references, m.Layers...)
	return references
}

func (m Manifest) GetMediaType() string {
	return m.MediaType
}

func (m Manifest) GetKind() schema.Kind {
	return schema.KindImage
}

func (m Manifest) GetLayers() []schema.Descriptor {
	return m.Layers
}

type DeserializedManifest struct {
	Manifest
}

func (d *DeserializedManifest) UnmarshalJSON(b []byte) error {
	var obj Manifest
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}
	d.Manifest = obj
	return nil
}
