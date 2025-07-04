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

package settings

type Key string

var (
	// KeySecretScanningEnabled [bool] enables secret scanning if set to true.
	KeySecretScanningEnabled     Key = "secret_scanning_enabled"
	DefaultSecretScanningEnabled     = false
	KeyAIReviewEnabled           Key = "ai_review_enabled"
	DefaultAIReviewEnabled           = false
	KeyFileSizeLimit             Key = "file_size_limit"
	DefaultFileSizeLimit             = int64(5e+8) // 500MB

	// ContainerReadOnly [bool] disable write operates, used for gc
	ContainerReadOnly Key = "container_read_only"

	KeyInstallID     Key = "install_id"
	DefaultInstallID     = string("")
)
