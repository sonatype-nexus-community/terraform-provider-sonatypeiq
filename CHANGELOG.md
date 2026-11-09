<!-- See https://developer.hashicorp.com/terraform/plugin/best-practices/versioning#changelog-specification -->

## X.Y.Z (Unreleased)

FEATURES:
* New resource `sonatypeiq_config_reverse_proxy_authentication` to manage reverse proxy authentication configuration for when users authenticate via a reverse proxy [GH-134]
* New resource `sonatypeiq_config_jira` to manage the Jira integration configuration for Sonatype IQ Server [GH-135]
* New resource `sonatypeiq_user_token_configuration` to manage the default expiration of user tokens [GH-133]

NOTES:
* Support for Sonatype IQ Server 186, 187, 188, 189, 190 and 191 has been dropped. The provider is now supported from release 192 onwards [GH-96] [GH-97] [GH-98] [GH-99] [GH-100] [GH-101]
* The automated tests now also run against Sonatype IQ Server 192, 193 and 194, so every supported release is tested
* New resource `sonatypeiq_config_oidc` to manage OpenID Connect (OIDC) SSO configuration [GH-125]

_tbc_

## 1.2.0 Nov 08, 2026

BUG FIXES:
* `sonatypeiq_system_config` no longer fails with "inconsistent result after apply" when `base_url` is configured without the trailing `/` that Sonatype IQ Server always adds [GH-92]
* The provider no longer requires administrator permissions just to start. It checked that it could reach Sonatype IQ Server by reading a user, which needs _Edit System Configuration and Users_, so any other user failed with `403 Forbidden` before any resource was used. It now reads the version from the anonymous `/ping` endpoint and checks the credentials with a call that needs no permissions, so a user with only the permissions for the resources they manage (for example `sonatypeiq_user_token`) can use it. The permission each resource and data source needs is now documented [GH-91]
* An unrecognised `Server` header (for example one rewritten by a reverse proxy) no longer produces a misleading "older than 186" warning; the provider now says the version could not be determined [GH-91]
* `sonatypeiq_organization_role_membership` and `sonatypeiq_application_role_membership` now explain that a global role must be granted with `sonatypeiq_global_role_membership` (and the reverse) instead of showing a bare `400 Bad Request` [GH-88]

FEATURES:
* New resource `sonatypeiq_global_role_membership` to grant a global role, such as System Administrator, to a user or a group. Global roles cannot be granted with `sonatypeiq_organization_role_membership` or `sonatypeiq_application_role_membership` [GH-88]
* Added `version_hint` to the provider configuration, for when a reverse proxy or load balancer removes or rewrites the `Server` header that the provider reads the version of Sonatype IQ Server from. It takes the full version (e.g. `1.205.4-01`) or the release number (e.g. `205`) and takes precedence over what the server reports [GH-91]
* `sonatypeiq_system_config` resource and data source now support all properties of the Sonatype IQ Server Configuration REST API that the API client models: `session_timeout`, `csrf_protection`, `policy_monitoring_hour`, `user_agent_suffix`, `hds_url`, `frame_ancestors_allowlist`, `event_bus_max_thread_pool_size`, `webhook_secret_passphrase`, `waived_component_upgrade_monitoring_enabled`, `api_access_allow_list`, `max_advanced_search_clause_count`, `advanced_search_csv_export_delimiter`, `advanced_reporting_insights_enabled`, `success_metrics_stage_id`, `purge_scan_files`, `alp_observed_license_detection_enabled` and `integrations_supported_version_count` [GH-92]
* Setting a `sonatypeiq_system_config` property that the connected Sonatype IQ Server is too old to support now fails at plan time, naming the release that introduced it. Each property's documentation states the release that introduced it [GH-92]
* Confirmed support for Sonatype IQ Server 209 [GH-109]
* Confirmed support for Sonatype IQ Server 208 [GH-108]
* Confirmed support for Sonatype IQ Server 207 [GH-107]
* Confirmed support for Sonatype IQ Server 206 [GH-106]

DEPENDENCIES:
* Bumped `github.com/sonatype-nexus-community/nexus-iq-api-client-go` to v0.209.1

## 1.1.0 Nov 07, 2026

BUG FIXES:
* CI: acceptance tests against Sonatype IQ Server 204+ now use Java 25 and fail fast with diagnostics if IQ Server does not start [GH-118]

FEATURES:
* Confirmed support for Sonatype IQ Server 205 [GH-105]
* Confirmed support for Sonatype IQ Server 204 [GH-104]
* Confirmed support for Sonatype IQ Server 203 [GH-103]
* Confirmed support for Sonatype IQ Server 202 [GH-102]

