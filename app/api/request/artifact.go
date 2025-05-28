// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package request

import (
	"net/http"
	"net/url"

	"github.com/easysoft/gitfox/types"
)

const (
	QueryArtifactFormat = "format"

	// container params
	ParamDigest    = "digest"
	PathParamName  = "name"
	PathParamUUID  = "uuid"
	PathParamRefer = "reference"

	QueryParamPackage = "package"
	QueryParamGroup   = "group"
	QueryParamVersion = "version"

	QueryParamNodeLevel = "level"
)

// ParseArtifactFilter extracts the artifact filter from the url.
func ParseArtifactFilter(r *http.Request) *types.ArtifactFilter {
	return &types.ArtifactFilter{
		Query:  ParseQuery(r),
		Order:  ParseOrder(r),
		Page:   ParsePage(r),
		Format: ParseArtifactFormat(r),
		Sort:   ParseSortSpace(r),
		Size:   ParseLimit(r),
	}
}

// ParseArtifactTreeFilter extracts the tree filter from the url.
func ParseArtifactTreeFilter(r *http.Request) *types.ArtifactTreeFilter {
	return &types.ArtifactTreeFilter{
		Path:   QueryParamOrDefault(r, QueryParamPath, types.ArtifactNodeRoot),
		Format: types.ArtifactFormat(QueryParamOrDefault(r, QueryArtifactFormat, string(types.ArtifactAllFormat))),
		Level:  QueryParamOrDefault(r, QueryParamNodeLevel, types.ArtifactNodeLevelAsset),
	}
}

func ParseArtifactFormat(r *http.Request) string {
	return r.URL.Query().Get(QueryArtifactFormat)
}

func ParseArtifactVersionFilter(r *http.Request) (*types.ArtifactVersionFilter, error) {
	pkgName, err := QueryParamOrError(r, QueryParamPackage)
	if err != nil {
		return nil, err
	}

	return &types.ArtifactVersionFilter{
		Query:   ParseQuery(r),
		Page:    ParsePage(r),
		Size:    ParseLimit(r),
		Package: pkgName,
		Group:   QueryParamOrDefault(r, QueryParamGroup, ""),
	}, nil
}

func GetContainerRepositoryFromPath(r *http.Request) (string, string, error) {
	rawSpace, err := PathParamOrError(r, "space")
	if err != nil {
		return "", "", err
	}

	space, err := url.PathUnescape(rawSpace)
	if err != nil {
		return "", "", err
	}

	repoName, err := PathParamOrError(r, PathParamName)
	if err != nil {
		return "", "", err
	}
	if err != nil {
		return "", "", err
	}

	return space, repoName, nil
}

func GetDigestFromPath(r *http.Request) (string, error) {
	return PathParamOrError(r, ParamDigest)
}

func GetUUIDFromPath(r *http.Request) (string, error) {
	return PathParamOrError(r, PathParamUUID)
}

func GetReferenceFromPath(r *http.Request) (string, error) {
	return PathParamOrError(r, PathParamRefer)
}
