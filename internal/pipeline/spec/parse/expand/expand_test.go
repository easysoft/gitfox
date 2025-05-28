// Copyright 2022 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package expand

// import (
// 	"encoding/json"
// 	"os"
// 	"testing"

// 	"github.com/easysoft/gitfox/internal/pipeline/spec/parse"
// 	"github.com/easysoft/gitfox/internal/pipeline/spec/parse/normalize"
// )

// func TestExpand(t *testing.T) {
// 	config, err := parse.ParseFile("../../../../samples/matrix.yaml")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	Expand(config)
// 	normalize.Normalize(config)
// 	json.NewEncoder(os.Stdout).Encode(config)
// 	t.Fail()
// }
