// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package dbtx

import (
	"context"
	"errors"

	"github.com/easysoft/gitfox/store"
	"github.com/easysoft/gitfox/store/database/dbtx"
)

// TxOptionRetryCount transaction option allows setting number of transaction executions reties.
// A transaction started with TxOptLock will be automatically retried in case of version conflict error.
type TxOptionRetryCount int

// TxOptionResetFunc transaction provides a function that will be executed before the transaction retry.
// A transaction started with TxOptLock will be automatically retried in case of version conflict error.
type TxOptionResetFunc func()

// TxOptLock runs the provided function inside a database transaction. If optimistic lock error occurs
// during the operation, the function will retry the whole transaction again (to the maximum of 5 times,
// but this can be overridden by providing an additional TxOptionRetryCount option).
func TxOptLock(ctx context.Context,
	tx dbtx.Transactor,
	txFn func(ctx context.Context) error,
	opts ...interface{},
) (err error) {
	tries := 5
	var resetFuncs []func()
	for _, opt := range opts {
		if n, ok := opt.(TxOptionRetryCount); ok {
			tries = int(n)
		}
		if fn, ok := opt.(TxOptionResetFunc); ok {
			resetFuncs = append(resetFuncs, fn)
		}
	}

	for try := 0; try < tries; try++ {
		err = tx.WithTx(ctx, txFn, opts...)
		if !errors.Is(err, store.ErrVersionConflict) {
			break
		}

		for _, fn := range resetFuncs {
			fn()
		}
	}

	return
}
