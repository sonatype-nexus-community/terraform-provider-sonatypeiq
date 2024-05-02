# Terraform Provider for Sonatype IQ Server

[![shield_tfr-version]][link_tfr]
[![shield_gh-workflow-test]][link_gh-workflow-test]
[![shield_license]][license_file]

---

This Terraform Provider allows you to use Configuration-as-Code (CasC) practises for 
managing the configuration of Sonatype IQ Server which powers: 
- [Sonatype Repository Firewall](https://www.sonatype.com/products/sonatype-repository-firewall)
- [Sonatype Lifecycle](https://www.sonatype.com/products/open-source-security-dependency-management) 
- [Sonatype Auditor](https://www.sonatype.com/products/auditor)

This provider does not provide functionality for actually deploying Sonatype IQ Server (i.e. Infrastructure or Application installation). For deployment and installation, see  the [official Help Documentation](https://help.sonatype.com/iqserver/installing).

## Usage

See our [documentation](./docs/index.md) and the [examples directory](./examples/).

## Development

See [CONTRIBUTING.md](./CONTRIBUTING.md) for details.

## The Fine Print

Remember:

It is worth noting that this is **NOT SUPPORTED** by Sonatype, and is a contribution of ours to the open source community (read: you!)

* Use this contribution at the risk tolerance that you have
* Do NOT file Sonatype support tickets related to `terraform-provider-sonatypeiq-pf`
* DO file issues here on GitHub, so that the community can pitch in

Phew, that was easier than I thought. Last but not least of all - have fun!


[shield_gh-workflow-test]: https://img.shields.io/github/actions/workflow/status/sonatype-nexus-community/terraform-provider-sonatypeiq/test.yml?branch=main&logo=GitHub&logoColor=white "build"
[shield_tfr-version]: https://img.shields.io/badge/Terraform%20Registry-8A2BE2
[shield_license]: https://img.shields.io/github/license/sonatype-nexus-community/terraform-provider-sonatypeiq?logo=open%20source%20initiative&logoColor=white "license"

[link_tfr]: https://registry.terraform.io/providers/sonatype-nexus-community/sonatypeiq/latest
[link_gh-workflow-test]: https://github.com/sonatype-nexus-community/terraform-provider-sonatypeiq/actions/workflows/test.yml?query=branch%3Amain
[license_file]: https://github.com/sonatype-nexus-community/terraform-provider-sonatypeiq/blob/main/LICENSE