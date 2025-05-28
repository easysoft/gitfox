// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package enum

type ArtifactRepoKind string

const (
	ArtifactRepoKindStandalone  ArtifactRepoKind = "standalone"
	ArtifactRepoKindProduct     ArtifactRepoKind = "product"
	ArtifactRepoKindProductLine ArtifactRepoKind = "product_line"
	ArtifactRepoKindGitRepo     ArtifactRepoKind = "git_repo"
)

func (ArtifactRepoKind) Enum() []interface{} { return toInterfaceSlice(ArtifactRepoKinds) }

func (t ArtifactRepoKind) Sanitize() (ArtifactRepoKind, bool) {
	return Sanitize(t, GetAllArtifactRepoKinds)
}

func GetAllArtifactRepoKinds() ([]ArtifactRepoKind, ArtifactRepoKind) {
	return ArtifactRepoKinds, "" // No default value
}

var ArtifactRepoKinds = sortEnum([]ArtifactRepoKind{
	ArtifactRepoKindStandalone,
	ArtifactRepoKindProduct,
	ArtifactRepoKindProductLine,
	ArtifactRepoKindGitRepo,
})

func (ArtifactRepoKind) CreatableEnum() []interface{} {
	return toInterfaceSlice(ArtifactRepoCreatableKinds)
}

var ArtifactRepoCreatableKinds = []ArtifactRepoKind{ArtifactRepoKindStandalone}
