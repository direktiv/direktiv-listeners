# SMTP Listener

## Build

- DOCKER_REPO: docker repo to use, defaults to localhost:5000
- RELEASE_TAG: version to relase, defaults to latest

```
make docker
```

## Install

Modify kubernetes/install.yaml if required (e.g. apikey, target namespace)

```
make install
```

## Modify Direktiv

Helm upgrade direktiv with something like the following. 

```yaml
ingress-nginx:
  tcp:
    2525: default/smtp-listener-service:2525
```

## Test 

A little test app is in test directory. It attaches everything in test/attachments. 

```
go run test.go MY_KUBERNETS_IP:2525 email1@email email2@email
```

## Version

direktiv/smtp-listener:v0.2