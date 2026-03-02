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
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
)

// HandleApiError delegates to the shared library's HandleAPIError function for centralized error handling
// with network error detection and standardized diagnostics formatting
func HandleApiError(message string, err *error, httpResponse *http.Response, respDiags *diag.Diagnostics) {
	sharederr.HandleAPIError(message, err, httpResponse, respDiags)
}

// HandleApiWarning delegates to the shared library's HandleAPIWarning function for non-critical API issues
func HandleApiWarning(message string, err *error, httpResponse *http.Response, respDiags *diag.Diagnostics) {
	sharederr.HandleAPIWarning(message, err, httpResponse, respDiags)
}
