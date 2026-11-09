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
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// RequireIqRelease adds an error diagnostic if the connected IQ Server version is known and
// older than the release that introduced a feature. A zero version (undetected) is never blocked.
func RequireIqRelease(diags *diag.Diagnostics, iqVersion, introducedIn int32, what string) {
	if iqVersion != 0 && iqVersion < introducedIn {
		diags.AddError(
			fmt.Sprintf("%s is not supported by this Sonatype IQ Server version", what),
			fmt.Sprintf("%s was introduced in Sonatype IQ Server release %d, but the connected server is release %d. Remove it from your configuration, or upgrade Sonatype IQ Server to release %d or later.",
				what, introducedIn, iqVersion, introducedIn),
		)
	}
}
