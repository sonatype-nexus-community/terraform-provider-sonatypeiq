# Migrating from v0.x.x to v1.x.x

Release 1.x.x is a breaking-change release. This means that some ways of working / names and conventions have
changed and are not backwards compatible with 0.x.x releases.

Below is a summary of the key breaking changes with information on how you can update your Terraform HCL to work with 1.x.x releases.

- [Breaking Changes in 1.0.0](#breaking-changes-in-100)
  - [Schema Changes](#schema-changes)
  - [Resources](#resources)
- [Improvements](#improvements)

## Breaking Changes in 1.0.0

### Schema Changes

#### Data Sources

The following data sources have schema changes:
- `sonatypeiq_organization`: `tags` has been renamed to `categories`
- `sonatypeiq_organizations`: `tags` has been renamed to `categories`

The following data sources have experienced field visibility changes as part of [GH-63]:
- `sonatypeiq_applications`
- `sonatypeiq_application_categories`
- `sonatypeiq_system_config`

### Resources

The following resources have schema changes:
- `sonatyepiq_config_mail`: `password_is_included` has been removed as it had no purpose
- `sonatypeiq_config_proxy_server`: `password_is_included` has been removed as it had no purpose
- `sonatypeiq_organization`: Added nested `categories`

The following resources have experienced field visibility changes as part of [GH-64]:
- `sonatypeiq_application`
- `sonatypeiq_application_category`
- `sonatypeiq_application_role_membership`
- `sonatyepiq_config_crowd`
- `sonatyepiq_config_mail`
- `sonatypeiq_config_product_license`
- `sonatypeiq_config_proxy_server`
- `sonatypeiq_config_saml`
- `sonatypeiq_organization`
- `sonatypeiq_organization_role_membership`
- `sonatypeiq_source_control`
- `sonatypeiq_system_config`
- `sonatypeiq_user`

## Improvements

The following resources now support import:
- `sonatypeiq_application_role_membership`
- `sonatypeiq_organization_role_membership`
- `sonatyepiq_config_crowd`
- `sonatyepiq_config_mail`
- `sonatypeiq_config_proxy_server`
- `sonatypeiq_system_config`
- `sonatypeiq_user`