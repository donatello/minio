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
	"errors"
	"fmt"
	"strings"

	etcd "go.etcd.io/etcd/client/v3"
)

var errEtcdUnreachable = errors.New("etcd is unreachable, please check your endpoints")

func etcdErrToErr(err error, etcdEndpoints []string) error {
	if err == nil {
		return nil
	}
	switch err {
	case context.DeadlineExceeded:
		return fmt.Errorf("%w %s", errEtcdUnreachable, etcdEndpoints)
	default:
		return fmt.Errorf("unexpected error %w from etcd, please check your endpoints %s", err, etcdEndpoints)
	}
}

func saveKeyEtcdWithTTL(ctx context.Context, client *etcd.Client, key string, data []byte, ttl int64) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
	defer cancel()
	lease, err := client.Grant(timeoutCtx, ttl)
	if err != nil {
		return etcdErrToErr(err, client.Endpoints())
	}
	_, err = client.Put(timeoutCtx, key, string(data), etcd.WithLease(lease.ID))
	return etcdErrToErr(err, client.Endpoints())
}

func saveKeyEtcd(ctx context.Context, client *etcd.Client, key string, data []byte, opts ...options) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
	defer cancel()
	if len(opts) > 0 {
		return saveKeyEtcdWithTTL(ctx, client, key, data, opts[0].ttl)
	}
	_, err := client.Put(timeoutCtx, key, string(data))
	return etcdErrToErr(err, client.Endpoints())
}

func deleteKeyEtcd(ctx context.Context, client *etcd.Client, key string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
	defer cancel()

	_, err := client.Delete(timeoutCtx, key)
	return etcdErrToErr(err, client.Endpoints())
}

func readKeyEtcd(ctx context.Context, client *etcd.Client, key string) ([]byte, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
	defer cancel()
	resp, err := client.Get(timeoutCtx, key)
	if err != nil {
		return nil, etcdErrToErr(err, client.Endpoints())
	}
	if resp.Count == 0 {
		return nil, errConfigNotFound
	}
	for _, ev := range resp.Kvs {
		if string(ev.Key) == key {
			return ev.Value, nil
		}
	}
	return nil, errConfigNotFound
}

// readKeyWithRevEtcd - same as readKeyEtcd, but returns the mod_revision of the
// key if it exists, or the current revision on etcd.
func readKeyWithRevEtcd(ctx context.Context, client *etcd.Client, key string) ([]byte, int64, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
	defer cancel()
	resp, err := client.Get(timeoutCtx, key)
	if err != nil {
		return nil, 0, etcdErrToErr(err, client.Endpoints())
	}
	if resp.Count == 0 {
		return nil, resp.Header.GetRevision(), errConfigNotFound
	}
	for _, ev := range resp.Kvs {
		if string(ev.Key) == key {
			return ev.Value, ev.ModRevision, nil
		}
	}
	return nil, 0, errConfigNotFound
}

type kvPair struct {
	Key   string
	Value []byte
}

// The result keys are stripped of the prefix given here.
func readKeysWithPrefixEtcd(ctx context.Context, client *etcd.Client, pfx string) ([]kvPair, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
	defer cancel()
	resp, err := client.Get(timeoutCtx, pfx, etcd.WithPrefix())
	if err != nil {
		return nil, etcdErrToErr(err, client.Endpoints())
	}
	rs := make([]kvPair, resp.Count)
	for i, ev := range resp.Kvs {
		rs[i].Key, rs[i].Value = strings.TrimPrefix(string(ev.Key), pfx), ev.Value
	}
	return rs, nil
}

func doEtcd(ctx context.Context, client *etcd.Client, op etcd.Op) (etcd.OpResponse, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
	defer cancel()
	resp, err := client.Do(timeoutCtx, op)
	return resp, etcdErrToErr(err, client.Endpoints())
}
