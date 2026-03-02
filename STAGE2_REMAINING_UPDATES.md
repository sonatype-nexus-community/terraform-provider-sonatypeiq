# Stage 2 Error Handling - Remaining File Updates Guide

This document provides a systematic approach for completing the remaining 17 files that need error handling refactoring.

## Quick Reference: Error Handling Patterns

### Network/API Error
```go
if err != nil {
    sharederr.HandleAPIError("describe operation", &err, httpResponse, &resp.Diagnostics)
    return
}
```

### HTTP Status Code Check
```go
if httpResponse.StatusCode != http.StatusOK {
    sharederr.AddAPIErrorDiagnostic(&resp.Diagnostics, "operation", "resource", httpResponse, err)
    return
}
```

### Not Found (404)
```go
if sharederr.IsNotFound(httpResponse.StatusCode) {
    sharederr.AddNotFoundDiagnostic(&resp.Diagnostics, "Resource Type", resourceID)
    return
}
```

### Conflict (409)
```go
if sharederr.IsConflict(httpResponse.StatusCode) {
    sharederr.AddConflictDiagnostic(&resp.Diagnostics, "Resource Type", "reason")
    return
}
```

### Unauthorized (401)
```go
if sharederr.IsUnauthorized(httpResponse.StatusCode) {
    sharederr.AddUnauthorizedDiagnostic(&resp.Diagnostics, "operation")
    return
}
```

### Forbidden (403)
```go
if sharederr.IsForbidden(httpResponse.StatusCode) {
    sharederr.AddForbiddenDiagnostic(&resp.Diagnostics, "operation")
    return
}
```

### Validation Error
```go
sharederr.AddValidationDiagnostic(&resp.Diagnostics, "field name", "validation reason")
```

---

## Files by Priority

### Priority 1: Large Resource Files (5 files - ~100+ error handling lines)

#### 1. `internal/provider/application/application_resource.go`
**Current Pattern:** Uses custom `resp.Diagnostics.AddError()`  
**Required Changes:** 
- Replace in Create, Read, Update, Delete methods
- Import: `net/http` (add if missing)
- Typical patterns: Generic API errors in CRUD

```bash
# Search for error handling
grep -n "resp.Diagnostics.AddError\|api_response\|httpResponse" application_resource.go
```

#### 2. `internal/provider/scm/source_control_resource.go`
**Current Pattern:** Uses `common.HandleApiError()` and `common.HandleApiWarning()`  
**Required Changes:**
- Replace `common.HandleApiError()` with `sharederr.HandleAPIError()`
- Already has error handling, just needs library swap

#### 3. `internal/provider/system/config_crowd_resource.go`
**Current Pattern:** Uses `common.HandleApiError()` and `common.HandleApiWarning()`  
**Required Changes:**
- Replace deprecated common functions with shared library

#### 4. `internal/provider/system/config_saml_data_source.go`
**Current Pattern:** Uses `common.HandleApiError()`  
**Required Changes:**
- Single-point updates in data source Read method

#### 5. `internal/provider/organization/organization_category_resource.go`
**Current Pattern:** Uses `common.HandleApiWarning()` and `common.HandleApiError()`  
**Required Changes:**
- Complex: Has both error and warning cases
- Particular attention to warning cases (e.g., LDAP)

---

### Priority 2: Medium Resource Files (7 files - 20-50 error handling lines)

#### 6. `internal/provider/application/application_role_membership_resource.go`
```bash
grep -c "resp.Diagnostics.Add" application_role_membership_resource.go
```

#### 7. `internal/provider/organization/organization_data_source.go`

#### 8. `internal/provider/organization/organization_role_membership_resource.go`

#### 9. `internal/provider/system/config_license_resource.go`

#### 10. `internal/provider/system/config_mail_resource.go`

#### 11. `internal/provider/system/config_proxy_server_resource.go`

#### 12. `internal/provider/system/system_config_resource.go`

---

### Priority 3: Smaller Data Sources & Resources (5 files - <20 error lines)

#### 13. `internal/provider/application/applications_data_source.go`

#### 14. `internal/provider/organization/organizations_data_source.go`

#### 15. `internal/provider/role/role_data_source.go`

#### 16. `internal/provider/system/system_config_data_source.go`

#### 17. `internal/provider/user/user_resource.go`

---

## Manual Update Process for Each File

### Step 1: Analyze Current Error Handling
```bash
grep -n "resp.Diagnostics.AddError\|resp.Diagnostics.AddWarning\|common.HandleApi" <file>
```

### Step 2: Count Error Handling Instances
```bash
grep -c "resp.Diagnostics.Add\|common.HandleApi" <file>
```

