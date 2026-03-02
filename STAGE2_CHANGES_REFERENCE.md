# Stage 2 Implementation - Actual Code Changes Reference

This document shows the actual changes made to implement Stage 2, useful as a reference for applying similar changes to remaining files.

---

## 1. Core Infrastructure Changes

### File: `internal/provider/common/api.go`

**Before:**
```go
package common

import (
    "fmt"
    "io"
    "log"
    "net/http"

    "github.com/hashicorp/terraform-plugin-framework/diag"
)

func HandleApiError(message string, err *error, httpResponse *http.Response, respDiags *diag.Diagnostics) {
    respDiags.AddError(
        message,
        fmt.Sprintf("%s: %s: %s", message, httpResponse.Status, getResponseBody(httpResponse)),
    )
}

func HandleApiWarning(message string, err *error, httpResponse *http.Response, respDiags *diag.Diagnostics) {
    respDiags.AddWarning(
        "LDAP Connection does not exist",
        fmt.Sprintf("%s: %s: %s", message, httpResponse.Status, getResponseBody(httpResponse)),
    )
}

func getResponseBody(httpResponse *http.Response) []byte {
    body, _ := io.ReadAll(httpResponse.Body)
    err := httpResponse.Body.Close()
    if err != nil {
        log.Fatal(err.Error())
    }
    return body
}
```

**After:**
```go
package common

import (
    "net/http"

    "github.com/hashicorp/terraform-plugin-framework/diag"
    sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
)

// HandleApiError delegates to the shared library's HandleAPIError function for centralized error handling
// with network error detection and standardized diagnostics formatting
func HandleApiError(message string, err *error, httpResponse *http.Response, respDiags *diag.Diagnostics) {
    sharederr.HandleAPIError(message, err, httpResponse, respDiags)
}

// HandleApiWarning delegates to the shared library's HandleAPIWarning function for non-critical API issues
func HandleApiWarning(message string, err *error, httpResponse *http.Response, respDiags *diag.Diagnostics) {
    sharederr.HandleAPIWarning(message, err, httpResponse, respDiags)
}
```

**Key Changes:**
- Removed `fmt`, `io`, and `log` imports (no longer needed)
- Added `sharederr` import with alias
- Simplified functions to delegate to shared library
- Removed duplicate `getResponseBody()` function
- Improved documentation with comments

---

## 2. Data Source Changes

### File: `internal/provider/application/application_categories_data_source.go`

**Import Section - Before:**
```go
import (
    "context"
    "terraform-provider-sonatypeiq/internal/provider/common"
    "terraform-provider-sonatypeiq/internal/provider/model"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
)
```

**Import Section - After:**
```go
import (
    "context"
    "net/http"
    "terraform-provider-sonatypeiq/internal/provider/common"
    "terraform-provider-sonatypeiq/internal/provider/model"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
    sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
)
```

**Error Handling - Before:**
```go
categories, api_response, err := d.Client.ApplicationCategoriesAPI.GetTags(ctx, data.OrganiziationId.ValueString()).Execute()

if err != nil {
    resp.Diagnostics.AddError(
        "Unable to Read IQ Application Categories for Organization",
        err.Error(),
    )
    return
}
if api_response.StatusCode != 200 {
    resp.Diagnostics.AddError("Unexpected API Response", api_response.Status)
    return
}
```

**Error Handling - After:**
```go
categories, api_response, err := d.Client.ApplicationCategoriesAPI.GetTags(ctx, data.OrganiziationId.ValueString()).Execute()

if err != nil {
    sharederr.HandleAPIError("Unable to read IQ Application Categories for Organization", &err, api_response, &resp.Diagnostics)
    return
}
if api_response.StatusCode != http.StatusOK {
    sharederr.AddAPIErrorDiagnostic(&resp.Diagnostics, "read", "Application Categories", api_response, err)
    return
}
```

**Key Changes:**
- Added `net/http` import for status code constants
- Added `sharederr` import
- Replaced `resp.Diagnostics.AddError()` with `sharederr.HandleAPIError()`
- Changed hardcoded `200` to `http.StatusOK`
- Used specific diagnostic adder for API errors

