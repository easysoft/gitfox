// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package util

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/easysoft/gitfox/git/command"
	"github.com/easysoft/gitfox/git/enum"

	"github.com/hashicorp/go-version"
)

var (
	gitVersion     *version.Version
	DefaultContext context.Context
)

// loadGitVersion returns current Git version from shell. Internal usage only.
func loadGitVersion() (*version.Version, error) {
	// doesn't need RWMutex because it's executed by Init()
	if gitVersion != nil {
		return gitVersion, nil
	}
	cmd := command.New("version")
	output := &bytes.Buffer{}
	if err := cmd.Run(context.Background(), command.WithStdout(output)); err != nil {
		return nil, err
	}
	stdout := output.String()
	fields := strings.Fields(stdout)
	if len(fields) < 3 {
		return nil, fmt.Errorf("invalid git version output: %s", stdout)
	}
	var versionString string
	// Handle special case on Windows.
	i := strings.Index(fields[2], "windows")
	if i >= 1 {
		versionString = fields[2][:i-1]
	} else {
		versionString = fields[2]
	}
	var err error
	gitVersion, err = version.NewVersion(versionString)
	return gitVersion, err
}

// CheckGitVersionAtLeast check git version is at least the constraint version
func CheckGitVersionAtLeast(atLeast string) error {
	if _, err := loadGitVersion(); err != nil {
		return err
	}
	atLeastVersion, err := version.NewVersion(atLeast)
	if err != nil {
		return err
	}
	if gitVersion.Compare(atLeastVersion) < 0 {
		return fmt.Errorf("installed git binary version %s is not at least %s", gitVersion.Original(), atLeast)
	}
	return nil
}

// SupportProcReceive returns true if the installed git binary supports proc-receive
func SupportProcReceive() bool {
	return CheckGitVersionAtLeast("2.29") != nil
}

func loadGitfoxBinPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return enum.DefaultGitfoxBinPath
	}
	if _, err := os.Stat(exePath); err != nil {
		return enum.DefaultGitfoxBinPath
	}
	return exePath
}

const windowsSharingViolationError syscall.Errno = 32

// Remove removes the named file or (empty) directory with at most 5 attempts.
func Remove(name string) error {
	var err error
	for i := 0; i < 5; i++ {
		err = os.Remove(name)
		if err == nil {
			break
		}
		unwrapped := err.(*os.PathError).Err
		if unwrapped == syscall.EBUSY || unwrapped == syscall.ENOTEMPTY || unwrapped == syscall.EPERM || unwrapped == syscall.EMFILE || unwrapped == syscall.ENFILE {
			// try again
			<-time.After(100 * time.Millisecond)
			continue
		}

		if unwrapped == windowsSharingViolationError && runtime.GOOS == "windows" {
			// try again
			<-time.After(100 * time.Millisecond)
			continue
		}

		if unwrapped == syscall.ENOENT {
			// it's already gone
			return nil
		}
	}
	return err
}

func EnsureExecutable(filename string) error {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return err
	}
	if (fileInfo.Mode() & 0100) > 0 {
		return nil
	}
	mode := fileInfo.Mode() | 0100
	return os.Chmod(filename, mode)
}
