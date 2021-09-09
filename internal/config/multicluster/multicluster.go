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

package multicluster

import (
	"crypto/tls"
	"crypto/x509"
	"strings"
	"time"

	"github.com/minio/minio/internal/config"
	"github.com/minio/pkg/env"
	xnet "github.com/minio/pkg/net"
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

const (
	// Default values used while communicating with etcd.
	defaultDialTimeout   = 5 * time.Second
	defaultDialKeepAlive = 30 * time.Second
)

// configuration parameter names
const (
	ControllerEndpoints     = "controller_endpoints"
	ControllerClientCert    = "controller_client_cert"
	ControllerClientCertKey = "controller_client_cert_key"

	EnvMultiClusterControllerEndpoints     = "MINIO_MULTI_CLUSTER_CONTROLLER_ENDPOINTS"
	EnvMultiClusterControllerClientCert    = "MINIO_MULTI_CLUSTER_CONTROLLER_CLIENT_CERT"
	EnvMultiClusterControllerClientCertKey = "MINIO_MULTI_CLUSTER_CONTROLLER_CLIENT_CERT_KEY"
)

// DefaultKVS - default KV for multi-cluster
var (
	DefaultKVS = config.KVS{
		config.KV{
			Key:   ControllerEndpoints,
			Value: "",
		},
		config.KV{
			Key:   ControllerClientCert,
			Value: "",
		},
		config.KV{
			Key:   ControllerClientCertKey,
			Value: "",
		},
	}
)

// Config - multicluster config
type Config struct {
	Enabled bool `json:"enabled"`

	etcd.Config
}

func New(cfg Config) (*etcd.Client, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	cli, err := etcd.New(cfg.Config)
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func parseEndpoints(endpoints string) ([]string, bool, error) {
	controllerEndpoints := strings.Split(endpoints, config.ValueSeparator)

	var isSecure bool
	for _, endpoint := range controllerEndpoints {
		u, err := xnet.ParseHTTPURL(endpoint)
		if err != nil {
			return nil, false, err
		}
		if isSecure && u.Scheme == "http" {
			return nil, false, config.Errorf("all endpoints should be https or http: %s", endpoint)
		}
		// If one of the endpoint is https, we will use https directly.
		isSecure = isSecure || u.Scheme == "https"
	}

	return controllerEndpoints, isSecure, nil
}

// Enabled returns if etcd is enabled.
func Enabled(kvs config.KVS) bool {
	endpoints := kvs.Get(ControllerEndpoints)
	return endpoints != ""
}

// LookupConfig - Initialize new etcd config.
func LookupConfig(kvs config.KVS, rootCAs *x509.CertPool) (Config, error) {
	cfg := Config{}
	if err := config.CheckValidKeys(config.MultiClusterSubSys, kvs, DefaultKVS); err != nil {
		return cfg, err
	}

	endpoints := env.Get(EnvMultiClusterControllerEndpoints, kvs.Get(ControllerEndpoints))
	if endpoints == "" {
		return cfg, nil
	}

	controllerEndpoints, isSecure, err := parseEndpoints(endpoints)
	if err != nil {
		return cfg, err
	}

	cfg.Enabled = true
	cfg.DialTimeout = defaultDialTimeout
	cfg.DialKeepAliveTime = defaultDialKeepAlive
	// Disable etcd client SDK logging, etcd client
	// incorrectly starts logging in unexpected data
	// format.
	cfg.LogConfig = &zap.Config{
		Level:    zap.NewAtomicLevelAt(zap.FatalLevel),
		Encoding: "console",
	}
	cfg.Endpoints = controllerEndpoints
	if isSecure {
		cfg.TLS = &tls.Config{
			RootCAs: rootCAs,
		}
		// This is only to support client side certificate authentication
		// https://coreos.com/etcd/docs/latest/op-guide/security.html
		etcdClientCertFile := env.Get(EnvMultiClusterControllerClientCert, kvs.Get(ControllerClientCert))
		etcdClientCertKey := env.Get(EnvMultiClusterControllerClientCertKey, kvs.Get(ControllerClientCertKey))
		if etcdClientCertFile != "" && etcdClientCertKey != "" {
			cfg.TLS.GetClientCertificate = func(unused *tls.CertificateRequestInfo) (*tls.Certificate, error) {
				cert, err := tls.LoadX509KeyPair(etcdClientCertFile, etcdClientCertKey)
				return &cert, err
			}
		}
	}
	return cfg, nil
}
