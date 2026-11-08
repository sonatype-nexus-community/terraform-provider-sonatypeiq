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
	"os"
	"sync"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"testing"
)

// CurrenTestIqVersion is the release of the Sonatype IQ Server under test, taken from NXIQ_VERSION (e.g.
// 1.201.0-02), which CI sets. It is 0 when NXIQ_VERSION is not set: use DetectedIqVersion, which then asks the
// server itself.
var CurrenTestIqVersion = common.ParseServerHeaderToVersion(fmt.Sprintf("NexusIQ/%s", os.Getenv("NXIQ_VERSION")))

func SkipIfNxiqVersionEq(t *testing.T, v int32) {
	t.Helper()

	if v == DetectedIqVersion(t) {
		t.Skipf("NXIQ Version is == %v - skipping", v)
	}
}

func SkipIfNxiqVersionOlderThan(t *testing.T, v int32) {
	t.Helper()

	if VersionOlderThan(DetectedIqVersion(t), v) {
		t.Skipf("NXIQ Version is older than %v - skipping", v)
	}
}

func SkipIfNxiqVersionInRange(t *testing.T, low, high int32) {
	t.Helper()

	inRange, err := VersionInRange(DetectedIqVersion(t), low, high)

	if err != nil {
		t.Errorf("Error comparing versions: %v", err)
		t.FailNow()
	}

	if inRange {
		t.Skipf("NXIQ Version within range %v and %v - skipping", low, high)
	}
}

func VersionInRange(ver, low, high int32) (bool, error) {
	if low <= ver && high >= ver {
		return true, nil
	}

	return false, nil
}

func VersionOlderThan(thisVer, testVer int32) bool {
	return thisVer < testVer
}

var (
	detectOnce    sync.Once
	detectedIq    int32
	detectedIqErr error
)

// DetectedIqVersion returns the release of the Sonatype IQ Server under test. It uses NXIQ_VERSION when set,
// otherwise it asks the server (IQ_SERVER_URL) for its version, so version-dependent tests are not silently
// skipped when only the connection details are known, e.g. when running locally. The test is skipped if the
// version cannot be determined. The server is asked only once.
func DetectedIqVersion(t *testing.T) int32 {
	t.Helper()

	detectOnce.Do(func() {
		if CurrenTestIqVersion != 0 {
			detectedIq = CurrenTestIqVersion
			return
		}
		detectedIq, detectedIqErr = ProbeIqVersion(os.Getenv("IQ_SERVER_URL"))
	})

	if detectedIqErr != nil {
		t.Skipf("Cannot determine Sonatype IQ Server version: %v", detectedIqErr)
	}
	return detectedIq
}

// ProbeIqVersion asks a Sonatype IQ Server for its release using the Server header of its anonymous /ping
// endpoint, so it needs no credentials.
func ProbeIqVersion(serverUrl string) (int32, error) {
	resp, err := http.Get(serverUrl + "/ping") //nolint:gosec // the URL is the Sonatype IQ Server under test
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	version := common.ParseServerHeaderToVersion(resp.Header.Get("server"))
	if version == 0 {
		return 0, fmt.Errorf("unrecognised Server header %q", resp.Header.Get("server"))
	}
	return version, nil
}
