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
	if thisVer < testVer {
		return true
	}
	return false
}
