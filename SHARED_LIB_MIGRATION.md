# Migration Plan: terraform-provider-sonatypeiq → terraform-provider-shared Adoption

## Overview

This document outlines the plan to adopt the shared Terraform library (`terraform-provider-shared`) in `terraform-provider-sonatypeiq`, following the pattern already established by `terraform-provider-sonatyperepo`.

**Current Status:** terraform-provider-sonatyperepo v0.7.2 uses `terraform-provider-shared v0.7.2`

**Goal:** Reduce code duplication, improve maintainability, and standardize error handling, schema builders, and utilities across Sonatype Terraform providers.

---

## Stage 1: Dependency Management & Infrastructure

### Tasks

1. **Add terraform-provider-shared dependency to go.mod**
   - Run: `go get github.com/sonatype-nexus-community/terraform-provider-shared@v0.7.2`
   - Match version with terraform-provider-sonatyperepo

2. **Update and verify dependencies**
   - Run: `go mod tidy`
   - Ensure no conflicts with existing dependencies

3. **Create internal shared utilities wrapper (optional but recommended)**
   - Create `internal/provider/common/shared_imports.go`
   - Purpose: Centralize imports from shared library for easier maintenance and future upgrades
   - Simplifies changing shared library imports across the codebase in one location

### Acceptance Criteria
- `go mod tidy` completes without errors
- All dependencies resolve correctly
- Build succeeds with no import errors

---

## Stage 2: Error Handling Refactoring

### Overview
Replace custom error handling patterns with standardized shared library functions from `github.com/sonatype-nexus-community/terraform-provider-shared/errors`.

### Tasks

1. **Audit current error handling**
   - Files: `internal/provider/common/` and all resource/data source files
   - Identify custom error functions and patterns
   - Document current error handling for each resource type

2. **Implement shared error helpers**
   - Import: `sharederr "github.com/sonatype-nexus-community/terraform-provider-shared/errors"`
   - Available functions:
     - `sharederr.HandleAPIError()` - centralized API error handling with network error detection
     - `sharederr.AddNotFoundDiagnostic()` - standardized 404 handling
     - `sharederr.AddValidationDiagnostic()` - validation errors
     - `sharederr.AddConflictDiagnostic()` - conflict errors (409)
     - `sharederr.AddUnauthorizedDiagnostic()` - permission errors (401)
     - `sharederr.AddTimeoutDiagnostic()` - timeout errors
   - Status code helpers: `IsNotFound()`, `IsForbidden()`, `IsUnauthorized()`, `IsConflict()`, `IsClientError()`, `IsServerError()`

3. **Update resource CRUD operations**
   - **Create operations:** Replace error handling with `sharederr.HandleAPIError()` or specific diagnostic adders
   - **Read operations:** Use `sharederr.AddNotFoundDiagnostic()` for 404 responses
   - **Update operations:** Use appropriate status-specific error handlers
   - **Delete operations:** Use `sharederr.AddNotFoundDiagnostic()` for already-deleted resources

4. **Standardize error messages**
   - Use shared error message patterns for consistency
   - Ensure all error summaries and details follow Terraform provider conventions

### Files to Update
- `internal/provider/common/model.go` and related common files
- All resource files: `internal/provider/application/*`, `internal/provider/organization/*`, etc.
- All data source files

### Acceptance Criteria
- All custom error handling replaced with shared library equivalents
- `go test ./...` passes
- Manual testing confirms error messages are user-friendly
- No regression in error reporting clarity

---

## Stage 3: Schema Builders Migration

### Overview
Replace manual schema attribute definitions with standardized builder functions from `github.com/sonatype-nexus-community/terraform-provider-shared/schema`.

### Tasks

1. **Audit schema definitions**
   - Scan all resource and data source files for manual attribute definitions
   - Identify patterns:
     - String attributes (required, optional, computed, computed-optional)
     - Integer attributes (same variants)
     - Boolean attributes (same variants)
     - Map attributes (string, int64, bool maps)
     - List attributes
     - Nested attributes (single and list)

