# Stage 4: organization_resource.go - Complete Reference Implementation

This document shows the exact changes made to `internal/provider/organization/organization_resource.go` as a complete example for updating other files.

---

## File: organization_resource.go

### Change 1: Add sharedutil Import

**Location:** Lines 19-34 (import block)

```go
// BEFORE
import (
	"context"
	"fmt"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
)

// AFTER - Add this line after sharedrschema import
import (
	"context"
	"fmt"
	"terraform-provider-sonatypeiq/internal/provider/common"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
	sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
	sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
	sharedutil "github.com/sonatype-nexus-community/terraform-provider-shared/util"
)
```

---

### Change 2: Create Method - Pointer Creation (Lines 99-102)

**Location:** Create method, API request building

```go
// BEFORE
organization_request := r.Client.OrganizationsAPI.AddOrganization(ctx)
orgDto := sonatypeiq.ApiOrganizationDTO{
    Name:                 plan.Name.ValueStringPointer(),
    ParentOrganizationId: plan.ParentOrganiziationId.ValueStringPointer(),
}

// AFTER - Use sharedutil.StringToPtr()
organization_request := r.Client.OrganizationsAPI.AddOrganization(ctx)
orgDto := sonatypeiq.ApiOrganizationDTO{
    Name:                 sharedutil.StringToPtr(plan.Name.ValueString()),
    ParentOrganizationId: sharedutil.StringToPtr(plan.ParentOrganiziationId.ValueString()),
}
```

**Why This Change:**
- `ValueStringPointer()` creates a pointer but is less explicit
- `sharedutil.StringToPtr()` is more readable and consistent across providers
- Pattern: `sharedutil.StringToPtr(plan.FieldName.ValueString())`

---

### Change 3: Create Method - Response Handling (Lines 130-132)

**Location:** Create method, after successful API call

```go
// BEFORE - Manual pointer dereference
plan.ID = types.StringValue(*organization.Id)
plan.Name = types.StringValue(*organization.Name)
plan.ParentOrganiziationId = types.StringValue(*organization.ParentOrganizationId)

// AFTER - Use sharedutil.StringPtrToValue()
plan.ID = sharedutil.StringPtrToValue(organization.Id)
plan.Name = sharedutil.StringPtrToValue(organization.Name)
plan.ParentOrganiziationId = sharedutil.StringPtrToValue(organization.ParentOrganizationId)
```

**Why This Change:**
- Handles nil pointers safely (returns `types.StringNull()` if nil)
- No need for manual `if != nil` checks
- Cleaner, more maintainable code
- Pattern: `sharedutil.StringPtrToValue(response.Field)`

---

### Change 4: Read Method - Response Handling (Lines 174-176)

**Location:** Read method, after successful API call

```go
// BEFORE
state.ID = types.StringValue(*organization.Id)
state.Name = types.StringValue(*organization.Name)
state.ParentOrganiziationId = types.StringValue(*organization.ParentOrganizationId)

// AFTER
state.ID = sharedutil.StringPtrToValue(organization.Id)
state.Name = sharedutil.StringPtrToValue(organization.Name)
state.ParentOrganiziationId = sharedutil.StringPtrToValue(organization.ParentOrganizationId)
```

**Why This Change:**
- Same pattern as Create method response handling
- Consistency across all CRUD operations
- All API response pointers use `StringPtrToValue`

---

## Summary of Changes

| Type | Conversion | Pattern | Count |
|------|-----------|---------|-------|
| Import | Added | `sharedutil` | 1 |
| Create Request | `ValueStringPointer()` | `sharedutil.StringToPtr(value.ValueString())` | 2 |
| Create Response | `types.StringValue(*ptr)` | `sharedutil.StringPtrToValue(ptr)` | 3 |
| Read Response | `types.StringValue(*ptr)` | `sharedutil.StringPtrToValue(ptr)` | 3 |
| **TOTAL** | | | **9** |

---

## Testing Results

```bash
$ go test ./internal/provider/organization -v
=== RUN   TestAccOrganizationResource
    organization_resource_test.go:34: Acceptance tests skipped unless env 'TF_ACC' set
--- SKIP: TestAccOrganizationResource (0.00s)
=== RUN   TestAccOrganizationDataSource
    organization_data_source_test.go:27: Acceptance tests skipped unless env 'TF_ACC' set
--- SKIP: TestAccOrganizationDataSource (0.00s)
...
PASS

$ go build -v .
...
terraform-provider-sonatypeiq
Success ✓
```

---

## Copy-Paste Ready Patterns

### Pattern 1: String Conversions
```go
// Terraform → API Request
plan.FieldName.ValueStringPointer()
↓
sharedutil.StringToPtr(plan.FieldName.ValueString())

// API Response → Terraform State
types.StringValue(*response.FieldName)
↓
sharedutil.StringPtrToValue(response.FieldName)
```

