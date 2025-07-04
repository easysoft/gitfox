// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package factory

import (
	"context"
	"fmt"

	storagedriver "github.com/easysoft/gitfox/pkg/storage/driver"
)

// driverFactories stores an internal mapping between storage driver names and their respective
// factories
var driverFactories = make(map[string]StorageDriverFactory)

// StorageDriverFactory is a factory interface for creating storagedriver.StorageDriver interfaces
// Storage drivers should call Register() with a factory to make the driver available by name.
// Individual StorageDriver implementations generally register with the factory via the Register
// func (below) in their init() funcs, and as such they should be imported anonymously before use.
// See below for an example of how to register and get a StorageDriver for S3
//
//	import _ "github.com/distribution/distribution/v3/registry/storage/driver/s3-aws"
//	s3Driver, err = factory.Create("s3", storageParams)
//	// assuming no error, s3Driver is the StorageDriver that communicates with S3 according to storageParams
type StorageDriverFactory interface {
	// Create returns a new storagedriver.StorageDriver with the given parameters
	// Parameters will vary by driver and may be ignored
	// Each parameter key must only consist of lowercase letters and numbers
	Create(ctx context.Context, parameters map[string]interface{}) (storagedriver.StorageDriver, error)
}

// Register makes a storage driver available by the provided name.
// If Register is called twice with the same name or if driver factory is nil, it panics.
// Additionally, it is not concurrency safe. Most Storage Drivers call this function
// in their init() functions. See the documentation for StorageDriverFactory for more.
func Register(name string, factory StorageDriverFactory) {
	if factory == nil {
		panic("Must not provide nil StorageDriverFactory")
	}
	_, registered := driverFactories[name]
	if registered {
		panic(fmt.Sprintf("StorageDriverFactory named %s already registered", name))
	}

	driverFactories[name] = factory
}

// Create a new storagedriver.StorageDriver with the given name and
// parameters. To use a driver, the StorageDriverFactory must first be
// registered with the given name. If no drivers are found, an
// InvalidStorageDriverError is returned
func Create(ctx context.Context, name string, parameters map[string]interface{}) (storagedriver.StorageDriver, error) {
	driverFactory, ok := driverFactories[name]
	if !ok {
		return nil, InvalidStorageDriverError{name}
	}
	return driverFactory.Create(ctx, parameters)
}

// InvalidStorageDriverError records an attempt to construct an unregistered storage driver
type InvalidStorageDriverError struct {
	Name string
}

func (err InvalidStorageDriverError) Error() string {
	return fmt.Sprintf("StorageDriver not registered: %s", err.Name)
}