2. **Available Schema Builder Functions**

   **String Attributes:**
   - `schema.ResourceRequiredString(description)`
   - `schema.ResourceOptionalString(description)`
   - `schema.ResourceComputedString(description)`
   - `schema.ResourceComputedOptionalString(description)`
   - `schema.ResourceStandardID()`
   - String variants with defaults and plan modifiers

   **Integer Attributes:**
   - `schema.ResourceRequiredInt64(description)`
   - `schema.ResourceOptionalInt64(description)`
   - `schema.ResourceComputedInt64(description)`
   - `schema.ResourceComputedOptionalInt64(description)`
   - `schema.ResourceComputedInt64WithDefault(description, default)`
   - `schema.ResourceOptionalInt64WithDefault(description, default)`
   - Variants with plan modifiers

   **Boolean Attributes:**
   - `schema.ResourceRequiredBool(description)`
   - `schema.ResourceOptionalBool(description)`
   - `schema.ResourceComputedBool(description)`
   - `schema.ResourceComputedOptionalBool(description)`
   - Variants with defaults and plan modifiers

   **Map Attributes:**
   - `schema.ResourceRequiredStringMap(description)`
   - `schema.ResourceOptionalStringMap(description)`
   - `schema.ResourceComputedStringMap(description)`
   - `schema.ResourceComputedOptionalStringMap(description)`
   - Similar variants for `Int64Map` and `BoolMap`

   **Nested Attributes:**
   - `schema.ResourceRequiredSingleNestedAttribute(description, attributes)`
   - `schema.ResourceOptionalSingleNestedAttribute(description, attributes)`
   - `schema.ResourceComputedSingleNestedAttribute(description, attributes)`
   - `schema.ResourceComputedOptionalSingleNestedAttribute(description, attributes)`
   - `schema.ResourceRequiredListNestedAttribute(description, nestedObject)`
   - `schema.ResourceOptionalListNestedAttribute(description, nestedObject)`
   - `schema.ResourceComputedListNestedAttribute(description, nestedObject)`

3. **Replace manual definitions systematically**
   - Start with simple resources (fewest nested attributes)
   - Progress to complex resources
   - Maintain consistent description formatting

4. **Apply standardized descriptions**
   - Use clear, user-friendly attribute descriptions
   - Reference terraform-provider-sonatyperepo examples for consistency

### Files to Update
- All resource definition files (Schema() methods)
- All data source definition files (Schema() methods)
- Nested attribute definitions in `internal/provider/*/` subdirectories

### Acceptance Criteria
- All manual schema definitions replaced with builder functions
- Schema attributes match exactly (same properties, defaults, validators)
- `go test ./...` passes
- Terraform schema validation passes
- Documentation generation still works correctly

---

## Stage 4: Utility Functions & Type Conversions

### Overview
Leverage type conversion and utility functions from `github.com/sonatype-nexus-community/terraform-provider-shared/util`.

### Tasks

1. **Identify custom conversion functions**
   - Search: `internal/provider/utils/` and other utility files
   - Find patterns:
     - string ↔ *string conversions
     - bool ↔ *bool conversions
     - int64 ↔ *int64 conversions
     - string ↔ int64 conversions

2. **Shared Library Utilities Available**

   **Type Conversions (util/conversion.go):**
   - `util.StringToPtr(s string) *string`
   - `util.StringPtrToValue(s *string) string`
   - `util.BoolToPtr(b bool) *bool`
   - `util.Int64ToPtr(i int64) *int64`
   - `util.StringToInt64(s string) int64`

   **Time/Timestamp Utilities (util/time.go):**
   - `util.CurrentTimestamp() string`
   - `util.ParseTimestamp(timestamp string) time.Time`
   - `util.UnixTimestamp(t time.Time) int64`
   - `util.StringToUnixTimestamp(s string) int64`

