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
	"strings"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// Sonatype IQ Server refuses to grant a role in the wrong context. These are the beginnings of its messages.
const (
	serverGlobalRoleInNonGlobalContext = "Cannot map members to global role"
	serverNonGlobalRoleInGlobalContext = "Cannot map members to application role in global context"
)

// RoleMembershipScopeAdvice explains a refusal by Sonatype IQ Server to grant a role because it is global (or not)
// for the context that was used, and says which resource to use instead. ok is false for any other error.
func RoleMembershipScopeAdvice(err error) (summary, detail string, ok bool) {
	if err == nil {
		return "", "", false
	}

	switch message := serverMessage(err); {
	case strings.Contains(message, serverGlobalRoleInNonGlobalContext):
		return "This role is a global role",
			"Global roles, such as System Administrator, cannot be granted on an organization or an application. " +
				"Use the `sonatypeiq_global_role_membership` resource to grant a global role.\n\n" +
				"Sonatype IQ Server said: " + message, true
	case strings.Contains(message, serverNonGlobalRoleInGlobalContext):
		return "This role is not a global role",
			"Only global roles, such as System Administrator, can be granted globally. " +
				"Use the `sonatypeiq_organization_role_membership` or `sonatypeiq_application_role_membership` resource to grant this role.\n\n" +
				"Sonatype IQ Server said: " + message, true
	}
	return "", "", false
}

// serverMessage is the text of an API error including the body of the response: the API client's own error text is
// only the status line, and Sonatype IQ Server's explanation is in the body.
func serverMessage(err error) string {
	message := err.Error()

	var apiError *sonatypeiq.GenericOpenAPIError
	if errors.As(err, &apiError) && len(apiError.Body()) > 0 {
		message += ": " + strings.TrimSpace(string(apiError.Body()))
	}
	return message
}
