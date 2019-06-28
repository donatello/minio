/*
 * MinIO Cloud Storage, (C) 2019 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"crypto/tls"
	"os"

	ldap "gopkg.in/ldap.v3"
)

// ldapServerConfig contains server connectivity information.
type ldapServerConfig struct {
	IsEnabled bool `json:"enabled"`

	// E.g. "ldap.minio.io:636"
	ServerAddr string `json:"serveraddr"`

	// Use plain non-TLS insecure connection
	Insecure bool `json:"insecure"`

	// Skips TLS verification (for testing, not
	// recommended in production).
	SkipTLSVerify bool `json:"skiptlsverify"`
}

func (l *ldapServerConfig) Connect() (ldapConn *ldap.Conn, err error) {
	if l == nil {
		// Happens when LDAP is not configured.
		return
	}
	if l.Insecure {
		ldapConn, err = ldap.Dial("tcp", l.ServerAddr)
	} else if l.SkipTLSVerify {
		ldapConn, err = ldap.DialTLS("tcp", l.ServerAddr, &tls.Config{InsecureSkipVerify: true})
	} else {
		ldapConn, err = ldap.DialTLS("tcp", l.ServerAddr, &tls.Config{})
	}
	return
}

// newLDAPConfigFromEnv loads configuration from the environment
func newLDAPConfigFromEnv() (l ldapServerConfig) {
	if ldapServer, ok := os.LookupEnv("MINIO_LDAP_SERVER_ADDR"); ok {
		l.IsEnabled = true
		l.ServerAddr = ldapServer

		if v := os.Getenv("MINIO_LDAP_INSECURE"); v != "" {
			l.Insecure = true
		}

		if v := os.Getenv("MINIO_LDAP_TLS_SKIP_VERIFY"); v != "" {
			l.SkipTLSVerify = true
		}
	}
	return
}
