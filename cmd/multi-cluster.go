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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/minio/madmin-go"
	"github.com/minio/minio-go/v7/pkg/set"
	"github.com/minio/minio/internal/auth"
	"github.com/minio/minio/internal/logger"
	iampolicy "github.com/minio/pkg/iam/policy"
	etcd "go.etcd.io/etcd/client/v3"
	etcdcc "go.etcd.io/etcd/client/v3/concurrency"
)

var (
	errDuplicateBucket = errors.New("bucket with this name already exists in the multi-cluster")
	errAlreadyJoined   = errors.New("cluster has already joined")
)

const (
	mccPrefix      = minioConfigPrefix + "/mcc"
	clustersPrefix = mccPrefix + "/deployments/"
	bucketsPrefix  = mccPrefix + "/buckets/"
	lockPath       = mccPrefix + "/lock"
)

type mccEntry int

const (
	mccEntryCluster mccEntry = iota
	mccEntryBucket
	mccEntryPolicy
	mccEntryServiceAccount
	mccEntrySTSAccount
	mccEntryPolicyDBServiceAccount
	mccEntryPolicyDBSTSUser
	mccEntryPolicyDBUser
	mccEntryPolicyDBGroup
)

func getMCCKeyPath(name string, entryType mccEntry) string {
	switch entryType {
	case mccEntryCluster:
		return clustersPrefix + name
	case mccEntryBucket:
		return bucketsPrefix + name
	case mccEntryPolicy:
		return iamConfigPoliciesPrefix + name
	case mccEntryServiceAccount:
		return iamConfigServiceAccountsPrefix + name
	case mccEntrySTSAccount:
		return iamConfigSTSPrefix + name
	case mccEntryPolicyDBServiceAccount:
		return iamConfigPolicyDBServiceAccountsPrefix + name
	case mccEntryPolicyDBSTSUser:
		return iamConfigPolicyDBSTSUsersPrefix + name
	case mccEntryPolicyDBUser:
		return iamConfigPolicyDBUsersPrefix + name
	case mccEntryPolicyDBGroup:
		return iamConfigPolicyDBGroupsPrefix + name
	}
	return ""
}

func getMCCOpPut(name string, entryType mccEntry, value interface{}, opts ...etcd.OpOption) (etcd.Op, error) {
	key := getMCCKeyPath(name, entryType)
	if key == "" {
		return etcd.Op{}, errors.New("invalid mcc entry type")
	}
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return etcd.Op{}, err
	}
	return etcd.OpPut(key, string(valueBytes), opts...), nil
}

// NewMultiCluster - create a new MultiCluster
func NewMultiCluster(client *etcd.Client) (*MultiCluster, error) {
	return &MultiCluster{
		client: client,
	}, nil
}

// MultiCluster contains info and methods for multi-cluster features.
type MultiCluster struct {
	client *etcd.Client

	// joinedValue == 0 iff not joined in a multi-cluster (after
	// initialization).
	joinedValue int64

	// context saved from Init() call.
	initContext context.Context
}

const (
	hasNotJoinedValue = 0
	hasJoinedValue    = 1
)

// IsJoined - check if the cluster has joined the multi-cluster.
func (mcs *MultiCluster) IsJoined() bool {
	if mcs == nil {
		return false
	}
	v := atomic.LoadInt64(&mcs.joinedValue)
	return v == hasJoinedValue
}

// Init - initialize the multi-cluster.
func (mcs *MultiCluster) Init(ctx context.Context) {
	if mcs == nil {
		return
	}

	fmt.Println("MultiCluster Init() started:", globalDeploymentID)
	mcs.initContext = ctx
	go mcs.watchMCCEvents(ctx)
}

type mccOpLock struct {
	mu      sync.Locker
	session *etcdcc.Session
}

func (m mccOpLock) Lock() {
	m.mu.Lock()
}

func (m mccOpLock) Unlock() {
	m.mu.Unlock()
	m.session.Close()
}

// NewOpLock - returns an mcc mutex starting in locked mode.
func (mcs *MultiCluster) NewOpLock(ctx context.Context) (l sync.Locker, err error) {
	session, err := etcdcc.NewSession(mcs.client, etcdcc.WithContext(ctx))
	if err != nil {
		return l, err
	}

	l = mccOpLock{
		mu:      etcdcc.NewLocker(session, lockPath),
		session: session,
	}
	l.Lock()
	return l, nil

}

