<!-- See https://developer.hashicorp.com/terraform/plugin/best-practices/versioning#changelog-specification -->

## X.Y.Z (Unreleased)

ENHANCEMENTS:

* resource/sonatypeiq_organization: Now supports `terraform import` [GH-46]

BUG FIXES:

* Refresh Plan not empty for `sonatypeiq_source_control` due to Token masking [GH-50]


## 0.10.0 July 14, 2025

FEATURES:

* **New Resource:** `sonatypeiq_user_token` [GH-45]

## 0.9.1 July 14, 2025

ENHANCEMENTS:

* Documentation improved for resource `sonatypeiq_system_config` as per [GH-20]