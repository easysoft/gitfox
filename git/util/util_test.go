// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package util

import "testing"

func TestCheckGitVersionAtLeast(t *testing.T) {
	type args struct {
		atLeast string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestCheckGitVersionAtLeast229",
			args: args{
				atLeast: "2.29",
			},
			wantErr: false,
		},
		{
			name: "TestCheckGitVersionAtLeast329",
			args: args{
				atLeast: "3.29",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckGitVersionAtLeast(tt.args.atLeast); (err != nil) != tt.wantErr {
				t.Errorf("CheckGitVersionAtLeast() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
