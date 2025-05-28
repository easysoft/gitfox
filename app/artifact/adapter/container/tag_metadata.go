// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package container

import (
	"context"

	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/artifact/adapter/container/schema"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/types"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type TagMetadataReader struct {
	artStore store.ArtifactStore
	view     *adapter.ViewDescriptor
}

type TagMetadata struct {
	Images []*ImageMeta `json:"images"`
}

type ImageMeta struct {
	Digest digest.Digest `json:"digest"`
	OS     string        `json:"os"`
	Arch   string        `json:"arch"`
	Size   int64         `json:"size"`
}

func NewTagMetadataReader(artStore store.ArtifactStore, view *adapter.ViewDescriptor) *TagMetadataReader {
	return &TagMetadataReader{artStore: artStore, view: view}
}

func (t *TagMetadataReader) Parse(ctx context.Context, mediaType string, digest digest.Digest) (*TagMetadata, error) {
	manifest, err := t.parseDigest(ctx, mediaType, digest)
	if err != nil {
		return nil, err
	}

	platforms := make([]*ImageMeta, 0)

	if manifest.GetKind() == schema.KindImageIndex {
		for _, reference := range manifest.References() {
			p := reference.Platform
			if p.OS == "unknown" {
				continue
			}

			nextManifest, e := t.parseDigest(ctx, reference.MediaType, reference.Digest)
			if e != nil {
				return nil, e
			}
			imageMani, e := t.parseImage(ctx, p, nextManifest)
			if e != nil {
				return nil, e
			}
			imageMani.Digest = reference.Digest

			platforms = append(platforms, imageMani)
		}
	} else {
		imgManifest, e := t.parseImage(ctx, nil, manifest)
		if e != nil {
			return nil, e
		}
		imgManifest.Digest = digest
		platforms = append(platforms, imgManifest)
	}

	return &TagMetadata{
		Images: platforms,
	}, nil
}

func (t *TagMetadataReader) parseImage(ctx context.Context, platform *v1.Platform, m schema.Manifest) (*ImageMeta, error) {
	im, err := schema.ToImage(m)
	if err != nil {
		return nil, err
	}

	var data = ImageMeta{}
	if platform != nil {
		data.OS = platform.OS
		data.Arch = platform.Architecture
	} else {
		data.OS = schema.DefaultOS
		data.Arch = schema.DefaultArch
	}

	for _, layer := range im.GetLayers() {
		data.Size = data.Size + layer.Size
	}
	return &data, nil
}

func (t *TagMetadataReader) parseDigest(ctx context.Context, mediaType string, digest digest.Digest) (schema.Manifest, error) {
	dbAsset, err := t.artStore.Assets().GetPath(ctx, digest.String(), types.ArtifactContainerFormat)
	if err != nil {
		return nil, err
	}

	dbBlob, err := t.artStore.Blobs().GetById(ctx, dbAsset.BlobID)
	if err != nil {
		return nil, err
	}

	b, err := t.view.Store.Get(ctx, adapter.BlobPath(dbBlob.Ref))
	if err != nil {
		return nil, err
	}

	return schema.UnmarshalManifest(mediaType, b)
}
