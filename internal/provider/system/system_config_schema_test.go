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
package system_test

import (
	"context"
	"os"
	"strings"
	"terraform-provider-sonatypeiq/internal/provider/model"
	"terraform-provider-sonatypeiq/internal/provider/system"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Every registered property must be an attribute of both the resource and the data source, and the
// attribute must say when the property was introduced, so the generated documentation never drifts.
func TestSystemConfigSchemasMatchTheRegistry(t *testing.T) {
	ctx := context.Background()

	resourceResp := &resource.SchemaResponse{}
	system.NewSystemConfigResource().Schema(ctx, resource.SchemaRequest{}, resourceResp)

	dataSourceResp := &datasource.SchemaResponse{}
	system.SystemConfigDataSource().Schema(ctx, datasource.SchemaRequest{}, dataSourceResp)

	for _, p := range model.SystemConfigProperties {
		resourceAttr, ok := resourceResp.Schema.Attributes[p.Attribute]
		if !ok {
			t.Errorf("resource schema has no attribute %q", p.Attribute)
			continue
		}
		if p.Attribute == "webhook_secret_passphrase" {
			if _, inDataSource := dataSourceResp.Schema.Attributes[p.Attribute]; inDataSource {
				t.Errorf("the data source must not expose the webhook secret")
			}
		} else if _, ok := dataSourceResp.Schema.Attributes[p.Attribute]; !ok {
			t.Errorf("data source schema has no attribute %q", p.Attribute)
		}

		if p.Attribute == "base_url" || p.Attribute == "force_base_url" {
			continue
		}
		if !resourceAttr.IsOptional() {
			t.Errorf("%s must be optional so that it is only managed when set", p.Attribute)
		}
		if want := p.AvailabilityNote(); !containsString(resourceAttr.GetDescription(), want) {
			t.Errorf("%s description does not state its availability %q", p.Attribute, want)
		}
	}
}

func containsString(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// The documentation templates carry the availability table. It must match the registry exactly, so the
// published documentation can never disagree with the version gating. Regenerate with `go generate` in tools.
func TestSystemConfigDocumentationTemplatesMatchTheRegistry(t *testing.T) {
	table := model.SystemConfigAvailabilityMarkdown()
	for _, file := range []string{
		"../../../templates/resources/system_config.md.tmpl",
		"../../../templates/data-sources/system_config.md.tmpl",
	} {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("reading %s: %v", file, err)
		}
		if !strings.Contains(string(content), table) {
			t.Errorf("%s does not contain the generated availability table - update it from model.SystemConfigAvailabilityMarkdown()", file)
		}
	}
}
