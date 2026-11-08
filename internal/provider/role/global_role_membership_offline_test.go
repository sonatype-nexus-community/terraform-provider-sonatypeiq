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
package role_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	globalRoleId    = "global-role-id"
	nonGlobalRoleId = "application-role-id"
)

// membershipIq is the part of Sonatype IQ Server that grants roles: the global context only takes global roles,
// and an organization only takes non-global ones, as the real server does.
type membershipIq struct {
	*httptest.Server
	mu      sync.Mutex
	members map[string]bool // "roleId,type,name"
}

func newMembershipIq(t *testing.T) *membershipIq {
	t.Helper()
	f := &membershipIq{members: map[string]bool{}}

	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "NexusIQ/1.205.4-01")
		f.mu.Lock()
		defer f.mu.Unlock()
		f.route(w, r)
	}))
	t.Cleanup(f.Close)

	t.Setenv("IQ_SERVER_URL", f.URL)
	t.Setenv("IQ_SERVER_USERNAME", "admin")
	t.Setenv("IQ_SERVER_PASSWORD", "admin123")
	return f
}

func (f *membershipIq) route(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/ping":
	case path == "/api/v2/userTokens/currentUser/hasToken":
		http.Error(w, "no token", http.StatusNotFound)
	case path == "/api/v2/roleMemberships/global":
		f.listGlobal(w)
	case strings.HasPrefix(path, "/api/v2/roleMemberships/global/role/"):
		f.changeMember(w, r, strings.TrimPrefix(path, "/api/v2/roleMemberships/global/role/"), true)
	case strings.HasPrefix(path, "/api/v2/roleMemberships/organization/"):
		// /organization/{orgId}/role/{roleId}/{memberType}/{memberName}
		parts := strings.SplitN(strings.TrimPrefix(path, "/api/v2/roleMemberships/organization/"), "/role/", 2)
		if len(parts) == 2 {
			f.changeMember(w, r, parts[1], false)
			return
		}
		http.NotFound(w, r)
	default:
		http.NotFound(w, r)
	}
}

// changeMember grants (PUT) or revokes (DELETE) "{roleId}/{memberType}/{memberName}".
func (f *membershipIq) changeMember(w http.ResponseWriter, r *http.Request, spec string, global bool) {
	parts := strings.SplitN(spec, "/", 3)
	if len(parts) != 3 {
		http.NotFound(w, r)
		return
	}
	roleId, memberType, name := parts[0], parts[1], parts[2]

	if r.Method == http.MethodPut {
		if global && roleId != globalRoleId {
			http.Error(w, "Cannot map members to application role in global context.", http.StatusBadRequest)
			return
		}
		if !global && roleId == globalRoleId {
			http.Error(w, "Cannot map members to global role in context of organization.", http.StatusBadRequest)
			return
		}
		f.members[roleId+","+memberType+","+name] = true
	} else {
		delete(f.members, roleId+","+memberType+","+name)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (f *membershipIq) listGlobal(w http.ResponseWriter) {
	byRole := map[string][]map[string]any{}
	for key := range f.members {
		parts := strings.SplitN(key, ",", 3)
		byRole[parts[0]] = append(byRole[parts[0]], map[string]any{
			"ownerType":       "GLOBAL",
			"ownerId":         "ROOT_ORGANIZATION_ID",
			"type":            strings.ToUpper(parts[1]), // the server reports the type in upper case
			"userOrGroupName": parts[2],
		})
	}
	mappings := []map[string]any{}
	for roleId, members := range byRole {
		mappings = append(mappings, map[string]any{"roleId": roleId, "members": members})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"memberMappings": mappings})
}

func (f *membershipIq) has(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.members[key]
}

func (f *membershipIq) revokeBehindTerraformsBack(key string) func() {
	return func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		delete(f.members, key)
	}
}

func (f *membershipIq) checkGranted(key string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if !f.has(key) {
			return fmt.Errorf("%q has not been granted", key)
		}
		return nil
	}
}

