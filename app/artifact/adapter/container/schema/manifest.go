// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package schema

import (
	"errors"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	DefaultOS   = "linux"
	DefaultArch = "amd64"
)

type Descriptor struct {
	// MediaType describe the type of the content. All text based formats are
	// encoded as utf-8.
	MediaType string `json:"mediaType,omitempty"`

	// Digest uniquely identifies the content. A byte stream can be verified
	// against this digest.
	Digest digest.Digest `json:"digest,omitempty"`

	// Size in bytes of content.
	Size int64 `json:"size,omitempty"`

	// URLs contains the source URLs of this content.
	URLs []string `json:"urls,omitempty"`

	// Annotations contains arbitrary metadata relating to the targeted content.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Platform describes the platform which the image in the manifest runs on.
	// This should only be used when referring to a manifest.
	Platform *v1.Platform `json:"platform,omitempty"`
}

type Kind string

const (
	KindImage      Kind = "image"
	KindImageIndex Kind = "index"
)

type Manifest interface {
	GetMediaType() string
	GetKind() Kind
	References() []Descriptor
}

type ImageManifest interface {
	Manifest
	GetLayers() []Descriptor
}

func ToImage(in Manifest) (ImageManifest, error) {
	if in.GetKind() != KindImage {
		return nil, errors.New("invalid kind")
	}
	im, ok := in.(ImageManifest)
	if !ok {
		return nil, errors.New("invalid image manifest")
	}
	return im, nil
}
