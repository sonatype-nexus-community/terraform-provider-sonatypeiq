/*
 * Copyright (c) 2019-present Sonatype, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the \"License\");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an \"AS IS\" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"context"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// AuthContext represents the authentication information needed for API calls
type AuthContext struct {
	Auth sonatypeiq.BasicAuth
}

// NewAuthContext creates an AuthContext from BasicAuth
func NewAuthContext(auth sonatypeiq.BasicAuth) *AuthContext {
	return &AuthContext{Auth: auth}
}

// WithAuthContext adds authentication to the context for API calls
func WithAuthContext(ctx context.Context, authCtx *AuthContext) context.Context {
	return context.WithValue(ctx, sonatypeiq.ContextBasicAuth, authCtx.Auth)
}

// WithAuth adds authentication directly from BasicAuth
func WithAuth(ctx context.Context, auth sonatypeiq.BasicAuth) context.Context {
	return context.WithValue(ctx, sonatypeiq.ContextBasicAuth, auth)
}
