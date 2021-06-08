# GSLB Controller

**This project is under active development and not usable yet**

A cloud native global server load balancer for providing multi-datacenter highly available ingress for k8s.


## Instructions

### Development

* `make generate` update the generated code for that resource type.
* `make manifests` Generating CRD manifests.
* `make test` Run tests.


### Build

* `make build` builds golang app locally.
* `make docker-build` build docker image locally.
* `make docker-push` push container image to registry.

### Run, Deploy
* `make run` run app locally
* `make deploy` deploy to k8s.

### Clean up

* `make undeploy` delete resouces in k8s.


## Security

### Reporting security vulnerabilities

If you find a security vulnerability or any security related issues, please DO NOT file a public issue, instead send your report privately to cloud@snapp.cab. Security reports are greatly appreciated and we will publicly thank you for it.

## License

Apache-2.0 License, see [LICENSE](LICENSE).
