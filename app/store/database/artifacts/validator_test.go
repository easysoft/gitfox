// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifacts

import (
	"testing"

	"github.com/easysoft/gitfox/types"

	"github.com/stretchr/testify/require"
)

func TestValidatorEmpty(t *testing.T) {
	type EmptyString string

	tests := []interface{}{
		"", EmptyString(""), nil,
		0, int8(0), int16(0), int32(0), int64(0),
		uint(0), uint8(0), uint16(0), uint32(0), uint64(0),
		0.0, float32(0.0),
		[]string{}, map[string]interface{}{},
	}

	for id, test := range tests {
		err := validatorEmpty(test)
		require.ErrorIs(t, err, types.ErrArgsValueEmpty, "loop index: %d", id)
	}

	tests2 := []interface{}{
		"a", EmptyString("b"), &struct{}{},
		1, int8(2), int16(3), int32(4), int64(5), -1,
		uint(6), uint8(7), uint16(8), uint32(9), uint64(10),
		0.1, float32(0.2), -0.1,
		[]string{""}, map[string]interface{}{"k": "v"},
	}

	for id, test := range tests2 {
		err := validatorEmpty(test)
		require.NoError(t, err, "loop index: %d", id)
	}
}
