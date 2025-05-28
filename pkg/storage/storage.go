// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package storage

import (
	"context"

	"github.com/easysoft/gitfox/pkg/storage/driver"
	"github.com/easysoft/gitfox/pkg/storage/driver/factory"

	_ "github.com/easysoft/gitfox/pkg/storage/driver/filesystem"
	_ "github.com/easysoft/gitfox/pkg/storage/driver/inmemory"
)

func NewDriver(ctx context.Context, name string, parameters map[string]interface{}) (driver.StorageDriver, error) {
	return factory.Create(ctx, name, parameters)
}

type DriverGetter interface {
	Get(ctx context.Context, provider DriverType, config Config) (driver.StorageDriver, error)
}