const (
	validationFailedMessage = "Failed due to validation error(s)"
	successMessage          = "Success!"
)

// Join - joins a multi-cluster
func (mcs *MultiCluster) Join(ctx context.Context) (*madmin.ClusterLinkResult, error) {
	// Acquire the mcc lock.
	timeoutCtx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
	defer cancel()

	mu, err := mcs.NewOpLock(timeoutCtx)
	if err != nil {
		return nil, err
	}
	defer mu.Unlock()

	// Check if already joined
	ce, _, err := mcs.getClusterRegistration(ctx)
	if err != nil {
		return nil, err
	}
	if ce != nil {
		return nil, errAlreadyJoined
	}

	var res madmin.ClusterLinkResult
	validationFailed := false
	// Validations: buckets, policy names, sts account IDs, service account
	// IDs.
	var bucketsInfo []BucketInfo
	{
		// Buckets validation.
		objAPI := newObjectLayerFn()
		bucketsInfo, err = objAPI.ListBuckets(timeoutCtx)
		if err != nil {
			return nil, err
		}
		kvs, err := readKeysWithPrefixEtcd(timeoutCtx, mcs.client, bucketsPrefix)
		if err != nil {
			return nil, err
		}
		existingBuckets := set.NewStringSet()
		for _, p := range kvs {
			existingBuckets.Add(p.Key)
		}
		fmt.Println(existingBuckets)
		for _, bi := range bucketsInfo {
			if existingBuckets.Contains(bi.Name) {
				res.UsedBucketNames = append(res.UsedBucketNames, bi.Name)
				validationFailed = true
			}
		}
	}

	myPolicyMap := make(map[string]iampolicy.Policy)
	{
		// Policy docs validation.
		err := globalIAMSys.store.loadPolicyDocs(timeoutCtx, myPolicyMap)
		if err != nil {
			return nil, err
		}
		kvs, err := readKeysWithPrefixEtcd(timeoutCtx, mcs.client, iamConfigPoliciesPrefix)
		if err != nil {
			return nil, err
		}
		existingPolicies := set.NewStringSet()
		for _, p := range kvs {
			existingPolicies.Add(p.Key)
		}
		for p := range myPolicyMap {
			if existingPolicies.Contains(p) {
				res.UsedPolicyNames = append(res.UsedPolicyNames, p)
				validationFailed = true
			}
		}
	}

	myServiceAccounts := make(map[string]auth.Credentials)
	{
		// Service account IDs validation.
		err := globalIAMSys.store.loadUsers(timeoutCtx, svcUser, myServiceAccounts)
		if err != nil {
			return nil, err
		}
		kvs, err := readKeysWithPrefixEtcd(timeoutCtx, mcs.client, iamConfigServiceAccountsPrefix)
		if err != nil {
			return nil, err
		}
		existingAccountIDs := set.NewStringSet()
		for _, p := range kvs {
			existingAccountIDs.Add(p.Key)
		}
		for p := range myServiceAccounts {
			if existingAccountIDs.Contains(p) {
				res.UsedServiceAccountIDs = append(res.UsedServiceAccountIDs, p)
				validationFailed = true
			}
		}
	}

	mySTSAccounts := make(map[string]auth.Credentials)
	{
		// Service account IDs validation.
		err := globalIAMSys.store.loadUsers(timeoutCtx, stsUser, mySTSAccounts)
		if err != nil {
			return nil, err
		}
		kvs, err := readKeysWithPrefixEtcd(timeoutCtx, mcs.client, iamConfigSTSPrefix)
		if err != nil {
			return nil, err
		}
		existingAccountIDs := set.NewStringSet()
		for _, p := range kvs {
			existingAccountIDs.Add(p.Key)
		}
		for p := range mySTSAccounts {
			if existingAccountIDs.Contains(p) {
				res.UsedSTSAccountIDs = append(res.UsedSTSAccountIDs, p)
				validationFailed = true
			}
		}
	}

	// Only load sts and group policy mappings to enable LDAP to work. Other
	// policy mappings are for internal IDP and are not relevant for
	// multi-cluster.
	myPolicyMappings := map[string]map[string]MappedPolicy{
		"stsUsers": make(map[string]MappedPolicy),
		"groups":   make(map[string]MappedPolicy),
	}

	{
		err := globalIAMSys.store.loadMappedPolicies(timeoutCtx, regUser, true, myPolicyMappings["groups"])
		if err != nil {
			return nil, err
		}
		kvs, err := readKeysWithPrefixEtcd(timeoutCtx, mcs.client, iamConfigPolicyDBGroupsPrefix)
		if err != nil {
			return nil, err
		}
		existing := set.NewStringSet()
		for _, p := range kvs {
			existing.Add(p.Key)
		}
		for p := range myPolicyMappings["groups"] {
			if existing.Contains(p) {
				res.ExistingGroupPolicyMappings = append(res.ExistingGroupPolicyMappings, p)
				validationFailed = true
			}
		}
	}
	{
		err := globalIAMSys.store.loadMappedPolicies(timeoutCtx, stsUser, false, myPolicyMappings["stsUsers"])
		if err != nil {
			return nil, err
		}
		kvs, err := readKeysWithPrefixEtcd(timeoutCtx, mcs.client, iamConfigPolicyDBSTSUsersPrefix)
		if err != nil {
			return nil, err
		}
		existing := set.NewStringSet()
		for _, p := range kvs {
			existing.Add(p.Key)
		}
		for p := range myPolicyMappings["groups"] {
			if existing.Contains(p) {
				res.ExistingSTSPolicyMappings = append(res.ExistingSTSPolicyMappings, p)
				validationFailed = true
			}
		}
	}

	if validationFailed {
		res.Status = validationFailedMessage
		return &res, nil
	}

	// Validations passed, we are holding the lock, so we can write our
	// entries to mcc/etcd.
	var putOps []etcd.Op

	// Add bucket entries.
	bEntry := bucketEntry{[]string{globalDeploymentID}}
	for _, bi := range bucketsInfo {
		op, err := getMCCOpPut(bi.Name, mccEntryBucket, bEntry)
		if err != nil {
			return nil, err
		}
		putOps = append(putOps, op)
	}

	// Add policy entries
	for pname, pdoc := range myPolicyMap {
		op, err := getMCCOpPut(pname, mccEntryPolicy, pdoc)
		if err != nil {
			return nil, err
		}
		putOps = append(putOps, op)
	}

	// Add service account entries
	for aid, cred := range myServiceAccounts {
		uIdentity := newUserIdentity(cred)
		op, err := getMCCOpPut(aid, mccEntryServiceAccount, uIdentity)
		if err != nil {
			return nil, err
		}
		putOps = append(putOps, op)
	}

	// Add STS account entries
	for aid, cred := range mySTSAccounts {
		uIdentity := newUserIdentity(cred)
		op, err := getMCCOpPut(aid, mccEntrySTSAccount, uIdentity)
		if err != nil {
			return nil, err
		}
		putOps = append(putOps, op)
	}

	// Add group policy mapping entries
	for g, mp := range myPolicyMappings["group"] {
		op, err := getMCCOpPut(g, mccEntryPolicyDBGroup, mp)
		if err != nil {
			return nil, err
		}
		putOps = append(putOps, op)
	}

	// Add sts user policy mapping entries
	for g, mp := range myPolicyMappings["stsUsers"] {
		op, err := getMCCOpPut(g, mccEntryPolicyDBSTSUser, mp)
		if err != nil {
			return nil, err
		}
		putOps = append(putOps, op)
	}

	// Add cluster registration entry
	clusterEntryOp, err := getClusterEntryPutOp()
	if err != nil {
		return nil, err
	}
	putOps = append(putOps, clusterEntryOp)

	kvc := etcd.NewKV(mcs.client)
	txCtx, txCancel := context.WithTimeout(ctx, defaultContextTimeout)
	defer txCancel()
	// The If in the txn always evaluates to false, as we have not yet
	// registered.
	txRes, err := kvc.Txn(txCtx).
		If(etcd.Compare(etcd.ModRevision(string(clusterEntryOp.KeyBytes())), ">", 0)).
		Then().
		Else(putOps...).
		Commit()
	if err != nil {
		return nil, err
	}
	if txRes.Succeeded {
		// NOT EXPECTED TO HAPPEN!
		return nil, errors.New("this should not happen!")
	}

	res.Status = successMessage
	return &res, nil
}

