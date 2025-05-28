// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package model

import (
	"context"
	"errors"
	"strings"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/types"
)

func AddTreeNode(ctx context.Context, artStore store.ArtifactStore, pkg *types.ArtifactPackage, ver *types.ArtifactVersion) error {
	newNode := types.ArtifactTreeNode{
		OwnerID: pkg.OwnerID,
		Name:    ver.Version,
		Type:    types.ArtifactTreeNodeTypeVersion,
		Format:  pkg.Format,
	}

	var err error
	newNode.Path, err = BuildPath(pkg.Namespace, pkg.Name, ver.Version)
	if err != nil {
		return err
	}

	if err = artStore.Nodes().RecurseCreate(ctx, &newNode); err != nil {
		return err
	}
	return nil
}

func BuildPath(namespace, name, version string) (string, error) {
	if namespace == "" && name == "" {
		return "", errors.New("empty namespace and name")
	}

	parts := make([]string, 0)
	if namespace != "" {
		parts = strings.Split(namespace, ".")
	}

	if name == "" {
		return "/" + strings.Join(parts, "/"), nil
	}
	parts = append(parts, name)

	if version == "" {
		return "/" + strings.Join(parts, "/"), nil
	}
	parts = append(parts, version)

	return "/" + strings.Join(parts, "/"), nil
}