NOTES:
* CI aligned to the Sonatype Community [Pull Request CI & Secrets standard](https://contribute.sonatype.com/docs/standards/pull-request-ci/): open checks and a Lifecycle evaluation run for every Pull Request without exposing secrets, and acceptance tests against a licensed Sonatype IQ Server run once a maintainer approves the commit with `/ok-to-test <sha>`. `release.yml` has per-job permissions and a protected `release` Environment [GH-93]

DEPENDENCIES:
* Bumped `github.com/sonatype-nexus-community/nexus-iq-api-client-go` to v0.205.0 (there is no v0.202.x release; the SAML configuration data source now calls the renamed `GetMetadata1`)
* Bumped `golang.org/x/net` to v0.62.0, resolving multiple CVEs (up to CVSS 9.3), including a Critical Sonatype Lifecycle policy violation. This also moves `golang.org/x/crypto`, `golang.org/x/sync`, `golang.org/x/sys`, `golang.org/x/text` and `golang.org/x/tools` to the minimum versions it requires

## 1.0.2 Sep 09, 2026

BUG FIXES:
* Unable to import `sonatypeiq_user` with a `username` containing `-` [GH-89]

DEPENDENCIES:
* Removed legacy `github.com/hashicorp/terraform-plugin-sdk` (v1) in favor of `terraform-plugin-sdk/v2`, resolving several policy violations flagged by Sonatype Lifecycle
* Bumped `golang.org/x/crypto` to v0.57.0, resolving multiple CVEs (up to CVSS 10.0)
* Bumped `golang.org/x/net` to v0.59.0, resolving multiple CVEs (up to CVSS 9.6)
* Bumped `google.golang.org/grpc` to v1.83.2, resolving CVEs (up to CVSS 8.7)
* Bumped `golang.org/x/mod` to v0.41.0, resolving CVEs (up to CVSS 8.4)
* Bumped `golang.org/x/text` to v0.42.0, resolving CVE-2026-56852
* Bumped `golang.org/x/sys` to v0.48.0, resolving CVE-2026-39824

## 1.0.1 May 05, 2026

BUG FIXES:
* Unable to update `token` on `sonatypeiq_source_control` resource [GH-84]

## 1.0.0 March 20, 2026

BREAKING CHANGES:

See the [Migration Guide](./MIGRATION-1.0.md) for help migrating from v0.x.x versions of this provider.

* Updated Data Sources to always have an `id` and be defined consistently [GH-63]. See [Migration Guide](./MIGRATION-1.0.md) for impacted data sources.
* Updated Resources to ensure they always have an `id` and `last_updated` for consistency [Gh-64]. See [Migration Guide](./MIGRATION-1.0.md) for impacted resources.

FEATURES:

* **New Resource:** `sonatypeiq_role` [GH-61]

ENHANCEMENTS:

* `sonatypeiq_role` data source now returns more complete fields including permissions [GH-67]
* Adopted shared library to improve mantainability and consistency [GH-60]
* A number of additional resources now support import - see [Migration Guide](./MIGRATION-1.0.md)
* Resource `sonatypeiq_organization` now includes nested `categories`

NOTES:

* Tested against [Sonatype IQ Server 197](https://help.sonatype.com/en/sonatype-iq-server-197-release-notes.html) [GH-70]
* Tested against [Sonatype IQ Server 198](https://help.sonatype.com/en/sonatype-iq-server-198-release-notes.html) [GH-71]
* Tested against [Sonatype IQ Server 199](https://help.sonatype.com/en/sonatype-iq-server-199-release-notes.html) [GH-72]
* Tested against [Sonatype IQ Server 200](https://help.sonatype.com/en/sonatype-iq-server-200-release-notes.html) [GH-73]
* Tested against [Sonatype IQ Server 201](https://help.sonatype.com/en/sonatype-iq-server-201-release-notes.html) [GH-74]

## 0.12.1 October 24, 2025

NOTES:
* Updated documentation to call out tested against Sonatype IQ up to and included 196.

## 0.12.0 October 24, 2025

FEATURES:

* **New Resource:** `sonatypeiq_application_category` [GH-1]
* **New Resource:** `sonatypeiq_config_crowd` [GH-1]
* **New Resource:** `sonatypeiq_config_product_license` [GH-1]
* **New Resource:** `sonatypeiq_config_saml` [GH-1]

NOTES:
* Tested against Sonatype IQ Server 194, 195 and 196.

## 0.11.0 July 30, 2025

ENHANCEMENTS:

* resource/sonatypeiq_organization: Now supports `terraform import` [GH-46]

BUG FIXES:

* Fix for using Terraform 1.12.x+ [GH-47]
* Refresh Plan not empty for `sonatypeiq_source_control` due to Token masking [GH-50]
* Data Source `sonatypeiq_config_saml` no longer errors when there is no SAML configuration present [GH-53]


## 0.10.0 July 14, 2025

FEATURES:

* **New Resource:** `sonatypeiq_user_token` [GH-45]

## 0.9.1 July 14, 2025

ENHANCEMENTS:

* Documentation improved for resource `sonatypeiq_system_config` as per [GH-20]