3. **Replace custom conversion helpers**
   - Remove duplicate conversion functions from `internal/provider/utils/`
   - Update all resource files to import and use shared utilities
   - Verify nil-safety of conversions (shared library handles this)

4. **Integrate timestamp handling**
   - If IQ API uses timestamps, replace custom parsing with shared utilities
   - Standardize timestamp formats across provider

### Files to Update
- `internal/provider/utils/` - remove redundant conversion functions
- All resource files using custom conversions
- Data source files with type conversions

### Acceptance Criteria
- All duplicate conversion functions removed
- All resource files use shared utility functions
- `go test ./...` passes
- No behavioral changes in type conversions
- Timestamp handling (if applicable) is standardized

---

## Stage 5: Validators Integration

### Overview
Implement standardized validation rules from `github.com/sonatype-nexus-community/terraform-provider-shared/validators`.

### Tasks

1. **Identify enum/constrained fields**
   - Scan all schema definitions for fields with limited valid values
   - Document current validation approaches (custom, framework validators, etc.)

2. **Available Shared Validators**

   **Current Implementation:**
   - `validators.StringOneOfValidator(validValues ...string) []validator.String`
     - Validates that a string attribute value matches one of the provided options
     - Example: `validators.StringOneOfValidator("active", "inactive", "pending")`

3. **Apply string enum validation**
   - Find all enum-like string attributes
   - Replace custom validators with `StringOneOfValidator`
   - Add to attribute builder calls (validators field)

4. **Document custom validation patterns**
   - If provider-specific validators are needed, keep them in `internal/provider/validators/`
   - Consider contributing commonly-needed validators back to shared library

### Files to Update
- All resource/data source schema definition files with enum fields
- Any custom validator files in `internal/provider/validators/`

### Acceptance Criteria
- All enum string fields use shared validators where applicable
- Custom validators remain only for provider-specific patterns
- `go test ./...` passes
- Terraform validation works correctly for constrained fields

---

## Stage 6: Resource Base Class (Optional but Recommended)

### Overview
Evaluate adopting the `BaseResource` pattern from `github.com/sonatype-nexus-community/terraform-provider-shared/resource` to reduce boilerplate code.

### Background
- `BaseResource` provides common configuration access
- Implements standard provider configuration pattern
- Methods: `Configure()`, `GetAuth()`, `GetBaseURL()`, `GetClient()`, `IsConfigured()`

### Decision Points

**Adopt BaseResource if:**
- Individual resources have significant boilerplate for configuration access
- Provider has 10+ resources with similar patterns
- Configuration pattern is consistent across all resources

**Skip BaseResource if:**
- Resources are already optimized
- Configuration access is highly variable per resource
- Adoption would require major refactoring

### If Adopting: Implementation Steps

1. **Create wrapper interface**
   - Define what the provider's `ProviderData` struct looks like
   - Ensure compatibility with `BaseResource` expectations

2. **Refactor resources to use BaseResource**
   - Replace custom Configure() implementations
   - Use GetClient(), GetAuth(), GetBaseURL() methods
   - Reduce repeated boilerplate code

3. **Update provider configuration**
   - Ensure all resources embed or inherit from `BaseResource`
   - Verify error handling in Configure() method

### Files to Update
- `internal/provider/provider.go` - provider configuration
- All resource files - Configure() methods
- All data source files - Configure() methods

### Acceptance Criteria
- Resources properly configured through BaseResource
- `go test ./...` passes
- No regression in provider functionality
- Reduced code duplication

---

## Stage 7: Testing & Validation

### Overview
Comprehensive testing and validation of all migration changes.

### Tasks

1. **Run full test suite**
   ```bash
   go test -v -cover ./...
   ```
   - Verify all unit tests pass
   - Review coverage to ensure quality

