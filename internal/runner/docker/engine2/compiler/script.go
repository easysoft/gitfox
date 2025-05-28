// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	schema "github.com/easysoft/gitfox/internal/pipeline/spec"
	"github.com/easysoft/gitfox/internal/runner/docker/engine/compiler/shell"
	"github.com/easysoft/gitfox/internal/runner/docker/engine/compiler/shell/powershell"
	"github.com/easysoft/gitfox/internal/runner/docker/engine2/engine"
)

// helper function configures the pipeline script for the
// target operating system.
func setupScript(dst *engine.Step, script, os string) {
	if script != "" {
		switch os {
		case "windows":
			setupScriptWindows(dst, script)
		default:
			setupScriptPosix(dst, script)
		}
	}
}

func setupScripts(dst *engine.Step, scripts schema.Stringorslice, os string) {
	if len(scripts) != 0 {
		switch os {
		case "windows":
			setupScriptWindows(dst, scripts...)
		default:
			setupScriptPosix(dst, scripts...)
		}
	}
}

// helper function configures the pipeline script for the
// windows operating system.
func setupScriptWindows(dst *engine.Step, commands ...string) {
	dst.Entrypoint = []string{"powershell", "-noprofile", "-noninteractive", "-command"}
	dst.Command = []string{"echo $Env:GITFOX_SCRIPT | iex"}
	dst.Envs["GITFOX_SCRIPT"] = powershell.Script(commands)
	dst.Envs["SHELL"] = "powershell.exe"
}

// helper function configures the pipeline script for the
// linux operating system.
func setupScriptPosix(dst *engine.Step, commands ...string) {
	dst.Entrypoint = []string{"/bin/sh", "-c"}
	dst.Command = []string{`echo "$GITFOX_SCRIPT" | /bin/sh`}
	dst.Envs["GITFOX_SCRIPT"] = shell.Script(commands)
}
