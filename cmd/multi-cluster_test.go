// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	etcdcfg "github.com/minio/minio/internal/config/etcd"
)

// Run like:
//
// MINIO_ETCD_ENDPOINTS=http://localhost:2379 MINIO_ETCD_MCC_ENABLED=true go test -v -run TestMCCOpLock

func newMultiCluster(t *testing.T) *MultiCluster {
	cfg, err := etcdcfg.LookupConfig(nil, nil)
	if err != nil {
		t.Fatalf("etcd config load error: %v", err)
	}
	if cfg.Enabled == false {
		// Test is not enabled
		return nil
	}

	client, err := etcdcfg.New(cfg)
	if err != nil {
		t.Fatalf("etcd client init error: %v", err)
	}

	return &MultiCluster{client}
}

// TestMCCOpLock - test lock sanity; run 100 go routines and increment a counter
// in etcd.
func TestMCCOpLock(t *testing.T) {
	mcs := newMultiCluster(t)
	if mcs == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := "mytestKey"
	val := 0
	valf := func(val int) []byte {
		return []byte(fmt.Sprintf("%d", val))
	}
	err := saveKeyEtcd(ctx, mcs.client, key, valf(val))
	if err != nil {
		t.Fatal(err)
	}

	incrementor := func(n int, mcs *MultiCluster) {
		l, err := mcs.NewOpLock(ctx)
		if err != nil {
			t.Fatal(n, err)
		}
		defer l.Unlock()

		c := mcs.client
		val, err := readKeyEtcd(ctx, c, key)
		if err != nil {
			t.Fatal(n, err)
		}
		i, err := strconv.Atoi(string(val))
		if err != nil {
			t.Fatal(n, err)
		}

		i++
		err = saveKeyEtcd(ctx, c, key, valf(i))
		if err != nil {
			t.Fatal(n, err)
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int, mcs *MultiCluster) {
			defer wg.Done()
			incrementor(i, mcs)
		}(i, mcs)
	}
	wg.Wait()

	{
		val, err := readKeyEtcd(ctx, mcs.client, key)
		if err != nil {
			t.Fatal(err)
		}
		i, err := strconv.Atoi(string(val))
		if err != nil {
			t.Fatal(err)
		}
		if i != 100 {
			t.Errorf("Expected key to be 100, but got: %d", i)
		}
	}
}
