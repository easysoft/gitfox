// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package extend

import (
	"github.com/easysoft/gitfox/app/pipeline/storage"
	"github.com/easysoft/gitfox/app/store"

	"github.com/google/wire"
)

// WireSet provides a wire set for this package.
var WireSet = wire.NewSet(
	ProvideDynamicStageProvider,
)

// ProvideDynamicStageProvider provides an stage dynamic env provider.
func ProvideDynamicStageProvider(
	store storage.PipelineStorage,
	stageStore store.StageStore,
) *StageDynamicEnvProvider {
	return NewStageDynamicEnvProvider(store, stageStore)
}
