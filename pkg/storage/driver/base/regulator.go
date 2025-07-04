// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package base

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"sync"

	storagedriver "github.com/easysoft/gitfox/pkg/storage/driver"
)

type regulator struct {
	storagedriver.StorageDriver
	*sync.Cond

	available uint64
}

// GetLimitFromParameter takes an interface type as decoded from the YAML
// configuration and returns a uint64 representing the maximum number of
// concurrent calls given a minimum limit and default.
//
// If the parameter supplied is of an invalid type this returns an error.
func GetLimitFromParameter(param interface{}, min, def uint64) (uint64, error) {
	limit := def

	switch v := param.(type) {
	case string:
		var err error
		if limit, err = strconv.ParseUint(v, 0, 64); err != nil {
			return limit, fmt.Errorf("parameter must be an integer, '%v' invalid", param)
		}
	case uint64:
		limit = v
	case int, int32, int64:
		val := reflect.ValueOf(v).Convert(reflect.TypeOf(param)).Int()
		// if param is negative casting to uint64 will wrap around and
		// give you the hugest thread limit ever. Let's be sensible, here
		if val > 0 {
			limit = uint64(val)
		} else {
			limit = min
		}
	case uint, uint32:
		limit = reflect.ValueOf(v).Convert(reflect.TypeOf(param)).Uint()
	case nil:
		// use the default
	default:
		return 0, fmt.Errorf("invalid value '%#v'", param)
	}

	if limit < min {
		return min, nil
	}

	return limit, nil
}

// NewRegulator wraps the given driver and is used to regulate concurrent calls
// to the given storage driver to a maximum of the given limit. This is useful
// for storage drivers that would otherwise create an unbounded number of OS
// threads if allowed to be called unregulated.
func NewRegulator(driver storagedriver.StorageDriver, limit uint64) storagedriver.StorageDriver {
	return &regulator{
		StorageDriver: driver,
		Cond:          sync.NewCond(&sync.Mutex{}),
		available:     limit,
	}
}

func (r *regulator) enter() {
	r.L.Lock()
	for r.available == 0 {
		r.Wait()
	}
	r.available--
	r.L.Unlock()
}

func (r *regulator) exit() {
	r.L.Lock()
	r.Signal()
	r.available++
	r.L.Unlock()
}

// Name returns the human-readable "name" of the driver, useful in error
// messages and logging. By convention, this will just be the registration
// name, but drivers may provide other information here.
func (r *regulator) Name() string {
	r.enter()
	defer r.exit()

	return r.StorageDriver.Name()
}

// GetContent retrieves the content stored at "path" as a []byte.
// This should primarily be used for small objects.
func (r *regulator) GetContent(ctx context.Context, path string) ([]byte, error) {
	r.enter()
	defer r.exit()

	return r.StorageDriver.GetContent(ctx, path)
}

// PutContent stores the []byte content at a location designated by "path".
// This should primarily be used for small objects.
func (r *regulator) PutContent(ctx context.Context, path string, content []byte) error {
	r.enter()
	defer r.exit()

	return r.StorageDriver.PutContent(ctx, path, content)
}

// Reader retrieves an io.ReadCloser for the content stored at "path"
// with a given byte offset.
// May be used to resume reading a stream by providing a nonzero offset.
func (r *regulator) Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	r.enter()
	defer r.exit()

	return r.StorageDriver.Reader(ctx, path, offset)
}

// Writer stores the contents of the provided io.ReadCloser at a
// location designated by the given path.
// May be used to resume writing a stream by providing a nonzero offset.
// The offset must be no larger than the CurrentSize for this path.
func (r *regulator) Writer(ctx context.Context, path string, append bool) (storagedriver.FileWriter, error) {
	r.enter()
	defer r.exit()

	return r.StorageDriver.Writer(ctx, path, append)
}

// Stat retrieves the FileInfo for the given path, including the current
// size in bytes and the creation time.
func (r *regulator) Stat(ctx context.Context, path string) (storagedriver.FileInfo, error) {
	r.enter()
	defer r.exit()

	return r.StorageDriver.Stat(ctx, path)
}

// List returns a list of the objects that are direct descendants of the
// given path.
func (r *regulator) List(ctx context.Context, path string) ([]string, error) {
	r.enter()
	defer r.exit()

	return r.StorageDriver.List(ctx, path)
}

// Move moves an object stored at sourcePath to destPath, removing the
// original object.
// Note: This may be no more efficient than a copy followed by a delete for
// many implementations.
func (r *regulator) Move(ctx context.Context, sourcePath string, destPath string) error {
	r.enter()
	defer r.exit()

	return r.StorageDriver.Move(ctx, sourcePath, destPath)
}

// Delete recursively deletes all objects stored at "path" and its subpaths.
func (r *regulator) Delete(ctx context.Context, path string) error {
	r.enter()
	defer r.exit()

	return r.StorageDriver.Delete(ctx, path)
}

// RedirectURL returns a URL which may be used to retrieve the content stored at
// the given path.
func (r *regulator) RedirectURL(req *http.Request, path string) (string, error) {
	r.enter()
	defer r.exit()

	return r.StorageDriver.RedirectURL(req, path)
}
