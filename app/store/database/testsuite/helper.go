// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package testsuite

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func readFieldValue(obj interface{}, fieldName string) (interface{}, error) {
	objValue := reflect.ValueOf(obj)

	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}

	if objValue.Kind() != reflect.Struct {
		return nil, errors.New("obj must be a struct or pointer to struct")
	}

	fieldValue := objValue.FieldByName(fieldName)

	if !fieldValue.IsValid() {
		return nil, fmt.Errorf("field %s not found", fieldName)
	}

	return fieldValue.Interface(), nil
}

func EqualFieldValue(t *testing.T, expectedValue interface{}, obj interface{}, fieldName string, msgAndArgs ...interface{}) {
	t.Helper()
	val, err := readFieldValue(obj, fieldName)
	require.NoError(t, err, msgAndArgs...)
	require.EqualValues(t, expectedValue, val, msgAndArgs...)
}

func EqualStructFieldValue(t *testing.T, expect interface{}, actual interface{}, fieldName string, msgAndArgs ...interface{}) {
	t.Helper()
	valExpect, err := readFieldValue(expect, fieldName)
	require.NoError(t, err, msgAndArgs...)
	valActual, err := readFieldValue(actual, fieldName)
	require.NoError(t, err, msgAndArgs...)
	require.EqualValues(t, valExpect, valActual, msgAndArgs...)
}

func NotEqualStructFieldValue(t *testing.T, expect interface{}, actual interface{}, fieldName string, msgAndArgs ...interface{}) {
	t.Helper()
	valExpect, err := readFieldValue(expect, fieldName)
	require.NoError(t, err, msgAndArgs...)
	valActual, err := readFieldValue(actual, fieldName)
	require.NoError(t, err, msgAndArgs...)
	require.NotEqualValues(t, valExpect, valActual, msgAndArgs...)
}
