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
	"terraform-provider-sonatypeiq/internal/provider/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccVersionOlderThan(t *testing.T) {
	var testVer = common.ParseServerHeaderToVersion("NexusIQ/1.201.0-02")
	assert.False(t, VersionOlderThan(testVer, 199), fmt.Sprintf("%v is not older than 199 as expected", testVer))
	assert.True(t, VersionOlderThan(testVer, 202))
}
