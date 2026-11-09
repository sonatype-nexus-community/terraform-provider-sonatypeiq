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

package common

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestRequireIqRelease_OlderServerIsRefused(t *testing.T) {
	var diags diag.Diagnostics
	RequireIqRelease(&diags, 190, 198, "user token configuration")

	if len(diags.Errors()) != 1 {
		t.Fatalf("expected 1 error, got %d", len(diags.Errors()))
	}
	want := "was introduced in Sonatype IQ Server release 198, but the connected server is release 190"
	if got := diags.Errors()[0].Detail(); !strings.Contains(got, want) {
		t.Errorf("error detail %q does not contain %q", got, want)
	}
}

func TestRequireIqRelease_SupportedOrUndetectedServerIsAccepted(t *testing.T) {
	for name, iqVersion := range map[string]int32{
		"equal":      198,
		"newer":      205,
		"undetected": 0,
	} {
		t.Run(name, func(t *testing.T) {
			var diags diag.Diagnostics
			RequireIqRelease(&diags, iqVersion, 198, "user token configuration")

			if diags.HasError() {
				t.Errorf("expected no error but got: %v", diags.Errors())
			}
		})
	}
}
