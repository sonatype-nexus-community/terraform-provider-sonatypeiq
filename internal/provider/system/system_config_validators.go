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
package system

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// frameAncestorsNone is the Content-Security-Policy keyword that forbids all framing.
const frameAncestorsNone = "'none'"

// noneMustBeAlone rejects a frame ancestors allowlist that combines 'none' with other values.
// Sonatype IQ Server stores only 'none' in that case, so the configuration could never be applied as written.
type noneMustBeAlone struct{}

func (noneMustBeAlone) Description(_ context.Context) string {
	return "'none' cannot be combined with other values"
}

func (v noneMustBeAlone) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (noneMustBeAlone) ValidateSet(_ context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || len(req.ConfigValue.Elements()) < 2 {
		return
	}
	for _, e := range req.ConfigValue.Elements() {
		if e.String() == `"`+frameAncestorsNone+`"` {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid frame ancestors allowlist",
				"`'none'` forbids all framing, and Sonatype IQ Server stores nothing else alongside it. "+
					"Use `'none'` on its own, or list the domains that may frame Sonatype IQ Server.",
			)
			return
		}
	}
}