type clusterEntry struct {
	Endpoints []string `json:"endpoints"`
}

func getClusterEntryPutOp() (etcd.Op, error) {
	cEntry := clusterEntry{Endpoints: collectServerURLs()}
	return getMCCOpPut(globalDeploymentID, mccEntryCluster, cEntry)
}

// getClusterRegistration gets the current cluster registration. It also returns
// mod_revision of the entry if found, or the revision of the etcd cluster.
func (mcs *MultiCluster) getClusterRegistration(ctx context.Context) (*clusterEntry, int64, error) {
	val, rev, err := readKeyWithRevEtcd(ctx, mcs.client, getMCCKeyPath(globalDeploymentID, mccEntryCluster))
	if err == errConfigNotFound {
		return nil, rev, nil
	}
	if err != nil {
		return nil, 0, err
	}
	var ce clusterEntry
	err = json.Unmarshal(val, &ce)
	return &ce, rev, err
}

// scanMCC scans MCC for IAM files that are not yet in IAM object store and
// saves them.
func (mcs *MultiCluster) scanMCC(ctx context.Context) {
	if !mcs.IsJoined() {
		return
	}

	resp, err := mcs.client.Get(ctx, iamConfigPrefix+"/", etcd.WithPrefix())
	if err != nil {
		logger.Error("Unable to get IAM prefixed keys: %v", err)
		logger.LogIf(ctx, err)
		return
	}

	for _, kv := range resp.Kvs {
		// TODO: we can skip IAM files that we already have - but we
		// need to handle mutations too!
		err := handleMCCIAMEvent(string(kv.Key), true, kv.Value)
		if err != nil {
			logger.Error("MCC error handling IAM event: key=%s value=%s", string(kv.Key), string(kv.Value))
			// We continue to handle other keys
		}
	}
}

