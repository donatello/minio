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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/minio/madmin-go"
	"github.com/minio/minio/internal/logger"
	iampolicy "github.com/minio/pkg/iam/policy"
)

// ClusterLink - links with the given cluster. Enables global namespace, global
// IAM and global administration. NOTE: Etcd must be configured with mcc
// enabled, prior to this operation on the current cluster, and not in the peer
// cluster.
func (a adminAPIHandlers) ClusterLink(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "ClusterLink")

	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))

	objectAPI, cred := validateAdminReq(ctx, w, r, iampolicy.ClusterLinkAdminAction)
	if objectAPI == nil {
		return
	}

	fmt.Println("ClusterLink")

	if globalMultiCluster == nil {
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(ErrAdminMultiClusterNotConfigured), r.URL)
		return
	}

	result, err := globalMultiCluster.Join(ctx)
	if err != nil {
		writeErrorResponseJSON(ctx, w, toAdminAPIErr(ctx, err), r.URL)
		return
	}

	data, err := json.Marshal(result)
	if err != nil {
		writeErrorResponseJSON(ctx, w, toAdminAPIErr(ctx, err), r.URL)
		return
	}

	encryptedData, err := madmin.EncryptData(cred.SecretKey, data)
	if err != nil {
		writeErrorResponseJSON(ctx, w, toAdminAPIErr(ctx, err), r.URL)
		return
	}

	writeSuccessResponseJSON(w, encryptedData)
}

// ClusterUnlink - removes the linking with the given cluster.
func (a adminAPIHandlers) ClusterUnlink(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "ClusterUnlink")

	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))

	objectAPI, _ := validateAdminReq(ctx, w, r, iampolicy.ClusterUnlinkAdminAction)
	if objectAPI == nil {
		return
	}

	fmt.Println("ClusterUnlink")
}

// ClusterInfo - returns cluster linking info about the cluster (self).
func (a adminAPIHandlers) ClusterInfo(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "ClusterInfo")

	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))

	objectAPI, _ := validateAdminReq(ctx, w, r, iampolicy.ClusterInfoAdminAction)
	if objectAPI == nil {
		return
	}

	fmt.Println("ClusterInfo")
}
