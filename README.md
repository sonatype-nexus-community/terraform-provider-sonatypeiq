# Terraform Provider for Sonatype IQ Server

This Terraform Provider allows you to use Configuration-as-Code (CasC) practises for 
managing the configuration deployments of Sonatype IQ Server which powers 
[Sonatype Repository Firewall](https://www.sonatype.com/products/sonatype-repository-firewall), 
[Sonatype Lifecycle](https://www.sonatype.com/products/open-source-security-dependency-management) and 
[Sonatype Auditor](https://www.sonatype.com/products/auditor).

This provider does not provide functionality for actually deploying Sonatype IQ Server 
(i.e. Infrastructure or Application installation). For deployment and installation, see 
the [official Help Documentation](https://help.sonatype.com/iqserver/installing).


## Usage

See our [documentation](./docs/index.md) and the [examples directory](./examples/).

## Development

This provider follows uses the Custom Provider Framework from HashiCorp. A great reference is available from HashiCorp [here](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider).

### Linting

`golangci-lint run`

### Acceptance Testing

Acceptance testing requires a valid licenses Sonatype IQ Server and administrative credentials.

Configure the IQ Server to use for Acceptance Testing using Environment Variables:

```bash
IQ_SERVER_URL=
IQ_SERVER_USERNAME=
IQ_SERVER_PASSWORD=
```

Then you can run the tests:

`TF_ACC=1 go test -v -cover ./internal/provider/`

## The Fine Print

Remember:

It is worth noting that this is **NOT SUPPORTED** by Sonatype, and is a contribution of ours to the open source community (read: you!)

* Use this contribution at the risk tolerance that you have
* Do NOT file Sonatype support tickets related to `terraform-provider-sonatypeiq-pf`
* DO file issues here on GitHub, so that the community can pitch in

Phew, that was easier than I thought. Last but not least of all - have fun!