// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifacts

import (
	"reflect"

	"github.com/easysoft/gitfox/types"
)

func validatorEmpty(args ...any) error {
	emptyArgs := make([]any, 0)
	for _, arg := range args {
		if isEmpty(arg) {
			emptyArgs = append(emptyArgs, arg)
		}
	}

	if len(emptyArgs) > 0 {
		return types.ErrArgsValueEmpty
	}

	return nil
}

func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch reflect.ValueOf(value).Kind() {
	case reflect.String:
		return reflect.ValueOf(value).String() == ""
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return reflect.ValueOf(value).Int() == 0
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return reflect.ValueOf(value).Uint() == 0
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(value).Float() == 0.0
	case reflect.Slice, reflect.Array, reflect.Map:
		return reflect.ValueOf(value).Len() == 0
	default:
		return false
	}
}
