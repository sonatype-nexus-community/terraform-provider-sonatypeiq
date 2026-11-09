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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	utils_test "terraform-provider-sonatypeiq/internal/provider/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// fakeUserTokenConfigIqServer is a minimal fake IQ Server for user token configuration
type fakeUserTokenConfigIqServer struct {
	*httptest.Server
	mu             sync.Mutex
	expirationDays *int32
}

func newFakeUserTokenConfigIqServer(t *testing.T, release int) *fakeUserTokenConfigIqServer {
	t.Helper()
	f := &fakeUserTokenConfigIqServer{}

	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", fmt.Sprintf("NexusIQ/1.%d.0-01", release))
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

func (f *fakeUserTokenConfigIqServer) route(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/ping":
		return
	case "/api/v2/userTokens/currentUser/hasToken":
		http.Error(w, "no token", http.StatusNotFound)
		return
	case "/api/v2/users/admin":
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"username":"admin"}`))
		return
	case "/api/v2/config/userTokens":
		f.handleUserTokenConfig(w, r)
		return
	}
	http.NotFound(w, r)
}

func (f *fakeUserTokenConfigIqServer) handleUserTokenConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		f.handleGetConfig(w, r)
	case http.MethodPut:
		f.handlePutConfig(w, r)
	case http.MethodDelete:
		f.handleDeleteConfig(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (f *fakeUserTokenConfigIqServer) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{}
	if f.expirationDays != nil {
		resp["userTokenDefaultExpirationDays"] = *f.expirationDays
	} else {
		resp["userTokenDefaultExpirationDays"] = nil
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (f *fakeUserTokenConfigIqServer) handlePutConfig(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if val, ok := body["userTokenDefaultExpirationDays"]; ok {
		if val == nil {
			f.expirationDays = nil
		} else if floatVal, ok := val.(float64); ok {
			intVal := int32(floatVal)
			f.expirationDays = &intVal
		}
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{}
	if f.expirationDays != nil {
		resp["userTokenDefaultExpirationDays"] = *f.expirationDays
	} else {
		resp["userTokenDefaultExpirationDays"] = nil
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (f *fakeUserTokenConfigIqServer) handleDeleteConfig(w http.ResponseWriter, r *http.Request) {
	// Check for property parameter
	properties := r.URL.Query()["property"]
	if len(properties) == 0 {
		http.Error(w, "No properties specified for reset", http.StatusBadRequest)
		return
	}

	// Reset the specified properties
	for _, prop := range properties {
		if prop == "userTokenDefaultExpirationDays" {
			f.expirationDays = nil
		}
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{"userTokenDefaultExpirationDays": nil}
	_ = json.NewEncoder(w).Encode(resp)
}

func (f *fakeUserTokenConfigIqServer) checkStored(expirationDays *int32) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.expirationDays == nil && expirationDays == nil {
			return nil
		}
		if f.expirationDays == nil || expirationDays == nil {
			return fmt.Errorf("server has expirationDays = %v, want %v", f.expirationDays, expirationDays)
		}
		if *f.expirationDays != *expirationDays {
			return fmt.Errorf("server has expirationDays = %d, want %d", *f.expirationDays, *expirationDays)
		}
		return nil
	}
}

func userTokenConfig(expirationDays int) string {
	return fmt.Sprintf(utils_test.ProviderConfig+`
resource "sonatypeiq_user_token_configuration" "test" {
  user_token_default_expiration_days = %d
}`, expirationDays)
}

func TestUserTokenConfigurationResource_CreateAndUpdate_Offline(t *testing.T) {
	f := newFakeUserTokenConfigIqServer(t, 199)
	resourceName := "sonatypeiq_user_token_configuration.test"

	ninetyDays := int32(90)
	fortyFiveDays := int32(45)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: userTokenConfig(90),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user_token_default_expiration_days", "90"),
					f.checkStored(&ninetyDays),
				),
			},
			{
				Config:   userTokenConfig(90),
				PlanOnly: true,
			},
			{
				Config: userTokenConfig(45),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user_token_default_expiration_days", "45"),
					f.checkStored(&fortyFiveDays),
				),
			},
		},
	})
}

func TestUserTokenConfigurationResource_Validation_Offline(t *testing.T) {
	newFakeUserTokenConfigIqServer(t, 199)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      userTokenConfig(0),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)user_token_default_expiration_days.*(between|1 and 365)`),
			},
			{
				Config:      userTokenConfig(366),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)user_token_default_expiration_days.*(between|1 and 365)`),
			},
		},
	})
}

func TestUserTokenConfigurationResource_DeleteResetsToDefault_Offline(t *testing.T) {
	f := newFakeUserTokenConfigIqServer(t, 199)
	resourceName := "sonatypeiq_user_token_configuration.test"

	ninetyDays := int32(90)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: utils_test.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: userTokenConfig(90),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user_token_default_expiration_days", "90"),
					f.checkStored(&ninetyDays),
				),
			},
			{
				Config: utils_test.ProviderConfig,
				Check:  f.checkStored(nil),
			},
		},
	})
}
