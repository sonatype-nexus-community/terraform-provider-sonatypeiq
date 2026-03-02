# Stage 4: Utility Functions & Type Conversions Implementation

## Status: IN PROGRESS (Phase 1 - Example Complete)

**Progress:** 1/21 files complete (5%)

### Overview
This stage replaces manual type conversions and utility functions with standardized implementations from `github.com/sonatype-nexus-community/terraform-provider-shared/util`.

Current shared library version: **v0.9.3** ✓ (already in go.mod)

---

## Analysis Summary

### Current Usage Patterns
- **118 occurrences** of `types.StringValue()`, `types.BoolValue()`, `types.Int64Value()`
- **72 occurrences** of `.ValueStringPointer()`, `.ValueBoolPointer()`, `.ValueInt64Pointer()`
- No custom conversion utility functions found in codebase (already using Terraform framework types directly)

### Key Finding
The codebase is already using Terraform plugin framework types correctly. The goal is to simplify and standardize conversion patterns, particularly when working with API client responses that return pointers.

---

## Shared Library Utilities Available

### Type Conversions (util/conversion.go)

#### Pointer Creation
```go
util.StringToPtr(s string) *string
util.BoolToPtr(b bool) *bool
util.Int64ToPtr(i int64) *int64
util.Int32ToPtr(i int32) *int32
```

#### Pointer to Terraform Types
```go
util.StringPtrToValue(s *string) types.String          // Handles nil
util.BoolPtrToValue(b *bool) types.Bool                 // Handles nil
util.Int64PtrToValue(i *int64) types.Int64              // Handles nil
util.Int32PtrToValue(i *int32) types.Int64              // Converts int32 → int64
```

#### Safe Unwrapping (Returns zero value if nil)
```go
util.SafeString(s *string) string      // "" if nil
util.SafeBool(b *bool) bool            // false if nil
util.SafeInt64(i *int64) int64         // 0 if nil
util.SafeInt32(i *int32) int32         // 0 if nil
```

#### String Conversions
```go
util.StringToInt64(s string) (int64, error)
util.StringToInt32(s string) (int32, error)
util.StringToFloat(s string) (float64, error)
util.StringToBool(s string) (bool, error)

util.Int64ToString(i int64) string
util.Int32ToString(i int32) string
util.FloatToString(f float64, precision int) string
util.BoolToString(b bool) string
```

#### Slice Conversions
```go
util.StringSliceToValue(ss []string) types.Set
util.StringPtrSliceToValue(ss []*string) types.Set
util.Int64SliceToValue(ii []int64) types.Set
```

#### Conditional Helpers
```go
util.NilIfEmpty(s string) *string      // nil if empty, else ptr to string
util.EmptyIfNil(s *string) string      // "" if nil, else dereference
```

---

## Implementation Tasks

### Task 1: Setup Import Alias
**Status:** COMPLETE ✓

Create a common imports file for consistency across the codebase.

**File:** `internal/provider/common/shared_imports.go`

```go
package common

import (
	sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
	sharedutil "github.com/sonatype-nexus-community/terraform-provider-shared/util"
	sharedvalidators "github.com/sonatype-nexus-community/terraform-provider-shared/validators"
)

// These packages are re-exported for convenience across the provider
// Import as:
//   sharederr "terraform-provider-sonatypeiq/internal/provider/common"
//   sharedutil "terraform-provider-sonatypeiq/internal/provider/common"
// Then access: sharederr.HandleAPIError(), sharedutil.StringPtrToValue(), etc.
```

### Task 2: Replace Pointer to Terraform Type Conversions
**Status:** IN PROGRESS (1/21 files complete)

Replace API response pointer handling with `util.StringPtrToValue()`, `util.BoolPtrToValue()`, etc.

**Pattern to Replace:**
```go
// BEFORE: Manual nil check
if org.Name != nil {
    state.Name = types.StringValue(*org.Name)
} else {
    state.Name = types.StringNull()
}

// AFTER: Using shared util
state.Name = sharedutil.StringPtrToValue(org.Name)
```

**Files to Update:** (118 occurrences across resources)
- `internal/provider/application/*` - 5 files
- `internal/provider/organization/*` - 5 files (1 DONE ✓)
- `internal/provider/role/*` - 1 file
- `internal/provider/user/*` - 2 files
- `internal/provider/system/*` - 7 files
- `internal/provider/scm/*` - 1 file

**Reference Implementation:**
1. ✓ `internal/provider/organization/organization_resource.go` (COMPLETED)

**Next Files (in order):**
2. `internal/provider/organization/organization_data_source.go`
3. `internal/provider/user/user_resource.go`
4. `internal/provider/application/application_resource.go`

See `STAGE4_FILES_TO_UPDATE.md` for detailed priority list.

### Task 3: Replace Pointer Creation Patterns
**Status:** TODO

Replace manual pointer creation with `util.*ToPtr()` functions.

**Pattern to Replace:**
```go
// BEFORE: Using &
createReq := &sonatypeiq.ApiOrganizationDTO{
    Name: &plan.Name.ValueString(),
}

// AFTER: Using shared util
createReq := &sonatypeiq.ApiOrganizationDTO{
    Name: sharedutil.StringToPtr(plan.Name.ValueString()),
}
```

