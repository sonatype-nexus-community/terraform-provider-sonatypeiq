# Stage 4: Files to Update - Priority List

## Completed ✓
- [x] `internal/provider/organization/organization_resource.go` - **EXAMPLE FILE** (DONE)

---

## Phase 1: High Priority (Simple Resources - No Nested Types)

### Application Module (5 files)
- [ ] `internal/provider/application/application_resource.go`
- [ ] `internal/provider/application/application_data_source.go`
- [ ] `internal/provider/application/application_categories_data_source.go`
- [ ] `internal/provider/application/applications_data_source.go`
- [ ] `internal/provider/application/application_role_membership_resource.go`

### Organization Module (4 files)
- [x] `internal/provider/organization/organization_resource.go` ✓
- [ ] `internal/provider/organization/organization_data_source.go`
- [ ] `internal/provider/organization/organization_category_resource.go`
- [ ] `internal/provider/organization/organizations_data_source.go`
- [ ] `internal/provider/organization/organization_role_membership_resource.go`

### Role Module (1 file)
- [ ] `internal/provider/role/role_data_source.go`

### User Module (2 files)
- [ ] `internal/provider/user/user_resource.go`
- [ ] `internal/provider/user/user_token_resource.go`

---

## Phase 2: Medium Priority (System Configuration Resources)

### System Module (7 files)
- [ ] `internal/provider/system/config_crowd_resource.go`
- [ ] `internal/provider/system/config_license_resource.go`
- [ ] `internal/provider/system/config_mail_resource.go`
- [ ] `internal/provider/system/config_proxy_server_resource.go`
- [ ] `internal/provider/system/config_saml_data_source.go`
- [ ] `internal/provider/system/system_config_data_source.go`
- [ ] `internal/provider/system/system_config_resource.go`

### SCM Module (1 file)
- [ ] `internal/provider/scm/source_control_resource.go`

---

## Total Files: 21 (1 completed, 20 remaining)

---

## Conversion Patterns to Apply

### Pattern 1: Import Addition
Add to imports section:
```go
sharedutil "github.com/sonatype-nexus-community/terraform-provider-shared/util"
```

### Pattern 2: Pointer Creation
```go
// BEFORE
Name: plan.Name.ValueStringPointer(),

// AFTER
Name: sharedutil.StringToPtr(plan.Name.ValueString()),
```

### Pattern 3: Pointer to Terraform Type
```go
// BEFORE
state.ID = types.StringValue(*organization.Id)

// AFTER
state.ID = sharedutil.StringPtrToValue(organization.Id)
```

### Pattern 4: Boolean Values (if applicable)
```go
// BEFORE
Enabled: plan.Enabled.ValueBoolPointer(),
state.Enabled = types.BoolValue(*organization.Enabled)

// AFTER
Enabled: sharedutil.BoolToPtr(plan.Enabled.ValueBool()),
state.Enabled = sharedutil.BoolPtrToValue(organization.Enabled)
```

### Pattern 5: Integer Values (if applicable)
```go
// BEFORE
Priority: plan.Priority.ValueInt64Pointer(),
state.Priority = types.Int64Value(*organization.Priority)

// AFTER
Priority: sharedutil.Int64ToPtr(plan.Priority.ValueInt64()),
state.Priority = sharedutil.Int64PtrToValue(organization.Priority)
```

---

## Testing Each File

After updating each file, run:
```bash
# Test the specific package
go test ./internal/provider/<module> -v

# Build the provider
go build -v .

# Format and lint (optional)
go fmt ./...
go vet ./...
```

---

## Notes

1. **organization_resource.go** serves as the reference implementation
   - Copy the import pattern from there
   - Follow the exact conversion patterns shown

2. **Be consistent** with spacing and formatting
   - Each conversion should follow the same style

3. **Test after each file** to catch errors early

4. **Commit message template:**
   ```
   Stage 4: Replace type conversions with shared util in <module>

   - Replace ValueStringPointer() with sharedutil.StringToPtr()
   - Replace types.StringValue(*) with sharedutil.StringPtrToValue()
   - Similar updates for Bool and Int64 types where applicable
   - No behavioral changes, purely internal refactoring
   ```

5. **Total Estimated Time:** ~2-2.5 hours
   - Per file: 5-8 minutes for straightforward resources
   - Per file: 10-15 minutes for complex resources with nested types
