// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package store

import (
	"context"
	"io"
)

// LogStore provides an interface for the persistent log store backend
type LogStore interface {
	// Find returns a log stream from the datastore.
	Find(ctx context.Context, stepID int64) (io.ReadCloser, error)

	// Create writes copies the log stream from Reader r to the datastore.
	Create(ctx context.Context, stepID int64, r io.Reader) error

	// Update copies the log stream from Reader r to the datastore.
	Update(ctx context.Context, stepID int64, r io.Reader) error

	// Delete purges the log stream from the datastore.
	Delete(ctx context.Context, stepID int64) error
}
