// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifact

import (
	"context"
	"errors"

	"github.com/easysoft/gitfox/pkg/storage"
	storagedriver "github.com/easysoft/gitfox/pkg/storage/driver"
	"github.com/easysoft/gitfox/types"
)

func LoadContentStorage(ctx context.Context, c Config, storageConfig storage.Config) (storage.ContentStorage, error) {
	var err error
	var driver storagedriver.StorageDriver
	if c.Storage.Provider == types.StorageProviderLocal {
		driver, err = storage.NewDriver(ctx, string(storageConfig.Local.Driver), storageConfig.Local.Parameters)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("unsupported s3 storage")
	}

	return storage.NewCommonContentStore(ctx, driver, storage.WithPrefix(c.Storage.Prefix))
}
