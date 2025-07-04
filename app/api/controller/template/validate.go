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

package template

import (
	"fmt"

	"github.com/easysoft/gitfox/internal/pipeline/spec/parse"
	"github.com/easysoft/gitfox/types/check"
	"github.com/easysoft/gitfox/types/enum"
)

// parseResolverType parses and validates the input yaml. It returns back the parsed
// template type.
func parseResolverType(data string) (enum.ResolverType, error) {
	config, err := parse.ParseString(data)
	if err != nil {
		return "", check.NewValidationError(fmt.Sprintf("could not parse template data: %s", err))
	}
	resolverTypeEnum, err := enum.ParseResolverType(config.Type)
	if err != nil {
		return "", check.NewValidationError(fmt.Sprintf("could not parse template type: %s", config.Type))
	}
	return resolverTypeEnum, nil
}
