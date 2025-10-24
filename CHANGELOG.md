<!-- See https://developer.hashicorp.com/terraform/plugin/best-practices/versioning#changelog-specification -->

## X.Y.Z (Unreleased)

FEATURES:

* **New Resource:** `sonatypeiq_application_category` [GH-1]
* **New Resource:** `sonatypeiq_config_crowd` [GH-1]
* **New Resource:** `sonatypeiq_config_product_license` [GH-1]

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