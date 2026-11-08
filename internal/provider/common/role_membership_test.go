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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleMembershipScopeAdviceForAGlobalRoleOnAnOrganization(t *testing.T) {
	err := errors.New("400 Bad Request: 400 Bad Request: Cannot map members to global role in context of organization.")

	summary, detail, ok := RoleMembershipScopeAdvice(err)

	assert.True(t, ok)
	assert.Equal(t, "This role is a global role", summary)
	assert.Contains(t, detail, "sonatypeiq_global_role_membership")
	assert.Contains(t, detail, "Cannot map members to global role", "the server's own message is kept")
}

func TestRoleMembershipScopeAdviceForANonGlobalRoleGlobally(t *testing.T) {
	err := errors.New("400 Bad Request: Cannot map members to application role in global context.")

	summary, detail, ok := RoleMembershipScopeAdvice(err)

	assert.True(t, ok)
	assert.Equal(t, "This role is not a global role", summary)
	assert.Contains(t, detail, "sonatypeiq_organization_role_membership")
	assert.Contains(t, detail, "sonatypeiq_application_role_membership")
}

func TestRoleMembershipScopeAdviceIgnoresOtherErrors(t *testing.T) {
	for _, err := range []error{nil, errors.New("403 Forbidden: Insufficient permissions"), errors.New("500 Server Error")} {
		_, _, ok := RoleMembershipScopeAdvice(err)
		assert.False(t, ok, "%v", err)
	}
}