2. **Lint and format code**
   ```bash
   go fmt ./...
   go vet ./...
   golangci-lint run ./...
   ```

3. **Build provider binary**
   ```bash
   go build -v -o terraform-provider-sonatypeiq_v0.1.0
   ```

4. **Test with local Terraform**
   - Use examples from `examples/` directory
   - Test all CRUD operations for each resource
   - Verify error messages are appropriate and user-friendly

5. **Integration testing**
   - Run against actual IQ instance (if available in test environment)
   - Test all example configurations
   - Verify no regressions in functionality

6. **Documentation updates**
   - Add CHANGELOG.md entry:
     - Note: "Internal refactoring: adopted terraform-provider-shared library for standardized error handling, schema builders, and utilities"
     - No API changes, no user-facing changes
   - Update CONTRIBUTING.md if needed to reference shared library patterns

7. **Registry validation** (if applicable)
   - If provider is published to Terraform Registry, verify registry integration still works
   - Check documentation generation

### Acceptance Criteria
- All tests pass: `go test -v ./...`
- No lint errors: `golangci-lint run ./...`
- Provider builds successfully
- Manual testing confirms all resource operations work
- No breaking changes to provider API
- CHANGELOG updated

---

## Missing Helper Methods in terraform-provider-shared

Based on analysis of terraform-provider-sonatypeiq's patterns and comparison with existing shared library functions, the following helper methods **may need to be added** to `terraform-provider-shared`:

### High Priority (Likely Needed Soon)

#### 1. Additional Diagnostic Error Adders
**Location:** `errors/errors.go`

```go
// For 403 Forbidden responses (currently has IsF orbidden() but no diagnostic adder)
func AddForbiddenDiagnostic(diags *diag.Diagnostics, operation string)

// For generic 5xx errors
func AddServerErrorDiagnostic(diags *diag.Diagnostics, message string, statusCode int)

// For generic 4xx errors
func AddClientErrorDiagnostic(diags *diag.Diagnostics, message string, statusCode int)
```

**Rationale:** `HandleAPIError()` is good for network detection, but these would provide simpler APIs for common HTTP status codes without needing to construct detailed error structures.

---

#### 2. List Nested Attribute Variants
**Location:** `schema/nested_attributes.go`

```go
// Missing variant (verify if truly missing)
func ResourceComputedListNestedAttribute(description string, nestedObject resourceschema.NestedAttributeObject) resourceschema.ListNestedAttribute

// If nested objects need default list values
func ResourceOptionalListNestedAttributeWithDefault(description string, nestedObject resourceschema.NestedAttributeObject, defaultValue []interface{}) resourceschema.ListNestedAttribute
```

**Rationale:** Consistency with other attribute builders; may be needed for complex list defaults.

---

#### 3. String Attribute Variants with Plan Modifiers
**Location:** `schema/string_attributes.go` (verify if file exists)

```go
// Additional variants for string attributes
func ResourceOptionalStringWithDefaultAndPlanModifier(description string, defaultValue string, planMods ...planmodifier.String) resourceschema.StringAttribute

func ResourceComputedStringWithDefaultAndPlanModifier(description string, defaultValue string, planMods ...planmodifier.String) resourceschema.StringAttribute
```

**Rationale:** Parallel pattern to int/float attributes; may be needed for IQ-specific attributes that require custom plan modifications.

---

### Medium Priority (Nice to Have)

#### 4. Provider Configuration Validation Helper
**Location:** `resource/base.go`

```go
// Validates that provider data is correctly configured
func ValidateBaseResourceConfig(providerData interface{}) error
```

**Rationale:** Reduces duplicate validation logic across providers; improves consistency in handling misconfigured providers.

---

#### 5. List/Pagination Response Marshaling
**Location:** New file `util/pagination.go`

