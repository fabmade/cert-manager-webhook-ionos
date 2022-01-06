# Ionos webhook for cert manager

Cert-manager ACME DNS webhook provider for ionos.
see: https://cert-manager.io/docs/configuration/acme/dns01/webhook/

## Install

### Install cert manager

see: https://cert-manager.io/docs/installation/kubernetes/

### Install webhook 

`helm install cert-manager-webhook-ionos ./deploy/cert-manager-webhook-ionos`


### Running the test suite

All DNS providers **must** run the DNS01 provider conformance testing suite,
else they will have undetermined behaviour when used with cert-manager.

**It is essential that you configure and run the test suite when creating a
DNS01 webhook.**

An example Go test file has been provided in [main_test.go](https://github.com/jetstack/cert-manager-webhook-example/blob/master/main_test.go).

You can run the test suite with:

```bash
$ TEST_ZONE_NAME=example.com. make test
```

The example file has a number of areas you must fill in and replace with your
own options in order for tests to pass.
