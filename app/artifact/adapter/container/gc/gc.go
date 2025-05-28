// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package gc

import (
	"context"
	"fmt"

	"github.com/easysoft/gitfox/app/artifact/adapter"
	"github.com/easysoft/gitfox/app/artifact/adapter/container/schema"
	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/pkg/storage"
	"github.com/easysoft/gitfox/types"

	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
)

func markManifestReferences(ctx context.Context, dgst digest.Digest, manifestService *manifestStore, ingester func(digest.Digest)) error {
	manifest, err := manifestService.Get(ctx, dgst)
	if err != nil {
		return fmt.Errorf("failed to retrieve manifest for digest %v: %v", dgst, err)
	}

	log.Ctx(ctx).Debug().Msgf("found manifest for digest %v", dgst)

	descriptors := manifest.References()
	for _, descriptor := range descriptors {
		log.Ctx(ctx).Debug().Msgf("found reference %s for manifest for manifest %s", descriptor.Digest.String(), dgst.String())
		// ensure digest marked
		ingester(descriptor.Digest)

		if slices.Contains(schema.MediaList, descriptor.MediaType) {
			// is manifest, recurse mark
			err = markManifestReferences(ctx, descriptor.Digest, manifestService, ingester)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type manifestStore struct {
	artStore  store.ArtifactStore
	fileStore storage.ContentStorage

	blobGetFn func(dgst digest.Digest) (*types.ArtifactAssetExtendBlob, bool)
}

func (m *manifestStore) Get(ctx context.Context, dgst digest.Digest) (schema.Manifest, error) {
	blob, ok := m.blobGetFn(dgst)
	if !ok {
		return nil, fmt.Errorf("digest not found")
	}

	b, e := m.fileStore.Get(ctx, adapter.BlobPath(blob.Ref))
	if e != nil {
		return nil, e
	}
	manifest, e := schema.UnmarshalManifest(blob.ContentType, b)
	if e != nil {
		return nil, e
	}
	return manifest, nil
}
