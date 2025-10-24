# Terraform Provider for Sonatype IQ Server

[![shield_tfr-version]][link_tfr]
![Terraform Provider Downloads](https://img.shields.io/terraform/provider/dm/4429)
[![shield_gh-workflow-test]][link_gh-workflow-test]
[![shield_license]][license_file]
![shield_tf_version]
[![shield_nxiq_version]][link_nxiq_release]

---

This Terraform Provider allows you to use Configuration-as-Code (CasC) practises for 
managing the configuration of Sonatype IQ Server which powers: 
- [Sonatype Repository Firewall](https://www.sonatype.com/products/sonatype-repository-firewall)
- [Sonatype Lifecycle](https://www.sonatype.com/products/open-source-security-dependency-management) 
- [Sonatype SBOM Manager](https://www.sonatype.com/products/sonatype-sbom-manager)
- [Sonatype Advanced Legal Pack](https://www.sonatype.com/products/advanced-legal-pack)

This provider does not provide functionality for actually deploying Sonatype IQ Server (i.e. Infrastructure or Application installation). For deployment and installation, see  the [official Help Documentation](https://help.sonatype.com/iqserver/installing).

## Version Support

We test this Provider against a range of Terraform versions and Sonatype IQ Server versions as noted below.

### Sonatype Nexus Repository Manager

We test on `N - 5` releases (where possible, but no earlier than `186`). See [here](https://github.com/sonatype-nexus-community/terraform-provider-sonatypeiq/blob/main/.github/workflows/test.yml) for the current list.

### Terraform Version support

We test on the latest patch release of each the earliest and latest version of Terraform i.e. `1.0.x` and  `1.12.x` - i.e. we aim to support all Terraform versions since `1.0.0`.

## Usage

See our [documentation](./docs/index.md) and the [examples directory](./examples/).

## Development

See [CONTRIBUTING.md](./CONTRIBUTING.md) for details.

## The Fine Print

Remember:

It is worth noting that this is **NOT SUPPORTED** by Sonatype, and is a contribution of ours to the open source community (read: you!)

* Use this contribution at the risk tolerance that you have
* Do NOT file Sonatype support tickets related to `terraform-provider-sonatypeiq`
* DO file issues here on GitHub, so that the community can pitch in

Phew, that was easier than I thought. Last but not least of all - have fun!


[shield_gh-workflow-test]: https://img.shields.io/github/actions/workflow/status/sonatype-nexus-community/terraform-provider-sonatypeiq/test.yml?branch=main&logo=GitHub&logoColor=white "build"
[shield_tfr-version]: https://img.shields.io/badge/Terraform%20Registry-0.11.0-8A2BE2
[shield_license]: https://img.shields.io/github/license/sonatype-nexus-community/terraform-provider-sonatypeiq?logo=open%20source%20initiative&logoColor=white "license"
[shield_tf_version]: https://img.shields.io/badge/Terraform-1.0.0+-blue
[shield_nxiq_version]: https://img.shields.io/badge/Sonatype_IQ-186&nbsp;&ndash;&nbsp;196-blue

[link_tfr]: https://registry.terraform.io/providers/sonatype-nexus-community/sonatypeiq/latest
[link_gh-workflow-test]: https://github.com/sonatype-nexus-community/terraform-provider-sonatypeiq/actions/workflows/test.yml?query=branch%3Amain
[license_file]: https://github.com/sonatype-nexus-community/terraform-provider-sonatypeiq/blob/main/LICENSE
[link_nxiq_release]: https://help.sonatype.com/en/iq-server-release-notes.html