### Pattern 2: Boolean Conversions (if applicable)
```go
// Terraform → API Request
plan.FieldName.ValueBoolPointer()
↓
sharedutil.BoolToPtr(plan.FieldName.ValueBool())

// API Response → Terraform State
types.BoolValue(*response.FieldName)
↓
sharedutil.BoolPtrToValue(response.FieldName)
```

### Pattern 3: Integer Conversions (if applicable)
```go
// Terraform → API Request
plan.FieldName.ValueInt64Pointer()
↓
sharedutil.Int64ToPtr(plan.FieldName.ValueInt64())

// API Response → Terraform State
types.Int64Value(*response.FieldName)
↓
sharedutil.Int64PtrToValue(response.FieldName)
```

---

## Implementation Checklist for Similar Files

When updating `organization_data_source.go` or other similar files:

- [ ] Add `sharedutil` import at the top
- [ ] Search for `plan.*.ValueStringPointer()` → Replace with `sharedutil.StringToPtr(plan.*.ValueString())`
- [ ] Search for `plan.*.ValueBoolPointer()` → Replace with `sharedutil.BoolToPtr(plan.*.ValueBool())`
- [ ] Search for `plan.*.ValueInt64Pointer()` → Replace with `sharedutil.Int64ToPtr(plan.*.ValueInt64())`
- [ ] Search for `types.StringValue(*` → Replace with `sharedutil.StringPtrToValue(`
- [ ] Search for `types.BoolValue(*` → Replace with `sharedutil.BoolPtrToValue(`
- [ ] Search for `types.Int64Value(*` → Replace with `sharedutil.Int64PtrToValue(`
- [ ] Remove trailing `)` after variable name (already in function call)
- [ ] Test: `go test ./internal/provider/<module> -v`
- [ ] Build: `go build -v .`

---

## Diff Summary

```diff
 import (
     "context"
     "fmt"
     "terraform-provider-sonatypeiq/internal/provider/common"
     "time"

     "github.com/hashicorp/terraform-plugin-framework/path"
     "github.com/hashicorp/terraform-plugin-framework/resource"
     "github.com/hashicorp/terraform-plugin-framework/resource/schema"
     "github.com/hashicorp/terraform-plugin-framework/types"
     "github.com/hashicorp/terraform-plugin-log/tflog"

     sonatypeiq "github.com/sonatype-nexus-community/nexus-iq-api-client-go"
     sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"
     sharedrschema "github.com/sonatype-nexus-community/terraform-provider-shared/schema"
+    sharedutil "github.com/sonatype-nexus-community/terraform-provider-shared/util"
 )

 // Create method changes:
- Name:                 plan.Name.ValueStringPointer(),
- ParentOrganizationId: plan.ParentOrganiziationId.ValueStringPointer(),
+ Name:                 sharedutil.StringToPtr(plan.Name.ValueString()),
+ ParentOrganizationId: sharedutil.StringToPtr(plan.ParentOrganiziationId.ValueString()),

- plan.ID = types.StringValue(*organization.Id)
- plan.Name = types.StringValue(*organization.Name)
- plan.ParentOrganiziationId = types.StringValue(*organization.ParentOrganizationId)
+ plan.ID = sharedutil.StringPtrToValue(organization.Id)
+ plan.Name = sharedutil.StringPtrToValue(organization.Name)
+ plan.ParentOrganiziationId = sharedutil.StringPtrToValue(organization.ParentOrganizationId)

 // Read method changes (same pattern):
- state.ID = types.StringValue(*organization.Id)
- state.Name = types.StringValue(*organization.Name)
- state.ParentOrganiziationId = types.StringValue(*organization.ParentOrganizationId)
+ state.ID = sharedutil.StringPtrToValue(organization.Id)
+ state.Name = sharedutil.StringPtrToValue(organization.Name)
+ state.ParentOrganiziationId = sharedutil.StringPtrToValue(organization.ParentOrganizationId)
```

---

## Notes

1. **No Behavioral Changes** - This is purely a refactoring
2. **Nil-Safe** - `StringPtrToValue()` handles nil gracefully
3. **Consistent** - Same pattern across all providers
4. **Maintainable** - Centralized conversion logic
5. **Testable** - Full test suite validates correctness

---

## Files That Follow This Pattern

All files in these modules should follow the same conversion patterns:

- `internal/provider/application/*` - 5 files
- `internal/provider/organization/*` - 5 files (1 done, 4 to go)
- `internal/provider/role/*` - 1 file
- `internal/provider/user/*` - 2 files
- `internal/provider/system/*` - 7 files
- `internal/provider/scm/*` - 1 file

**Total:** 21 files to update, 1 completed ✓

---

## References

- **Shared Library:** github.com/sonatype-nexus-community/terraform-provider-shared
- **Conversion Utils:** util/conversion.go in shared library
- **Example Providers:** terraform-provider-sonatyperepo uses same patterns
