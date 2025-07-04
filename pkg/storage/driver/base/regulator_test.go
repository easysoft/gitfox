// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package base

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRegulatorEnterExit(t *testing.T) {
	const limit = 500

	r := NewRegulator(nil, limit).(*regulator)

	for try := 0; try < 50; try++ {
		run := make(chan struct{})

		var firstGroupReady sync.WaitGroup
		var firstGroupDone sync.WaitGroup
		firstGroupReady.Add(limit)
		firstGroupDone.Add(limit)
		for i := 0; i < limit; i++ {
			go func() {
				r.enter()
				firstGroupReady.Done()
				<-run
				r.exit()
				firstGroupDone.Done()
			}()
		}
		firstGroupReady.Wait()

		// now we exhausted all the limit, let's run a little bit more
		var secondGroupReady sync.WaitGroup
		var secondGroupDone sync.WaitGroup
		for i := 0; i < 50; i++ {
			secondGroupReady.Add(1)
			secondGroupDone.Add(1)
			go func() {
				secondGroupReady.Done()
				r.enter()
				r.exit()
				secondGroupDone.Done()
			}()
		}
		secondGroupReady.Wait()

		// allow the first group to return resources
		close(run)

		done := make(chan struct{})
		go func() {
			secondGroupDone.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("some r.enter() are still locked")
		}

		firstGroupDone.Wait()

		if r.available != limit {
			t.Fatalf("r.available: got %d, want %d", r.available, limit)
		}
	}
}

func TestGetLimitFromParameter(t *testing.T) {
	tests := []struct {
		Input    interface{}
		Expected uint64
		Min      uint64
		Default  uint64
		Err      error
	}{
		{"foo", 0, 5, 5, fmt.Errorf("parameter must be an integer, 'foo' invalid")},
		{"50", 50, 5, 5, nil},
		{"5", 25, 25, 50, nil}, // lower than Min returns Min
		{nil, 50, 25, 50, nil}, // nil returns default
		{812, 812, 25, 50, nil},
	}

	for _, item := range tests {
		t.Run(fmt.Sprint(item.Input), func(t *testing.T) {
			actual, err := GetLimitFromParameter(item.Input, item.Min, item.Default)

			if err != nil && item.Err != nil && err.Error() != item.Err.Error() {
				t.Fatalf("GetLimitFromParameter error, expected %#v got %#v", item.Err, err)
			}

			if actual != item.Expected {
				t.Fatalf("GetLimitFromParameter result error, expected %d got %d", item.Expected, actual)
			}
		})
	}
}
