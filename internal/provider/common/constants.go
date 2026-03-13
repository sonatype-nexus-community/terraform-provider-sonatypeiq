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

import "regexp"

const (
	DEFAULT_MAIL_SERVER_PORT          int32  = 465
	DEFAULT_MAIL_SSL_ENABLED          bool   = true
	DEFAULT_MAIL_START_TLS_ENABLED    bool   = true
	DEFAULT_USER_REALM                string = "Internal"
	MEMBER_TYPE_GROUP                 string = "group"
	MEMBER_TYPE_USER                  string = "user"
	OWNER_TYPE_APPLICATION            string = "application"
	OWNER_TYPE_ORGANIZATION           string = "organization"
	ROOT_ORGANIZATION_ID              string = "ROOT_ORGANIZATION_ID"
	SAML_DEFAULT_EMAIL_ATTRIBUTE      string = "email"
	SAML_DEFAULT_FIRST_NAME_ATTRIBUTE string = "firstName"
	SAML_DEFAULT_GROUPS_ATTRIBUTE     string = "groups"
	SAML_DEFAULT_LAST_NAME_ATTRIBUTE  string = "lastName"
	SCM_CONFIG_ID_FORMAT              string = "scm-config-%s-%s"
	SCM_PROVIDER_AZURE_DEVOPS         string = "azure"
	SCM_PROVIDER_BITBUCKET            string = "bitbucket"
	SCM_PROVIDER_GITHUB               string = "github"
	SCM_PROVIDER_GITLAB               string = "gitlab"
	STATE_ID_CROWD_CONFIGURATION      string = "system-crowd-configuration"
	STATE_ID_MAIL_CONFIGURATION       string = "system-mail-configuration"
	STATE_ID_IQ_PRODUCT_LICENSE       string = "system-product-license"
	STATE_ID_PROXY_CONFIGURATION      string = "system-proxy-configuration"
	STATE_ID_SAML_CONFIGURATION       string = "system-saml-configuration"
	USER_REALM_INTERNAL               string = "Internal"
	USER_REALM_SAML                   string = "SAML"
	USER_REALM_OAUTH2                 string = "OAUTH2"
	USER_REALM_CROWD                  string = "CROWD"
)

var (
	internalIdRegex, _            = regexp.Compile(`^[a-z0-9]{32}$`)
	APPLICATION_INTERNAL_ID_REGEX = internalIdRegex
	ORGANIZATION_ID_REGEX         = internalIdRegex
	ROLE_ID_REGEX                 = internalIdRegex
)
