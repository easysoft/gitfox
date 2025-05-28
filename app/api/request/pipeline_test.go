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

package request

import (
	"net/url"
	"testing"

	"github.com/golang-module/carbon/v2"
)

func TestParseQueryTimeUnixFromRequest(t *testing.T) {
	t1 := "2014-02-17T09:24:18Z"
	t1unix := carbon.Parse(t1).StdTime().Unix()
	t2 := "2014-02-17T09%3A24%3A18Z"
	t2unix := carbon.Parse(t2).StdTime().Unix()
	t3 := "2014-02-17 09:24:18"
	t3unix := carbon.Parse(t3).StdTime().Unix()
	t4 := "2014-02-17T09%3A24%3A18Z"
	t4, _ = url.QueryUnescape(t4)
	t4unix := carbon.Parse(t4).StdTime().Unix()
	t.Logf("t1unix: %v, t2unix: %v, t3unix: %v, t4unix: %v",
		t1unix, t2unix, t3unix, t4unix)
}