func (f *membershipIq) checkRevoked(key string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if f.has(key) {
			return fmt.Errorf("%q has not been revoked", key)
		}
		return nil
	}
}

func globalMembershipConfig(body string) string {
	return utils_test.ProviderConfig + body
}

// #88: a global role (e.g. System Administrator) can be granted to a group, and revoked again.
func TestGlobalRoleMembershipResource_Group_Offline(t *testing.T) {
	f := newMembershipIq(t)
	const resourceName = "sonatypeiq_global_role_membership.test"
	const key = globalRoleId + ",group,My Admin Group"

	config := globalMembershipConfig(fmt.Sprintf(`
resource "sonatypeiq_global_role_membership" "test" {
  role_id    = %q
  group_name = "My Admin Group"
}`, globalRoleId))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		CheckDestroy:             f.checkRevoked(key),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", key),
					resource.TestCheckResourceAttr(resourceName, "role_id", globalRoleId),
					resource.TestCheckResourceAttr(resourceName, "group_name", "My Admin Group"),
					resource.TestCheckNoResourceAttr(resourceName, "user_name"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
					f.checkGranted(key),
				),
			},
			// No drift
			{Config: config, PlanOnly: true},
			// Import: <role id>,<type>,<name>
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateId:           key,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Revoked outside of Terraform: it is noticed and would be granted again
			{
				PreConfig:          f.revokeBehindTerraformsBack(key),
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestGlobalRoleMembershipResource_User_Offline(t *testing.T) {
	f := newMembershipIq(t)
	const key = globalRoleId + ",user,jane"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		CheckDestroy:             f.checkRevoked(key),
		Steps: []resource.TestStep{
			{
				Config: globalMembershipConfig(fmt.Sprintf(`
resource "sonatypeiq_global_role_membership" "test" {
  role_id   = %q
  user_name = "jane"
}`, globalRoleId)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sonatypeiq_global_role_membership.test", "user_name", "jane"),
					resource.TestCheckNoResourceAttr("sonatypeiq_global_role_membership.test", "group_name"),
					f.checkGranted(key),
				),
			},
		},
	})
}

// A role that is not global cannot be granted globally: say so, and say which resource to use.
func TestGlobalRoleMembershipResource_NonGlobalRole_Offline(t *testing.T) {
	newMembershipIq(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: globalMembershipConfig(fmt.Sprintf(`
resource "sonatypeiq_global_role_membership" "test" {
  role_id    = %q
  group_name = "Developers"
}`, nonGlobalRoleId)),
			ExpectError: regexp.MustCompile(`(?s)This\s+role\s+is\s+not\s+a\s+global\s+role.*sonatypeiq_organization_role_membership`),
		}},
	})
}

// The reported error: a global role on an organization now explains what to use instead of a bare 400.
func TestOrganizationRoleMembershipResource_GlobalRoleExplainsWhatToUse_Offline(t *testing.T) {
	newMembershipIq(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: globalMembershipConfig(fmt.Sprintf(`
resource "sonatypeiq_organization_role_membership" "test" {
  role_id         = %q
  organization_id = "root-organization-id"
  group_name      = "My Admin Group"
}`, globalRoleId)),
			ExpectError: regexp.MustCompile(`(?s)This\s+role\s+is\s+a\s+global\s+role.*sonatypeiq_global_role_membership`),
		}},
	})
}

func TestGlobalRoleMembershipResource_ExactlyOneMember_Offline(t *testing.T) {
	newMembershipIq(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: globalMembershipConfig(fmt.Sprintf(`
resource "sonatypeiq_global_role_membership" "test" {
  role_id    = %q
  user_name  = "jane"
  group_name = "Admins"
}`, globalRoleId)),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)user_name.*group_name`),
			},
			{
				Config: globalMembershipConfig(fmt.Sprintf(`
resource "sonatypeiq_global_role_membership" "test" {
  role_id = %q
}`, globalRoleId)),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)user_name.*group_name`),
			},
		},
	})
}