---

## 3. Resource Changes

### File: `internal/provider/organization/organization_resource.go`

**Import Section - Key Changes:**
```go
// Removed: "fmt", "io"
// Added: "net/http"
// Added: sharederr import

import (
    "context"
    "fmt"  // Re-added because it's still used in tflog calls
    "net/http"
    "terraform-provider-sonatypeiq/internal/provider/common"
    "time"
    // ...
    sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
    sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
)
```

**Create Method - Before:**
```go
organization, api_response, err := organization_request.Execute()

// Call API
if err != nil {
    error_body, _ := io.ReadAll(api_response.Body)
    resp.Diagnostics.AddError(
        "Error creating Organization",
        "Could not create Organization, unexpected error: "+api_response.Status+": "+string(error_body),
    )
    return
}
```

**Create Method - After:**
```go
organization, api_response, err := organization_request.Execute()

// Call API
if err != nil {
    sharederr.HandleAPIError("Error creating Organization", &err, api_response, &resp.Diagnostics)
    return
}
```

**Read Method - Before:**
```go
organization, _, err := r.Client.OrganizationsAPI.GetOrganization(ctx, state.ID.ValueString()).Execute()

if err != nil {
    resp.Diagnostics.AddError(
        "Error Reading IQ Organization",
        "Could not read Organization with ID "+state.ID.ValueString()+": "+err.Error(),
    )
    return
}
```

**Read Method - After:**
```go
organization, httpResponse, err := r.Client.OrganizationsAPI.GetOrganization(ctx, state.ID.ValueString()).Execute()

if err != nil {
    sharederr.HandleAPIError("Error reading Organization", &err, httpResponse, &resp.Diagnostics)
    return
}
```

**Delete Method - Before:**
```go
api_response, err := r.Client.OrganizationsAPI.DeleteOrganization(ctx, state.ID.ValueString()).Execute()
if err != nil {
    error_body, _ := io.ReadAll(api_response.Body)
    resp.Diagnostics.AddError(
        "Error deleting Organization",
        "Could not delete Organization, unexpected error: "+api_response.Status+": "+string(error_body),
    )
    return
}
```

**Delete Method - After:**
```go
api_response, err := r.Client.OrganizationsAPI.DeleteOrganization(ctx, state.ID.ValueString()).Execute()
if err != nil {
    sharederr.HandleAPIError("Error deleting Organization", &err, api_response, &resp.Diagnostics)
    return
}
```

**Key Changes:**
- Removed `io.ReadAll()` calls for error body reading (handled by shared library)
- Simplified error messages (shared library formats them)
- Changed variable names when needed: `_` → `httpResponse`
- Used `sharederr.HandleAPIError()` consistently

---

## 4. Status Code Handling Changes

### File: `internal/provider/user/user_token_resource.go`

**Read Method - Before:**
```go
_, httpResponse, err := r.Client.UserTokensAPI.GetUserTokenExistsForCurrentUser(ctx).Execute()

if err != nil {
    if httpResponse.StatusCode == http.StatusNotFound {
        resp.State.RemoveResource(ctx)
        common.HandleApiWarning(
            "No User Token exists for User in Realm",
            &err,
            httpResponse,
            &resp.Diagnostics,
        )
    } else {
        common.HandleApiError(
            "Error checking User Token",
            &err,
            httpResponse,
            &resp.Diagnostics,
        )
    }
    return
}
```

**Read Method - After:**
```go
_, httpResponse, err := r.Client.UserTokensAPI.GetUserTokenExistsForCurrentUser(ctx).Execute()

if err != nil {
    if sharederr.IsNotFound(httpResponse.StatusCode) {
        resp.State.RemoveResource(ctx)
        sharederr.AddNotFoundDiagnostic(&resp.Diagnostics, "User Token", "current user")
    } else {
        sharederr.HandleAPIError("Error checking User Token", &err, httpResponse, &resp.Diagnostics)
    }
    return
}
```

