// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package extend

import (
	"bytes"
	"fmt"
)

const customScript = `
test -d $(dirname "$GITFOX_CUSTOM_ENV") || mkdir -p $(dirname "$GITFOX_CUSTOM_ENV")
if [ -f "${GITFOX_CUSTOM_ENV}" ];then
	. "${GITFOX_CUSTOM_ENV}"
fi
`

func IncludeShell(buf *bytes.Buffer) {
	fmt.Fprintf(buf, customScript)
}
