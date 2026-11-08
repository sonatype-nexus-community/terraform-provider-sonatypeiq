/*
 * Copyright (c) 2019-present Sonatype, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccVersionOlderThan(t *testing.T) {
	var testVer = common.ParseServerHeaderToVersion("NexusIQ/1.201.0-02")
	assert.False(t, VersionOlderThan(testVer, 199), fmt.Sprintf("%v is not older than 199 as expected", testVer))
	assert.True(t, VersionOlderThan(testVer, 202))
}

func TestProbeIqVersionReadsTheServerHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _, hasCredentials := r.BasicAuth()
		assert.False(t, hasCredentials, "the probe must not need credentials")
		assert.Equal(t, "/ping", r.URL.Path)
		w.Header().Set("Server", "NexusIQ/1.205.4-01")
	}))
	defer server.Close()

	version, err := ProbeIqVersion(server.URL)

	assert.NoError(t, err)
	assert.Equal(t, int32(205), version)
}

func TestProbeIqVersionRejectsAnUnknownServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "nginx")
	}))
	defer server.Close()

	_, err := ProbeIqVersion(server.URL)

	assert.Error(t, err)
}

func TestProbeIqVersionFailsWhenTheServerIsUnreachable(t *testing.T) {
	_, err := ProbeIqVersion("http://127.0.0.1:1")

	assert.Error(t, err)
}
