// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package publicaccess

import (
	"context"
	"fmt"

	"github.com/easysoft/gitfox/app/store"
	"github.com/easysoft/gitfox/store/database"
	"github.com/easysoft/gitfox/types/enum"

	"gorm.io/gorm"
)

var _ store.PublicAccessStore = (*PublicAccessStore)(nil)

// NewPublicAccessStore returns a new PublicAccessStore.
func NewPublicAccessStore(db *gorm.DB) *PublicAccessStore {
	return &PublicAccessStore{
		db: db,
	}
}

// PublicAccessStore implements store.PublicAccessStore backed by a relational database.
type PublicAccessStore struct {
	db *gorm.DB
}

func (p *PublicAccessStore) Find(
	ctx context.Context,
	typ enum.PublicResourceType,
	id int64,
) (bool, error) {
	var exists bool

	switch typ {
	case enum.PublicResourceTypeRepo:
		err := p.db.Table("public_access_repo").Where("public_access_repo_id = ?", id).Select("1").Scan(&exists).Error
		if err != nil {
			return false, database.ProcessGormSQLErrorf(ctx, err, "Select query failed")
		}
	case enum.PublicResourceTypeSpace:
		err := p.db.Table("public_access_space").Where("public_access_space_id = ?", id).Select("1").Scan(&exists).Error
		if err != nil {
			return false, database.ProcessGormSQLErrorf(ctx, err, "Select query failed")
		}
	default:
		return false, fmt.Errorf("public resource type %q is not supported", typ)
	}
	return exists, nil
}

func (p *PublicAccessStore) Create(
	ctx context.Context,
	typ enum.PublicResourceType,
	id int64,
) error {
	var sqlQuery string
	switch typ {
	case enum.PublicResourceTypeRepo:
		sqlQuery = `INSERT INTO public_access_repo(public_access_repo_id) VALUES(?)`
	case enum.PublicResourceTypeSpace:
		sqlQuery = `INSERT INTO public_access_space(public_access_space_id) VALUES(?)`
	default:
		return fmt.Errorf("public resource type %q is not supported", typ)
	}

	err := p.db.Exec(sqlQuery, id).Error
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "Insert query failed")
	}

	return nil
}

func (p *PublicAccessStore) Delete(
	ctx context.Context,
	typ enum.PublicResourceType,
	id int64,
) error {
	var sqlQuery string
	switch typ {
	case enum.PublicResourceTypeRepo:
		sqlQuery = `DELETE FROM public_access_repo WHERE public_access_repo_id = ?`
	case enum.PublicResourceTypeSpace:
		sqlQuery = `DELETE FROM public_access_space WHERE public_access_space_id = ?`
	default:
		return fmt.Errorf("public resource type %q is not supported", typ)
	}

	err := p.db.Exec(sqlQuery, id).Error
	if err != nil {
		return database.ProcessGormSQLErrorf(ctx, err, "the delete query failed")
	}

	return nil
}
