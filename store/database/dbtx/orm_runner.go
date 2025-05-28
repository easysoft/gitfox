// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package dbtx

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

type ormRunnerDB struct {
	db *gorm.DB
	//mx locker
}

func NewOrm(db *gorm.DB) Transactor {
	return &ormRunnerDB{db: db}
}

func (r ormRunnerDB) WithTx(ctx context.Context, txFn func(context.Context) error, opts ...interface{}) error {
	var txOpts *sql.TxOptions
	for _, opt := range opts {
		if v, ok := opt.(*sql.TxOptions); ok {
			txOpts = v
		}
	}

	if txOpts == nil {
		txOpts = TxDefault
	}

	//if txOpts.ReadOnly {
	//	r.mx.RLock()
	//	defer r.mx.RUnlock()
	//} else {
	//	r.mx.Lock()
	//	defer r.mx.Unlock()
	//}

	tx := r.db.Begin(txOpts)

	rtx := &ormRunnerTx{
		DB:       tx,
		commit:   false,
		rollback: false,
	}

	defer func() {
		if rtx.commit || rtx.rollback {
			return
		}
		_ = tx.Rollback() // ignoring the rollback error
	}()

	err := txFn(context.WithValue(ctx, ctxKeyTx{}, rtx))
	if err != nil {
		return err
	}

	if !rtx.commit && !rtx.rollback {
		err = rtx.Commit()
		if errors.Is(err, sql.ErrTxDone) {
			// Check if the transaction failed because of the context, if yes return the ctx error.
			if ctxErr := ctx.Err(); errors.Is(ctxErr, context.Canceled) || errors.Is(ctxErr, context.DeadlineExceeded) {
				err = ctxErr
			}
		}
	}

	return err
}

// ormRunnerTx executes gorm database transaction calls.
// Locking is not used because ormRunnerDB locks the entire transaction.
type ormRunnerTx struct {
	*gorm.DB
	commit   bool
	rollback bool
}

func (r *ormRunnerTx) Commit() error {
	err := r.DB.Commit().Error
	if err == nil {
		r.commit = true
	}
	return err
}

func (r *ormRunnerTx) Rollback() error {
	err := r.DB.Rollback().Error
	if err == nil {
		r.rollback = true
	}
	return err
}

func GetOrmAccessor(ctx context.Context, db *gorm.DB) *gorm.DB {
	if a, ok := ctx.Value(ctxKeyTx{}).(*ormRunnerTx); ok {
		return a.DB
	}
	return db.WithContext(ctx)
}