**Note:** This is primarily for creating request objects from Terraform state.

### Task 4: Replace Safe Unwrapping
**Status:** TODO

Replace patterns where nil values need safe unwrapping with `util.Safe*()` functions.

**Pattern to Replace:**
```go
// BEFORE: Manual nil check
var name string
if organization.Name != nil {
    name = *organization.Name
} else {
    name = ""
}

// AFTER: Using shared util
name := sharedutil.SafeString(organization.Name)
```

### Task 5: Update Helper Methods
**Status:** TODO

Check `internal/provider/model/` for any custom conversion helpers and replace if duplicate to shared library.

**Files to Review:**
- `internal/provider/model/*.go`

---

## Implementation Order

1. **Phase 1 (High Impact - Start Here)**
   - [ ] Add import alias file
   - [ ] Update `organization_resource.go` (most straightforward)
   - [ ] Update `organization_data_source.go`
   - [ ] Update `user_resource.go`

2. **Phase 2 (Medium Impact)**
   - [ ] Update all remaining resource files
   - [ ] Update all data source files

3. **Phase 3 (Low Impact)**
   - [ ] Review and update model files
   - [ ] Review and update any helper utilities

4. **Phase 4 (Validation)**
   - [ ] Run `go test ./...`
   - [ ] Run `go fmt ./...`
   - [ ] Run `golangci-lint run ./...`

---

## Key Patterns in Codebase

### Pattern 1: Reading API Response → Terraform State
```go
// Organization read operation example
organization, _, err := r.Client.OrganizationsAPI.GetOrganization(ctx, stateID).Execute()
if err != nil {
    sharederr.HandleAPIError("Error reading Organization", &err, httpResponse, &resp.Diagnostics)
    return
}

// CURRENT: Direct pointer dereference
state.ID = types.StringValue(*organization.Id)
state.Name = types.StringValue(*organization.Name)

// PROPOSED: Using shared util
state.ID = sharedutil.StringPtrToValue(organization.Id)
state.Name = sharedutil.StringPtrToValue(organization.Name)
```

### Pattern 2: Terraform State → API Request
```go
// CURRENT: Getting pointer from Terraform value
createReq := &sonatypeiq.ApiOrganizationDTO{
    Name: &plan.Name.ValueString(),  // Creates new pointer
}

// PROPOSED: Using shared util
nameValue := plan.Name.ValueString()
createReq := &sonatypeiq.ApiOrganizationDTO{
    Name: sharedutil.StringToPtr(nameValue),
}
```

### Pattern 3: Conditional Assignment
```go
// CURRENT: Manual nil check
if organization.ParentOrganizationId != nil {
    state.ParentOrgID = types.StringValue(*organization.ParentOrganizationId)
} else {
    state.ParentOrgID = types.StringNull()
}

// PROPOSED: Using shared util
state.ParentOrgID = sharedutil.StringPtrToValue(organization.ParentOrganizationId)
```

---

## Testing Strategy

### Unit Testing
```bash
go test ./internal/provider/organization -v
go test ./internal/provider/user -v
go test ./internal/provider/system -v
```

### Full Test Suite
```bash
go test -v -cover ./...
```

### Manual Acceptance Testing
```bash
# Build provider
go build -o terraform-provider-sonatypeiq

# Test with local Terraform (examples in examples/ directory)
terraform init -upgrade
terraform plan
terraform apply
```

---

## Acceptance Criteria

- [ ] All type conversions use `util.*PtrToValue()` helpers where applicable
- [ ] All pointer creation uses `util.*ToPtr()` helpers
- [ ] All nil-safe unwrapping uses `util.Safe*()` helpers
- [ ] No behavioral changes (identical test results)
- [ ] `go test ./...` passes
- [ ] `golangci-lint run ./...` shows no new errors
- [ ] Code builds successfully
- [ ] Manual testing confirms all CRUD operations work

---

## Risk Assessment

**Risk Level:** LOW

**Rationale:**
- Purely internal refactoring with no API changes
- Shared utilities have well-defined contracts
- Test coverage validates behavior preservation
- Changes are mechanical/straightforward (search and replace)

**Mitigation:**
- Run full test suite after each file update
- Keep git commits atomic (one resource at a time)
- Test manually if acceptance tests are not comprehensive

---

## Effort Estimate

| Phase | Files | Complexity | Time |
|-------|-------|-----------|------|
| 1 | 4 | Low | 30 min |
| 2 | ~20 | Low | 1 hr |
| 3 | ~10 | Low | 30 min |
| 4 | - | Low | 15 min |
| **Total** | **~35** | **Low** | **2-2.5 hrs** |

---

## References

- Shared Library: `github.com/sonatype-nexus-community/terraform-provider-shared`
- Conversion Utils: `util/conversion.go` in shared library
- Time Utils: `util/time.go` in shared library (if needed)
- Example Implementation: `terraform-provider-sonatyperepo` uses these patterns

---

## Notes

- Nil handling is critical in Terraform provider development; shared utilities provide safe defaults
- The codebase already follows good patterns; this is optimization/standardization
- No custom conversion functions need to be removed (none exist in custom code)
- Import alias file makes future maintenance easier if shared library API changes