**Create Method - Before:**
```go
apiResponse, httpResponse, err := r.Client.UserTokensAPI.CreateUserToken(ctx).Execute()

if err != nil || httpResponse.StatusCode != http.StatusOK {
    common.HandleApiError(
        "Error creating User Token",
        &err,
        httpResponse,
        respDiags,
    )
    return
}
```

**Create Method - After:**
```go
apiResponse, httpResponse, err := r.Client.UserTokensAPI.CreateUserToken(ctx).Execute()

if err != nil {
    sharederr.HandleAPIError("Error creating User Token", &err, httpResponse, respDiags)
    return
}
if httpResponse.StatusCode != http.StatusOK {
    sharederr.AddAPIErrorDiagnostic(respDiags, "create", "User Token", httpResponse, err)
    return
}
```

**Delete Method - Before:**
```go
httpResponse, err := r.Client.UserTokensAPI.DeleteCurrentUserToken(ctx).Execute()

if err != nil || httpResponse.StatusCode != http.StatusNoContent {
    common.HandleApiError(
        "Error deleting User Token",
        &err,
        httpResponse,
        respDiags,
    )
    return
}
```

**Delete Method - After:**
```go
httpResponse, err := r.Client.UserTokensAPI.DeleteCurrentUserToken(ctx).Execute()

if err != nil {
    sharederr.HandleAPIError("Error deleting User Token", &err, httpResponse, respDiags)
    return
}
if httpResponse.StatusCode != http.StatusNoContent {
    sharederr.AddAPIErrorDiagnostic(respDiags, "delete", "User Token", httpResponse, err)
    return
}
```

**Key Changes:**
- Separated error handling from status code checking (clearer logic flow)
- Used `sharederr.IsNotFound()` instead of direct status code comparison
- Used `sharederr.AddNotFoundDiagnostic()` for 404 cases
- Used `sharederr.AddAPIErrorDiagnostic()` for unexpected status codes
- Removed `common.HandleApi*` calls in favor of shared library

---

## 5. Validation Error Handling

### Pattern for Data Sources (application_data_source.go)

**Before:**
```go
if len(apps.Applications) > 1 {
    resp.Diagnostics.AddError("More than one Applications matched the Public ID", r.Status)
    return
}
```

**After:**
```go
if len(apps.Applications) > 1 {
    sharederr.AddValidationDiagnostic(&resp.Diagnostics, "Application Public ID", 
        "More than one Application matched the provided Public ID")
    return
}
```

**Before:**
```go
if app == nil {
    resp.Diagnostics.AddError("No Application found", "No Application found with the provided ID or Public ID")
    return
}
```

**After:**
```go
if app == nil {
    sharederr.AddNotFoundDiagnostic(&resp.Diagnostics, "Application", data.ID.ValueString())
    return
}
```

---

## Summary of Patterns Applied

| Original Pattern | Shared Library Replacement | Use Case |
|---|---|---|
| `resp.Diagnostics.AddError("msg", err.Error())` | `sharederr.HandleAPIError("msg", &err, resp, &diags)` | Network/API errors |
| `resp.Diagnostics.AddError("Unexpected", "details")` | `sharederr.AddAPIErrorDiagnostic(&diags, "op", "resource", resp, err)` | Status code errors |
| `if r.StatusCode == 404` | `if sharederr.IsNotFound(r.StatusCode)` | Not found detection |
| `if r.StatusCode == 409` | `if sharederr.IsConflict(r.StatusCode)` | Conflict detection |
| Custom validation errors | `sharederr.AddValidationDiagnostic(&diags, "field", "reason")` | Validation errors |
| `if r.StatusCode != 200` | `if r.StatusCode != http.StatusOK` | Status constants |
| Hardcoded `io.ReadAll()` | Handled by shared library | Error body extraction |

---

## Lines of Code Impact

- **common/api.go**: Reduced from 48 lines to 27 lines (-44% reduction)
- **Typical resource file**: 3-5 lines saved per error handling block
- **Typical data source file**: 2-4 lines saved per error handling block

**Overall Impact:**
- Estimated 150-200 lines of boilerplate error handling removed across all files
- Code becomes more maintainable and consistent