// watchMCCEvents watches mcc for new events - including joining or leaving the
// multi-cluster, and new IAM files added by other clusters.
func (mcs *MultiCluster) watchMCCEvents(ctx context.Context) {
	// At any time mcc is enabled, we are either joined or not. While we are
	// joined in the multi-cluster, watch for events in mcc and save to
	// local storage.

	var cEntry *clusterEntry
	var cluRegKeyRev, iamKeyRev int64
	var err error
	for {
		cEntry, cluRegKeyRev, err = mcs.getClusterRegistration(ctx)
		if err != nil {
			logger.Error("unable to check cluster registration: %v", err)
			time.Sleep(time.Second)
			continue
		}
		break
	}

	if cEntry != nil {
		atomic.StoreInt64(&mcs.joinedValue, hasJoinedValue)
	}

	watcher := etcd.NewWatcher(mcs.client)
	regEventWatchFn := func(rev int64) etcd.WatchChan {
		return watcher.Watch(ctx, getMCCKeyPath(globalDeploymentID, mccEntryCluster), etcd.WithRev(rev))
	}
	iamEventWatchFn := func(rev int64) etcd.WatchChan {
		return watcher.Watch(ctx, iamConfigPrefix+"/", etcd.WithRev(rev), etcd.WithPrefix())
	}
	cluRegWatchCh := regEventWatchFn(cluRegKeyRev)
	var iamWatchCh etcd.WatchChan

	for {
		select {
		case cluRegEvent := <-cluRegWatchCh:
			// Watch for registration key events. If the key is
			// created or modified, the cluster is joined and we
			// sync IAM entries from the server. If not, we stop IAM
			// sycing.
			if cluRegEvent.Canceled {
				// watch failed, we restart it after a delay.
				time.Sleep(time.Second)
				cluRegWatchCh = regEventWatchFn(cluRegKeyRev)
				continue
			}

			for _, event := range cluRegEvent.Events {
				// Update the rev, so that in case the watch is
				// restarted, we only read new events instead of
				// replaying old ones.
				cluRegKeyRev = event.Kv.ModRevision

				// Since we are watching exactly one key, we do
				// not need to examine the event - just check
				// the kind of operation it was.
				if event.IsCreate() || event.IsModify() {
					atomic.StoreInt64(&mcs.joinedValue, hasJoinedValue)
					fmt.Printf("joined!\n")
					if iamWatchCh == nil {
						iamKeyRev = cluRegKeyRev
						iamWatchCh = iamEventWatchFn(iamKeyRev)
					}

					// scan mcc if we joined
					go mcs.scanMCC(ctx)
				} else {
					atomic.StoreInt64(&mcs.joinedValue, hasNotJoinedValue)
					iamWatchCh = nil
				}

			}

		case iamEvent := <-iamWatchCh:
			if iamEvent.Canceled {
				// watch failed, we restart it after a delay.
				time.Sleep(time.Second)
				iamWatchCh = iamEventWatchFn(iamKeyRev)
				continue
			}

			for _, event := range iamEvent.Events {
				// Update the rev, so that in case the watch is
				// restarted, we only read new events instead of
				// replaying old ones.
				iamKeyRev = event.Kv.ModRevision

				err := handleMCCIAMEvent(string(event.Kv.Key), event.IsCreate() || event.IsModify(), event.Kv.Value)
				if err != nil {
					logger.Error("MCC error handling IAM event: %v (key=%s)", err, string(event.Kv.Key))
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

type bucketEntry struct {
	Deployments []string `json:"deployments,omitempty"`
}

func (be *bucketEntry) contains(depID string) bool {
	for _, d := range be.Deployments {
		if d == depID {
			return true
		}
	}
	return false
}

func (be *bucketEntry) append(depID string) {
	be.Deployments = append(be.Deployments, depID)
}

func getBucketEntryEtcd(ctx context.Context, client *etcd.Client, bucket string) (*bucketEntry, error) {
	val, err := readKeyEtcd(ctx, client, getMCCKeyPath(bucket, mccEntryBucket))
	if err == errConfigNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var bEntry bucketEntry
	err = json.Unmarshal(val, &bEntry)
	if err != nil {
		return nil, err
	}

	return &bEntry, nil
}

// DoesBucketExist - check if a bucket is present in the multi-cluster.
func (mcs *MultiCluster) DoesBucketExist(ctx context.Context, bucket string) (bool, error) {
	if !mcs.IsJoined() {
		return false, nil
	}

	bEntry, err := getBucketEntryEtcd(ctx, mcs.client, bucket)
	return bEntry != nil, err
}

// RegisterNewBucket - create an entry in mcc to record the bucket exists in the
// multi-cluster.
func (mcs *MultiCluster) RegisterNewBucket(ctx context.Context, bucket string, isReplicated bool) error {
	if !mcs.IsJoined() {
		return nil
	}

	bEntry, err := getBucketEntryEtcd(ctx, mcs.client, bucket)
	if err != nil {
		return err
	}
	if bEntry != nil {
		if bEntry.contains(globalDeploymentID) {
			// Nothing to do.
			return nil
		}

		if !isReplicated {
			return errDuplicateBucket
		}
	} else {
		bEntry = &bucketEntry{}
		bEntry.append(globalDeploymentID)
	}
	return mcs.putMCC(ctx, bucket, mccEntryBucket, bEntry)
}

// DeregisterBucket - deregisters a bucket from mcc when it is deleted.
func (mcs *MultiCluster) DeregisterBucket(ctx context.Context, bucket string) error {
	if !mcs.IsJoined() {
		return nil
	}

	bEntry, err := getBucketEntryEtcd(ctx, mcs.client, bucket)
	if err != nil {
		return err
	}
	if bEntry == nil {
		// Bucket key is not present - this is not expected to happen.
		return nil
	}

	idx := -1
	for i, d := range bEntry.Deployments {
		if d == globalDeploymentID {
			idx = i
			break
		}
	}
	if idx == -1 {
		// Already not present in mcc.
		return nil
	}
	bEntry.Deployments = append(bEntry.Deployments[0:idx], bEntry.Deployments[idx+1:]...)

	// If the slice becomes empty after removing the
	// globalDeploymentID, we need to delete it.
	if len(bEntry.Deployments) > 0 {
		return mcs.putMCC(ctx, bucket, mccEntryBucket, bEntry)
	}
	return deleteKeyEtcd(ctx, mcs.client, getMCCKeyPath(bucket, mccEntryBucket))
}

// RegisterNewPolicy - register a new policy with mcc.
func (mcs *MultiCluster) RegisterNewPolicy(ctx context.Context, policyName string, policy iampolicy.Policy) error {
	if !mcs.IsJoined() {
		return nil
	}
	return mcs.putMCC(ctx, policyName, mccEntryPolicy, policy)
}

// DeregisterPolicy - remove a new policy from mcc.
func (mcs *MultiCluster) DeregisterPolicy(ctx context.Context, policyName string) error {
	if !mcs.IsJoined() {
		return nil
	}
	return deleteKeyEtcd(ctx, mcs.client, getMCCKeyPath(policyName, mccEntryPolicy))
}

func getMCCIdentityEntryFor(uType IAMUserType) (entry mccEntry, err error) {
	switch uType {
	case svcUser:
		entry = mccEntryServiceAccount
	case stsUser:
		entry = mccEntrySTSAccount
	default:
		err = errors.New("unsupported IAMUserType")
	}
	return
}

func getUserIdentityEtcd(ctx context.Context, client *etcd.Client, uType IAMUserType, accessKey string) (*UserIdentity, error) {
	entry, err := getMCCIdentityEntryFor(uType)
	if err != nil {
		return nil, err
	}
	val, err := readKeyEtcd(ctx, client, getMCCKeyPath(accessKey, entry))
	if err == errConfigNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var u UserIdentity
	err = json.Unmarshal(val, &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// DoesAccessKeyExist - check if an access key is present in the multi-cluster.
func (mcs *MultiCluster) DoesAccessKeyExist(ctx context.Context, t IAMUserType, accessKey string) (bool, error) {
	if !mcs.IsJoined() {
		return false, nil
	}

	u, err := getUserIdentityEtcd(ctx, mcs.client, t, accessKey)
	return u != nil, err
}

// RegisterUserIdentity - add a user identity to mcc.
func (mcs *MultiCluster) RegisterUserIdentity(ctx context.Context, t IAMUserType, u UserIdentity) error {
	if !mcs.IsJoined() {
		return nil
	}
	entry, err := getMCCIdentityEntryFor(t)
	if err != nil {
		return err
	}
	var putOp etcd.Op
	if t == stsUser {
		lease, err := mcs.client.Grant(ctx, int64(time.Until(u.Credentials.Expiration).Seconds()))
		if err != nil {
			return etcdErrToErr(err, mcs.client.Endpoints())
		}
		putOp, err = getMCCOpPut(u.Credentials.AccessKey, entry, u, etcd.WithLease(lease.ID))
		if err != nil {
			return etcdErrToErr(err, mcs.client.Endpoints())
		}
	} else {
		putOp, err = getMCCOpPut(u.Credentials.AccessKey, entry, u)
		if err != nil {
			return etcdErrToErr(err, mcs.client.Endpoints())
		}
	}
	_, err = doEtcd(ctx, mcs.client, putOp)
	return err
}

// DeregisterUserIdentity - remove a user identity from mcc.
func (mcs *MultiCluster) DeregisterUserIdentity(ctx context.Context, t IAMUserType, accessKey string) error {
	if !mcs.IsJoined() {
		return nil
	}
	entry, err := getMCCIdentityEntryFor(t)
	if err != nil {
		return err
	}
	return deleteKeyEtcd(ctx, mcs.client, getMCCKeyPath(accessKey, entry))
}

func getMCCPolicyDBEntryFor(uType IAMUserType, isGroup bool) (entry mccEntry, err error) {
	if isGroup {
		entry = mccEntryPolicyDBGroup
	} else {
		switch uType {
		case stsUser:
			entry = mccEntryPolicyDBSTSUser
		case svcUser:
			entry = mccEntryPolicyDBServiceAccount
		case regUser:
			entry = mccEntryPolicyDBUser
		default:
			err = errors.New("unsupported IAMUserType")
		}
	}
	return
}

// RegisterPolicyMapping - register a policy mapping in mcc.
func (mcs *MultiCluster) RegisterPolicyMapping(ctx context.Context, name string, t IAMUserType, isGroup bool, mp MappedPolicy, ttl int64) error {
	if !mcs.IsJoined() {
		return nil
	}
	entry, err := getMCCPolicyDBEntryFor(t, isGroup)
	if err != nil {
		return err
	}
	var putOp etcd.Op
	if ttl > 0 {
		lease, err := mcs.client.Grant(ctx, ttl)
		if err != nil {
			return etcdErrToErr(err, mcs.client.Endpoints())
		}
		putOp, err = getMCCOpPut(name, entry, mp, etcd.WithLease(lease.ID))
		if err != nil {
			return etcdErrToErr(err, mcs.client.Endpoints())
		}
	} else {
		putOp, err = getMCCOpPut(name, entry, mp)
		if err != nil {
			return etcdErrToErr(err, mcs.client.Endpoints())
		}
	}
	_, err = doEtcd(ctx, mcs.client, putOp)
	return err
}

// DeregisterPolicyMapping - remove a policy mapping in mcc.
func (mcs *MultiCluster) DeregisterPolicyMapping(ctx context.Context, name string, t IAMUserType, isGroup bool) error {
	if mcs == nil || !mcs.IsJoined() {
		return nil
	}
	entry, err := getMCCPolicyDBEntryFor(t, isGroup)
	if err != nil {
		return err
	}
	return deleteKeyEtcd(ctx, mcs.client, getMCCKeyPath(name, entry))
}

func (mcs *MultiCluster) putMCC(ctx context.Context, name string, entryType mccEntry, value interface{}) error {
	op, err := getMCCOpPut(name, entryType, value)
	if err != nil {
		return err
	}

	_, err = doEtcd(ctx, mcs.client, op)
	return err
}

func handleMCCIAMEvent(key string, isCreateOrModify bool, value []byte) error {
	fmt.Printf("Got EVENT: key=%s isCreateOrModify=%v\n", key, isCreateOrModify)
	switch {
	case strings.HasPrefix(key, iamConfigPoliciesPrefix):
		policyName := strings.TrimPrefix(key, iamConfigPoliciesPrefix)
		if isCreateOrModify {
			return savePolicy(policyName, value)
		}
		return deletePolicy(policyName)

	case strings.HasPrefix(key, iamConfigServiceAccountsPrefix):
		accessKey := strings.TrimPrefix(key, iamConfigServiceAccountsPrefix)
		if isCreateOrModify {
			return saveServiceAccount(accessKey, value)
		}
		return deleteServiceAccount(accessKey)

	case strings.HasPrefix(key, iamConfigSTSPrefix):
		accessKey := strings.TrimPrefix(key, iamConfigSTSPrefix)
		if isCreateOrModify {
			return saveSTSAccount(accessKey, value)
		}
		// There is no delete for STS keys - they expire
		// according to their TTL.
		return nil

	case strings.HasPrefix(key, iamConfigPolicyDBUsersPrefix):
		username := strings.TrimPrefix(key, iamConfigPolicyDBUsersPrefix)
		if isCreateOrModify {
			return savePolicyMapping(username, regUser, false, value)
		}
		return deletePolicyMapping(username, regUser, false)

	case strings.HasPrefix(key, iamConfigPolicyDBSTSUsersPrefix):
		username := strings.TrimPrefix(key, iamConfigPolicyDBSTSUsersPrefix)
		if isCreateOrModify {
			return savePolicyMapping(username, stsUser, false, value)
		}
		return deletePolicyMapping(username, stsUser, false)

	case strings.HasPrefix(key, iamConfigPolicyDBServiceAccountsPrefix):
		username := strings.TrimPrefix(key, iamConfigPolicyDBServiceAccountsPrefix)
		if isCreateOrModify {
			return savePolicyMapping(username, svcUser, false, value)
		}
		return deletePolicyMapping(username, svcUser, false)

	case strings.HasPrefix(key, iamConfigPolicyDBGroupsPrefix):
		username := strings.TrimPrefix(key, iamConfigPolicyDBGroupsPrefix)
		if isCreateOrModify {
			return savePolicyMapping(username, regUser, true, value)
		}
		return deletePolicyMapping(username, regUser, true)

	}
	return errors.New("unknown event type")
}

func savePolicy(policyName string, pbs []byte) error {
	iamPolicy, err := iampolicy.ParseConfig(bytes.NewReader(pbs))
	if err != nil {
		return err
	}
	if err = globalIAMSys.SetPolicy(policyName, *iamPolicy); err != nil {
		return err
	}
	// Notify all other MinIO peers to reload policy
	for _, nerr := range globalNotificationSys.LoadPolicy(policyName) {
		if nerr.Err != nil {
			logger.Error("SetPolicy notification error (%s): Host:%s - %v", policyName, nerr.Host.String(), nerr.Err)
		}
	}
	return nil
}

func deletePolicy(policyName string) error {
	if err := globalIAMSys.DeletePolicy(policyName); err != nil {
		return err
	}

	// Notify all other MinIO peers to delete policy
	for _, nerr := range globalNotificationSys.DeletePolicy(policyName) {
		if nerr.Err != nil {
			logger.Error("DeletePolicy notification error (%s): Host:%s - %v", policyName, nerr.Host.String(), nerr.Err)
		}
	}
	return nil
}

func saveServiceAccount(accessKey string, bs []byte) error {
	var u UserIdentity
	err := json.Unmarshal(bs, &u)
	if err != nil {
		return err
	}

	if err = globalIAMSys.store.saveUserIdentity(context.Background(), u.Credentials.AccessKey, svcUser, u); err != nil {
		return err
	}
	objAPI := newObjectLayerFn()
	err = globalIAMSys.LoadUser(objAPI, u.Credentials.AccessKey, svcUser)
	if err != nil {
		return err
	}

	for _, nerr := range globalNotificationSys.LoadServiceAccount(u.Credentials.AccessKey) {
		if nerr.Err != nil {
			logger.Error("LoadServiceAccount notification error (%s): Host:%s - %v", u.Credentials.AccessKey, nerr.Host.String(), nerr.Err)
		}
	}
	return nil
}

func deleteServiceAccount(accessKey string) error {
	err := globalIAMSys.DeleteUser(accessKey)
	if err != nil && err != errNoSuchUser {
		return err
	}

	for _, nerr := range globalNotificationSys.DeleteServiceAccount(accessKey) {
		if nerr.Err != nil {
			logger.Error("DeleteServiceAccount notification error (%s): Host:%s - %v", accessKey, nerr.Host.String(), nerr.Err)
		}
	}
	return nil
}

func saveSTSAccount(accessKey string, bs []byte) error {
	var u UserIdentity
	err := json.Unmarshal(bs, &u)
	if err != nil {
		return err
	}

	if err = globalIAMSys.store.saveUserIdentity(context.Background(), u.Credentials.AccessKey, stsUser, u); err != nil {
		return err
	}
	objAPI := newObjectLayerFn()
	err = globalIAMSys.LoadUser(objAPI, u.Credentials.AccessKey, stsUser)
	if err != nil {
		return err
	}

	// Notify all other Minio peers to reload user the service account
	for _, nerr := range globalNotificationSys.LoadUser(u.Credentials.AccessKey, true) {
		if nerr.Err != nil {
			logger.Error("LoadUser notification error (%s): Host:%s - %v", u.Credentials.AccessKey, nerr.Host.String(), nerr.Err)
		}
	}
	return nil
}

func savePolicyMapping(name string, t IAMUserType, isGroup bool, value []byte) error {
	var mp MappedPolicy
	err := json.Unmarshal(value, &mp)
	if err != nil {
		return err
	}
	err = globalIAMSys.store.saveMappedPolicy(context.Background(), name, t, isGroup, mp)
	if err != nil {
		return err
	}
	objAPI := newObjectLayerFn()
	err = globalIAMSys.LoadPolicyMapping(objAPI, name, isGroup)
	if err != nil {
		return err
	}

	// Notify all other MinIO peers to reload policy
	for _, nerr := range globalNotificationSys.LoadPolicyMapping(name, isGroup) {
		if nerr.Err != nil {
			logger.Error("LoadPolicyMapping notification error (%s): Host:%s - %v", name, nerr.Host.String(), nerr.Err)
		}
	}

	return nil
}

func deletePolicyMapping(name string, t IAMUserType, isGroup bool) error {
	err := globalIAMSys.store.deleteMappedPolicy(context.Background(), name, t, isGroup)
	if err != nil {
		return err
	}
	objAPI := newObjectLayerFn()
	globalIAMSys.LoadPolicyMapping(objAPI, name, isGroup)
	if err != nil {
		return err
	}

	// Notify all other MinIO peers to reload policy
	for _, nerr := range globalNotificationSys.LoadPolicyMapping(name, isGroup) {
		if nerr.Err != nil {
			logger.Error("LoadPolicyMapping notification error (%s): Host:%s - %v", name, nerr.Host.String(), nerr.Err)
		}
	}
	return nil
}

func collectServerURLs() []string {
	endpointURLs := make(map[url.URL]bool)
	for _, ep := range globalEndpoints {
		for _, endpoint := range ep.Endpoints {
			url := url.URL{
				Scheme: endpoint.Scheme,
				Host:   endpoint.Host,
			}
			endpointURLs[url] = true
		}
	}
	urls := make([]string, len(endpointURLs))
	i := 0
	for k := range endpointURLs {
		urls[i] = k.String()
		i++
	}
	return urls
}
