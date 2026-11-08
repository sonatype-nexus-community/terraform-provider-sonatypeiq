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
	"terraform-provider-sonatypeiq/internal/provider/common"
	"testing"
)

var CurrenTestIqVersion = common.ParseServerHeaderToVersion(fmt.Sprintf("NexusIQ/%s", os.Getenv("NXIQ_VERSION")))

func SkipIfNxiqVersionEq(t *testing.T, v int32) {
	t.Helper()

	if v == CurrenTestIqVersion {
		t.Skipf("NXIQ Version is == %v - skipping", v)
	}
}

func SkipIfNxiqVersionOlderThan(t *testing.T, v int32) {
	t.Helper()

	if VersionOlderThan(CurrenTestIqVersion, v) {
		t.Skipf("NXIQ Version is older than %v - skipping", v)
	}
}

func SkipIfNxiqVersionInRange(t *testing.T, low, high int32) {
	t.Helper()

	inRange, err := VersionInRange(CurrenTestIqVersion, low, high)

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

// DetectedIqVersion returns the release of the Sonatype IQ Server under test. It uses NXIQ_VERSION when set,
// otherwise it asks the server (IQ_SERVER_URL) for its version, so that version-dependent tests also run
// when only the connection details are known. The test is skipped if the version cannot be determined.
func DetectedIqVersion(t *testing.T) int32 {
	t.Helper()

	if CurrenTestIqVersion != 0 {
		return CurrenTestIqVersion
	}

	req, err := http.NewRequest(http.MethodGet, os.Getenv("IQ_SERVER_URL")+"/api/v2/users/"+os.Getenv("IQ_SERVER_USERNAME"), nil)
	if err != nil {
		t.Skipf("Cannot determine Sonatype IQ Server version: %v", err)
	}
	req.SetBasicAuth(os.Getenv("IQ_SERVER_USERNAME"), os.Getenv("IQ_SERVER_PASSWORD"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Skipf("Cannot determine Sonatype IQ Server version: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	version := common.ParseServerHeaderToVersion(resp.Header.Get("server"))
	if version == 0 {
		t.Skipf("Cannot determine Sonatype IQ Server version from Server header %q", resp.Header.Get("server"))
	}
	return version
}