### Step 3: View Error Handling Context
```bash
grep -B 3 -A 3 "resp.Diagnostics.AddError" <file> | head -50
```

### Step 4: Update Error Handling
For each error handling block:
1. Identify error type (network error, status code check, validation, etc.)
2. Replace with appropriate `sharederr` function
3. Ensure status code comparisons use `http.Status*` constants
4. Verify imports are correct

### Step 5: Compile and Check
```bash
go build -v . 2>&1 | grep -E "error:|undefined:"
```

### Step 6: Verify No New Errors
```bash
go vet ./...
```

---

## Common Error Handling Patterns Found

### Pattern A: Generic Error from API Call
```go
// BEFORE
if err != nil {
    resp.Diagnostics.AddError(
        "Error doing X",
        err.Error(),
    )
    return
}

// AFTER
if err != nil {
    sharederr.HandleAPIError("Error doing X", &err, httpResponse, &resp.Diagnostics)
    return
}
```

### Pattern B: Error with Response Body
```go
// BEFORE
if err != nil {
    error_body, _ := io.ReadAll(api_response.Body)
    resp.Diagnostics.AddError(
        "Error creating X",
        "Could not create X: "+api_response.Status+": "+string(error_body),
    )
    return
}

// AFTER
if err != nil {
    sharederr.HandleAPIError("Error creating X", &err, api_response, &resp.Diagnostics)
    return
}
```

### Pattern C: Status Code Check
```go
// BEFORE
if api_response.StatusCode != 200 {
    resp.Diagnostics.AddError("Unexpected API Response", api_response.Status)
    return
}

// AFTER
if api_response.StatusCode != http.StatusOK {
    sharederr.AddAPIErrorDiagnostic(&resp.Diagnostics, "operation", "resource", api_response, err)
    return
}
```

### Pattern D: Specific Status Code Handling
```go
// BEFORE
if api_response.StatusCode == 404 {
    resp.Diagnostics.AddError("Not Found", "Resource not found with ID: "+id)
    return
}

// AFTER
if sharederr.IsNotFound(api_response.StatusCode) {
    sharederr.AddNotFoundDiagnostic(&resp.Diagnostics, "Resource Type", id)
    return
}
```

### Pattern E: Warning Messages
```go
// BEFORE
resp.Diagnostics.AddWarning(
    "LDAP Connection does not exist",
    err.Error(),
)

// AFTER
// Warnings can be added directly or converted to info-level errors
// depending on context. For connection issues:
sharederr.AddValidationDiagnostic(&resp.Diagnostics, "LDAP Configuration", 
    "LDAP connection not configured or could not be verified")
```

---

## Testing After Updates

### 1. Build Check
```bash
go build -v .
```

### 2. Unit Tests
```bash
go test -v ./...
```

### 3. Lint Check
```bash
golangci-lint run ./...
```

### 4. Format Check
```bash
go fmt ./...
```

---

## Tracking Checklist

Use this to track completion:

- [ ] application_resource.go
- [ ] application_role_membership_resource.go
- [ ] applications_data_source.go
- [ ] organization_category_resource.go
- [ ] organization_data_source.go
- [ ] organization_role_membership_resource.go
- [ ] organizations_data_source.go
- [ ] role_data_source.go
- [ ] scm/source_control_resource.go
- [ ] system/config_crowd_resource.go
- [ ] system/config_license_resource.go
- [ ] system/config_mail_resource.go
- [ ] system/config_proxy_server_resource.go
- [ ] system/config_saml_data_source.go
- [ ] system/system_config_data_source.go
- [ ] system/system_config_resource.go
- [ ] user/user_resource.go

---

## Helpful Commands

### Find all error handling in a file
```bash
grep -n "Diagnostics.Add\|HandleApi" <file>
```

### See context around errors
```bash
grep -B 2 -A 5 "resp.Diagnostics.AddError" <file>
```

### Count error instances
```bash
grep -c "resp.Diagnostics.Add\|common.HandleApi" <file>
```

### Find files with most errors to handle
```bash
for f in internal/provider/**/*.go; do echo "$f: $(grep -c 'resp.Diagnostics.Add\|common.HandleApi' $f)"; done | sort -t: -k2 -rn
```

---

## Notes

1. **Import Management**: `net/http` import may be needed for `http.Status*` constants
2. **Error Variables**: Change from `_` discards to actual variable names when needed
3. **Deprecation**: Once all files are updated, `common.HandleApiError` and `common.HandleApiWarning` in common/api.go can be deprecated
4. **Testing**: Each file update should be tested to ensure error messages remain user-friendly
5. **Backwards Compatibility**: These changes are internal-only with no user API changes
