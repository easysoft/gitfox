// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package check

import "regexp"

const (
	standardNameRegex = "^[a-z][a-z0-9-_]*[a-z0-9]$"
)

// StandardName checks the provided string and returns an error if it isn't valid.
func StandardName(name string) error {
	l := len(name)
	if l < minIdentifierLength || l > MaxIdentifierLength {
		return ErrIdentifierLength
	}

	if ok, _ := regexp.Match(standardNameRegex, []byte(name)); !ok {
		return ErrIdentifierRegex
	}

	return nil
}
