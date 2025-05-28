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

package main

import (
	"github.com/easysoft/gitfox/app/api/openapi"
	"github.com/easysoft/gitfox/cli"
	"github.com/easysoft/gitfox/cli/operations/account"
	"github.com/easysoft/gitfox/cli/operations/hooks"
	"github.com/easysoft/gitfox/cli/operations/migrate"
	"github.com/easysoft/gitfox/cli/operations/server"
	"github.com/easysoft/gitfox/cli/operations/swagger"
	"github.com/easysoft/gitfox/cli/operations/user"
	"github.com/easysoft/gitfox/cli/operations/users"
	"github.com/easysoft/gitfox/pkg/util/common"

	"github.com/quicklyon/kingpin/v2"
)

const (
	application = "gitfox"
	description = "Gitfox"
)

func main() {
	args := cli.GetArguments()

	app := kingpin.New(application, description)

	migrate.Register(app)
	server.Register(app, initSystem)

	user.Register(app)
	users.Register(app)

	account.RegisterLogin(app)
	account.RegisterRegister(app)
	account.RegisterLogout(app)

	hooks.Register(app)

	swagger.Register(app, openapi.NewOpenAPIService())

	kingpin.Version(common.GetVersionWithHash())
	kingpin.MustParse(app.Parse(args))
}
