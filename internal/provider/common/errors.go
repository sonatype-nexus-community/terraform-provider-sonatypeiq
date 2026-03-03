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
	ERR_TF_GETTING_PLAN  string = "Getting Plan has errors: %v"
	ERR_TF_GETTING_STATE string = "Getting State has errors: %v"

	ERR_APPLICATION_DID_NOT_EXIST                     string = "Application did not exist: %s"
	ERR_APPLICATION_CATEGORY_FOR_ORG_DID_NOT_EXIST    string = "Application Category for Organization did not exist: %s"
	ERR_FAILED_DELETING_APPLICATION_ROLE_MAPPING      string = "Failed to delete Application Role Mapping: %s"
	ERR_FAILED_MOVING_APPLICATION                     string = "Failed moving Application to a new Organization"
	ERR_FAILED_READING_APPLICATION                    string = "Unable to read Application"
	ERR_FAILED_READING_APPLICATIONS                   string = "Unable to read Applications"
	ERR_FAILED_READING_APPLICATION_CATEGORIES_FOR_ORG string = "Unable to read IQ Application Categories for Organization"
	ERR_FAILED_READING_ORGANIZATIONS                  string = "Unable to read Organizations"
	ERR_FAILED_READING_ROLES                          string = "Unable to read Roles"
	ERR_FAILED_READING_SAML_METADATA                  string = "Unable to read SAML Metadata"
	ERR_FAILED_READING_SYSTEM_CONFIG                  string = "Unable to read System Configuration"
)