```go
// Helper to marshal paginated API responses to Terraform list state
func MarshalListResponse(items interface{}, limit int) []map[string]interface{}

// Convert paginated items to proper Terraform types
func ConvertItemsToTfTypes(items interface{}) ([]map[string]attr.Value, error)
```

**Rationale:** Many APIs use pagination; shared pattern could reduce code duplication.

---

#### 6. Schema Description Helpers
**Location:** `schema/descriptions.go` (new file)

```go
// Standard descriptions for common attributes
const (
    DescriptionID        = "The unique identifier for this resource."
    DescriptionCreatedAt = "The timestamp when this resource was created."
    DescriptionUpdatedAt = "The timestamp when this resource was last updated."
    DescriptionName      = "The name of this resource."
)

// Generate consistent attribute descriptions
func DescribeAttribute(fieldName string, customSuffix string) string
```

**Rationale:** Improves consistency in documentation; reduces boilerplate description text.

---

### Lower Priority (Context-Specific)

#### 7. Sonatype-Specific Helpers
**Potential Location:** New package `sonatype/` within shared library

```go
// Version checking for IQ/Repo APIs
func IsVersionGreaterThan(currentVersion, targetVersion string) (bool, error)

// Common auth patterns for Sonatype APIs
func ValidateBasicAuth(username, password string) error

// IQ-specific API response handling
func ExtractIQErrorMessage(response *http.Response) string
```

**Rationale:** These are highly specific to Sonatype APIs and might better belong in a separate package within shared library or in provider-specific code.

---

## Implementation Recommendations

### Code Organization
- Create import aliases for clarity: `sharederr`, `sharedrschema`, `sharedutil`, `sharedvalidators`
- Keep shared library imports in dedicated common file for easy updates
- Document why shared library is used vs. custom implementations

### Backward Compatibility
- This is an internal refactoring with no breaking changes to the provider API
- Existing Terraform configurations will continue to work unchanged

### Testing Strategy
- Unit tests: Verify shared library functions work as expected in IQ context
- Integration tests: Confirm all resource operations still function correctly
- Manual testing: Validate error messages are clear and helpful

### Future Considerations
- Monitor `terraform-provider-shared` for new releases and features
- Contribute back IQ-specific patterns if they would benefit all providers
- Consider contributing missing helper methods (High Priority) back to shared library
- Plan quarterly reviews of shared library adoption to identify optimization opportunities

---

## Success Metrics

- [ ] All stages completed without regressions
- [ ] Test coverage maintained or improved
- [ ] Code duplication reduced by >30% (estimated)
- [ ] No changes to external provider API
- [ ] Documentation updated
- [ ] CHANGELOG entry completed
- [ ] All resources tested manually
- [ ] Provider builds and deploys successfully

---

## Timeline Estimate

| Stage | Complexity | Time | Notes |
|-------|-----------|------|-------|
| 1 | Low | 30 min | Straightforward dependency add |
| 2 | Medium | 2-3 hrs | Largest number of files to update |
| 3 | Medium | 2-3 hrs | Many schema definitions to refactor |
| 4 | Low | 45 min | Mostly search-and-replace |
| 5 | Low | 30 min | Only specific enum fields |
| 6 | Medium | 1-2 hrs | Optional, depends on adoption decision |
| 7 | High | 2-3 hrs | Testing all resources thoroughly |
| **Total** | **-** | **9-14 hrs** | Can be split into multiple sessions |

---

## Rollback Plan

If issues arise during migration:

1. Revert to the most recent stable commit before migration started
2. Create feature branch for incremental approach (migrate one stage at a time)
3. If specific stage fails, complete that stage in isolation before moving to next
4. Use git bisect to identify problematic changes

---

## Questions & Contact

For questions about this plan:
- Review terraform-provider-sonatyperepo implementation as reference
- Check terraform-provider-shared documentation in its README and examples/
- Reference shared library source code for implementation details
- Consult with team on adoption of BaseResource (optional Stage 6)

