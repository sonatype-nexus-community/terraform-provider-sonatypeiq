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

const (
	ROLE_PERMISSION_CATEGORY_ADNIN       string = "Administrator"
	ROLE_PERMISSION_CATEGORY_IQ          string = "IQ"
	ROLE_PERMISSION_CATEGORY_REMEDIATION string = "Remediation"

	// Administrator
	ROLE_ID_ACCESS_AUDIT_LOGS string = "ACCESS_AUDIT_LOG"
	ROLE_ID_VIEW_ROLES        string = "VIEW_ROLES"

	// IQ
	ROLE_ID_ADD_APPLICATIONS                      string = "ADD_APPLICATION"
	ROLE_ID_CLAIM_COMPONENTS                      string = "CLAIM_COMPONENT"
	ROLE_ID_EDIT_ACCESS_CONTROL                   string = "EDIT_ACCESS_CONTROL"
	ROLE_ID_EDIT_IQ_ELEMENTS                      string = "WRITE"
	ROLE_ID_EDIT_PROPRIETARY_COMPONENTS           string = "MANAGE_PROPRIETARY"
	ROLE_ID_EVALUATE_APPLICATIONS                 string = "EVALUATE_APPLICATION"
	ROLE_ID_EVALUATE_INDIVIDUAL_COMPONENTS        string = "EVALUATE_COMPONENT"
	ROLE_ID_MANAGE_AUTOMATIC_APPLICATION_CREATION string = "MANAGE_AUTOMATIC_APPLICATION_CREATION"
	ROLE_ID_MANAGE_AUTOMATIC_SCM_CONFIGURATION    string = "MANAGE_AUTOMATIC_SCM_CONFIGURATION"
	ROLE_ID_VIEW_IQ_ELEMENTS                      string = "READ"

	// Remediation
	ROLE_ID_CHANGE_LICENSES                 string = "CHANGE_LICENSES"
	ROLE_ID_CHANGE_SECURITY_VULNERABILITIES string = "CHANGE_SECURITY_VULNERABILITIES"
	ROLE_ID_CREATE_PULL_REQUESTS            string = "CREATE_PULL_REQUESTS"
	ROLE_ID_REVIEW_LEGAL_OBLIGATIONS        string = "LEGAL_REVIEWER"
	ROLE_ID_WAIVE_POLICY_VIOLATIONS         string = "WAIVE_POLICY_VIOLATIONS"
)
