// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package database

import (
	"errors"
	"fmt"

	"github.com/easysoft/gitfox/store/database"
)

func buildDSN(dbConfig database.Config) (string, error) {
	if dbConfig.Host == "" || dbConfig.Port == 0 {
		return "", errors.New("db host or port is required")
	}
	if dbConfig.Username == "" || dbConfig.Password == "" {
		return "", errors.New("db username or password is required")
	}
	if dbConfig.DBName == "" {
		return "", errors.New("dbname is required")
	}

	var datasource string
	if dbConfig.Driver == "mysql" {
		datasource = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DBName)
		if dbConfig.Options != "" {
			datasource += "?" + dbConfig.Options
		}
	} else if dbConfig.Driver == "postgres" {
		datasource = fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
			dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DBName)
	} else {
		return "", errors.New("unsupported driver")
	}

	return datasource, nil
}
