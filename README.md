# Ionos webhook for cert manager

Cert-manager ACME DNS webhook provider for ionos.
see: https://cert-manager.io/docs/configuration/acme/dns01/webhook/

## Install

### Install cert manager

see: https://cert-manager.io/docs/installation/kubernetes/

### Install webhook 

Add helm repo

`helm repo add cert-manager-webhook-ionos https://fabmade.github.io/cert-manager-webhook-ionos`

install helm chart

`helm install cert-manager-webhook-ionos cert-manager-webhook-ionos/cert-manager-webhook-ionos -ncert-manager`

add secret

```
apiVersion: v1
stringData:
  IONOS_PUBLIC_PREFIX: <your-public-key>
  IONOS_SECRET: <your-private-key>
kind: Secret
metadata:
  name: ionos-secret
  namespace: cert-manager
type: Opaque
```

add staging issuer

```
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-ionos-staging
spec:
  acme:
    # The ACME server URL
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    # Email address used for ACME registration
    email: <your-email>
    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-ionos-staging-key
    # Enable the dns01 challenge provider
    solvers:
      - dns01:
          webhook:
            groupName: acme.fabmade.de
            solverName: ionos
            config:
              apiUrl: https://api.hosting.ionos.com/dns/v1
              publicKeySecretRef:
                key: IONOS_PUBLIC_PREFIX
                name: ionos-secret
              secretKeySecretRef:
                key: IONOS_SECRET
                name: ionos-secret
```
add prod issuer

```
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-ionos-prod
spec:
  acme:
    # The ACME server URL
    server: https://acme-v02.api.letsencrypt.org/directory
    # Email address used for ACME registration
    email: <your-email-address>
    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-ionos-prod
    # Enable the dns01 challenge provider
    solvers:
      - dns01:
          webhook:
            groupName: acme.fabmade.de
            solverName: ionos
            config:
              apiUrl: https://api.hosting.ionos.com/dns/v1
              publicKeySecretRef:
                key: IONOS_PUBLIC_PREFIX
                name: ionos-secret
              secretKeySecretRef:
                key: IONOS_SECRET
                name: ionos-secret
```

add ingress or certificate for example.com domain (replace it with your domain)

```
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-test-com
spec:
  dnsNames:
    - '*.example.com'
  issuerRef:
    name: letsencrypt-ionos-staging
  secretName: example-test-com-tls
```

replace service "mybackend" with your own service

```
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cert-manager.io/issuer: letsencrypt-ionos-staging
  name: example-wildcard-ingress
spec:
  rules:
    - host: '*.example.com'
      http:
        paths:
          - backend:
              service:
                name: mybackend
                port:
                  number: 80
            path: /
            pathType: Prefix
  tls:
    - hosts:
        - '*.example.com'
      secretName: example-ionos-tls-prod
```

share secrets accross namespaces (optional)

https://cert-manager.io/docs/faq/kubed/

### Uninstall webhook

```helm uninstall cert-manager-webhook-ionos -ncert-manager```

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
