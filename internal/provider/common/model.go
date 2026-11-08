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
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
)

type SonatypeDataSourceData struct {
	Auth      sonatypeiq.BasicAuth
	BaseUrl   string
	Client    *sonatypeiq.APIClient
	IqVersion int32
}

// CheckConnectionAndGetVersion checks that Sonatype IQ Server can be reached with the supplied credentials and
// works out which release it is. It needs no permissions beyond being a valid user, so it works for the
// unprivileged users that only manage the resources they have permissions for.
//
// The release comes from versionHint when one is supplied (e.g. when a reverse proxy rewrites the Server header),
// otherwise from the Server header of the anonymous /ping endpoint, falling back to the header of the credentials
// check. It stays 0 when it cannot be determined, which never blocks anything: features that depend on the
// release are then simply not restricted.
func (p *SonatypeDataSourceData) CheckConnectionAndGetVersion(ctx context.Context, respDiags *diag.Diagnostics, versionHint *string) {
	pingHeader := p.pingServerHeader(ctx)
	credentialsHeader := p.checkCredentials(ctx, respDiags)
	if respDiags.HasError() {
		return
	}

	switch {
	case versionHint != nil:
		p.IqVersion = ParseVersionHint(*versionHint)
		tflog.Debug(ctx, fmt.Sprintf("Version Hint: %v", p.IqVersion))
	case ParseServerHeaderToVersion(pingHeader) != 0:
		p.IqVersion = ParseServerHeaderToVersion(pingHeader)
		tflog.Debug(ctx, fmt.Sprintf("Server Header from /ping: %v", p.IqVersion))
	default:
		p.IqVersion = ParseServerHeaderToVersion(credentialsHeader)
		tflog.Debug(ctx, fmt.Sprintf("Server Header from credentials check: %v", p.IqVersion))
	}

	tflog.Info(ctx, fmt.Sprintf("Determined Sonatype IQ Server to be version %v", p.IqVersion))
}

// pingServerHeader returns the Server header of the anonymous /ping endpoint, or "" if it cannot be read.
func (p *SonatypeDataSourceData) pingServerHeader(ctx context.Context) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.BaseUrl+"/ping", nil)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("Cannot build /ping request: %v", err))
		return ""
	}

	httpClient := http.DefaultClient
	if configured := p.Client.GetConfig().HTTPClient; configured != nil {
		httpClient = configured
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("Cannot reach /ping: %v", err))
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.Header.Get("server")
}

// checkCredentials makes a call that any authenticated user is allowed to make: it asks whether the current user
// has a User Token. Only an unreachable server or rejected credentials are errors - any other response (for example
// 404 when the user has no token yet) means the server was reached and the credentials were accepted.
// It returns the Server header of the response.
func (p *SonatypeDataSourceData) checkCredentials(ctx context.Context, respDiags *diag.Diagnostics) string {
	_, httpResponse, err := p.Client.UserTokensAPI.GetUserTokenExistsForCurrentUser(
		context.WithValue(ctx, sonatypeiq.ContextBasicAuth, p.Auth),
	).Execute()

	if httpResponse == nil {
		respDiags.AddError("Sonatype IQ Server cannot be accessed.", fmt.Sprintf("%v", err))
		return ""
	}

	if httpResponse.StatusCode == http.StatusUnauthorized {
		sharederr.HandleAPIError("Sonatype IQ Server rejected the supplied credentials.", &err, httpResponse, respDiags)
		return ""
	}

	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("Credentials check returned %v - treating the server as reachable", httpResponse.StatusCode))
	}

	return httpResponse.Header.Get("server")
}

var (
	versionHintFullExp    = regexp.MustCompile(`^1\.\d+\.\d+-\d+$`)
	versionHintReleaseExp = regexp.MustCompile(`^\d+$`)

	// VersionHintExp matches the values accepted for the provider's version_hint
	VersionHintExp = regexp.MustCompile(`^(1\.\d+\.\d+-\d+|\d+)$`)
)

// ParseVersionHint turns a version_hint into a release number. It accepts the full version as shown in the
// Server header (e.g. 1.205.4-01) or just the release number (e.g. 205), and returns 0 for anything else.
func ParseVersionHint(hint string) int32 {
	hint = strings.TrimSpace(hint)
	switch {
	case versionHintFullExp.MatchString(hint):
		return ParseServerHeaderToVersion("NexusIQ/" + hint)
	case versionHintReleaseExp.MatchString(hint):
		return GetStringAsInt32(hint)
	}
	return 0
}

// NexusIQ/1.201.0-02
var nxiqServerVersionExp = regexp.MustCompile(`^NEXUSIQ\/(?P<MAJOR>\d+)\.(?P<MINOR>\d+)\.(?P<PATCH>\d+)\-(?P<BUILD>\d+)$`)

func ParseServerHeaderToVersion(headerStr string) int32 {
	match := FindAllGroups(nxiqServerVersionExp, strings.ToUpper(headerStr))
	var iqVersion int32
	if match == nil {
		return iqVersion
	}
	for k, v := range match {
		switch k {
		case "MINOR":
			iqVersion = GetStringAsInt32(v)
		}
	}
	return iqVersion
}

func FindAllGroups(re *regexp.Regexp, s string) map[string]string {
	matches := re.FindStringSubmatch(s)
	subnames := re.SubexpNames()
	if matches == nil || len(matches) != len(subnames) {
		return nil
	}

	matchMap := map[string]string{}
	for i := 1; i < len(matches); i++ {
		matchMap[subnames[i]] = matches[i]
	}
	return matchMap
}

func GetStringAsInt32(s string) int32 {
	i64, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0
	}
	return int32(i64)
}
