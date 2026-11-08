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
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)

// fakeServer is a Sonatype IQ Server stand-in for the start-up check. usersCalls counts calls to the
// administrator-only users endpoint, which the start-up check must never make.
type fakeServer struct {
	*httptest.Server
	usersCalls atomic.Int32
}

func newFakeServer(t *testing.T, pingServerHeader string, tokenStatus int, tokenServerHeader string) *fakeServer {
	t.Helper()
	f := &fakeServer{}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ping":
			if pingServerHeader != "" {
				w.Header().Set("Server", pingServerHeader)
			}
		case "/api/v2/userTokens/currentUser/hasToken":
			if tokenServerHeader != "" {
				w.Header().Set("Server", tokenServerHeader)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(tokenStatus)
			_, _ = w.Write([]byte(`{}`))
		default:
			f.usersCalls.Add(1)
			http.Error(w, "Insufficient permissions", http.StatusForbidden)
		}
	}))
	t.Cleanup(f.Close)
	return f
}

func dataSourceDataFor(url string) *SonatypeDataSourceData {
	configuration := sonatypeiq.NewConfiguration()
	configuration.Servers = []sonatypeiq.ServerConfiguration{{URL: url}}
	return &SonatypeDataSourceData{
		Auth:    sonatypeiq.BasicAuth{UserName: "limited", Password: "secret"},
		BaseUrl: url,
		Client:  sonatypeiq.NewAPIClient(configuration),
	}
}

func check(t *testing.T, url string, hint *string) (*SonatypeDataSourceData, diag.Diagnostics) {
	t.Helper()
	data := dataSourceDataFor(url)
	var diags diag.Diagnostics
	data.CheckConnectionAndGetVersion(context.Background(), &diags, hint)
	return data, diags
}

func strPtr(s string) *string { return &s }

// The reported problem: a user without the permission to read users must still be able to use the provider.
func TestCheckConnectionDoesNotNeedPermissionToReadUsers(t *testing.T) {
	f := newFakeServer(t, "NexusIQ/1.204.2-01", http.StatusNotFound, "NexusIQ/1.204.2-01") // 404: no token yet

	data, diags := check(t, f.URL, nil)

	assert.False(t, diags.HasError(), "unexpected errors: %v", diags)
	assert.Equal(t, int32(204), data.IqVersion)
	assert.Zero(t, f.usersCalls.Load(), "the administrator-only users endpoint must not be called")
}

func TestCheckConnectionFallsBackToTheServerHeaderOfTheCredentialsCheck(t *testing.T) {
	f := newFakeServer(t, "", http.StatusOK, "NexusIQ/1.198.0-01") // /ping has no usable header

	data, diags := check(t, f.URL, nil)

	assert.False(t, diags.HasError())
	assert.Equal(t, int32(198), data.IqVersion)
}

// A reverse proxy that rewrites the Server header must not stop the provider; the version is just unknown.
func TestCheckConnectionWithARewrittenServerHeaderLeavesTheVersionUnknown(t *testing.T) {
	f := newFakeServer(t, "nginx/1.25", http.StatusOK, "nginx/1.25")

	data, diags := check(t, f.URL, nil)

	assert.False(t, diags.HasError(), "an unknown version must not be an error")
	assert.Equal(t, int32(0), data.IqVersion)
}

func TestCheckConnectionVersionHintTakesPrecedence(t *testing.T) {
	f := newFakeServer(t, "NexusIQ/1.205.4-01", http.StatusOK, "NexusIQ/1.205.4-01")

	for hint, want := range map[string]int32{"1.190.0-01": 190, "198": 198} {
		data, diags := check(t, f.URL, strPtr(hint))
		assert.False(t, diags.HasError())
		assert.Equal(t, want, data.IqVersion, "hint %q", hint)
	}
}

func TestCheckConnectionVersionHintWorksWhenTheHeaderIsRewritten(t *testing.T) {
	f := newFakeServer(t, "nginx", http.StatusOK, "nginx")

	data, diags := check(t, f.URL, strPtr("1.201.0-02"))

	assert.False(t, diags.HasError())
	assert.Equal(t, int32(201), data.IqVersion)
}

func TestCheckConnectionRejectedCredentialsAreAnError(t *testing.T) {
	f := newFakeServer(t, "NexusIQ/1.205.4-01", http.StatusUnauthorized, "NexusIQ/1.205.4-01")

	_, diags := check(t, f.URL, nil)

	assert.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Summary(), "rejected the supplied credentials")
}

func TestCheckConnectionUnreachableServerIsAnError(t *testing.T) {
	_, diags := check(t, "http://127.0.0.1:1", nil)

	assert.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Summary(), "cannot be accessed")
}

func TestParseVersionHint(t *testing.T) {
	cases := map[string]int32{
		"1.205.4-01": 205,
		"1.186.0-01": 186,
		"205":        205,
		" 198 ":      198,
		"":           0,
		"abc":        0,
		"1.205":      0,
		"2.205.4-01": 0,
	}
	for hint, want := range cases {
		assert.Equal(t, want, ParseVersionHint(hint), "hint %q", hint)
	}
}

func TestVersionHintExpMatchesWhatParseVersionHintAccepts(t *testing.T) {
	for _, valid := range []string{"1.205.4-01", "205", "186"} {
		assert.True(t, VersionHintExp.MatchString(valid), valid)
	}
	for _, invalid := range []string{"", "abc", "1.205", "205.1", "NexusIQ/1.205.4-01", "1.205.4"} {
		assert.False(t, VersionHintExp.MatchString(invalid), invalid)
	}
}
