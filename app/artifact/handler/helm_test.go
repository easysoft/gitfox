// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package handler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelmChartName(t *testing.T) {
	correctNames := [][3]string{
		{"gitlab-1.2.3.tgz", "gitlab", "1.2.3"},
		{"gitlab-1.2.3-beta1.tgz", "gitlab", "1.2.3-beta1"},
		{"gitlab-cron-0.0.1-rc1.tgz", "gitlab-cron", "0.0.1-rc1"},
	}

	for _, test := range correctNames {
		chartName, chartVersion, err := parseHelmPath(test[0])
		require.NoError(t, err)
		require.Equal(t, test[1], chartName)
		require.Equal(t, test[2], chartVersion)
	}
}
