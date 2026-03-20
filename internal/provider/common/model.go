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

func (p *SonatypeDataSourceData) CheckWritableAndGetVersion(ctx context.Context, respDiags *diag.Diagnostics) {
	_, httpResponse, err := p.Client.UsersAPI.Get1(
		context.WithValue(ctx, sonatypeiq.ContextBasicAuth, p.Auth),
		p.Auth.UserName,
	).Execute()

	if err != nil {
		sharederr.HandleAPIError(
			"Sonatype IQ Server cannot be accessed.",
			&err,
			httpResponse,
			respDiags,
		)
		return
	}

	if httpResponse.StatusCode == http.StatusOK {
		p.IqVersion = ParseServerHeaderToVersion(httpResponse.Header.Get("server"))
		tflog.Debug(ctx, fmt.Sprintf("Server Header: %v", p.IqVersion))
	}

	tflog.Info(ctx, fmt.Sprintf("Determined Sonatype IQ Server to be version %v", p.IqVersion))
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
