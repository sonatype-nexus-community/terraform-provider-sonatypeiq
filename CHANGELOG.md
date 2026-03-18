<!-- See https://developer.hashicorp.com/terraform/plugin/best-practices/versioning#changelog-specification -->

## X.Y.Z (Unreleased)

BREAKING CHANGES:

See the [Migration Guide](./MIGRATION-1.0.md) for help migrating from v0.x.x versions of this provider.

* Updated Data Sources to always have an `id` and be defined consistently [GH-63]. See [Migration Guide](./MIGRATION-1.0.md) for impacted data sources.
* Updated Resources to ensure they always have an `id` and `last_updated` for consistency [Gh-64]. See [Migration Guide](./MIGRATION-1.0.md) for impacted resources.

ENHANCEMENTS:

* Adopted shared library to improve mantainability and consistency [GH-60]
* A number of additional resources now support import - see [Migration Guide](./MIGRATION-1.0.md)
* Resource `sonatypeiq_organization` now includes nested `categories`

NOTES:

* Tested against [Sonatype IQ Server 197](https://help.sonatype.com/en/sonatype-iq-server-197-release-notes.html) [GH-70]
* Tested against [Sonatype IQ Server 198](https://help.sonatype.com/en/sonatype-iq-server-198-release-notes.html) [GH-71]
* Tested against [Sonatype IQ Server 199](https://help.sonatype.com/en/sonatype-iq-server-199-release-notes.html) [GH-72]
* Tested against [Sonatype IQ Server 200](https://help.sonatype.com/en/sonatype-iq-server-200-release-notes.html) [GH-73]

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