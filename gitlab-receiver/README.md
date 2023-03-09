# Gitlab Event Listener

The Gitlab Event Listener is a server which creates CloudEvents internally to Direktiv in response to received [Gitlab Webhook events](https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html). 

## Build

Create the following environmental variables in the `Makefile`:

- DOCKER_REPO: docker repo to use, defaults to localhost:5000
- RELEASE_TAG: version to relase, defaults to latest

then run the following command to build the container image to the repository

```
make docker
```

## Install

Modify kubernetes/install.yaml and add the following information:

```yaml
    server:
      bind: ":8080"
      tls: true
      certFile: "/certfilepath"
      keyFile: "/keyfilepath"

    gitlab:
      token: <user-level API key is your authentication token for the Equinix Metal API across the projects and organizations that you have access to>

    direktiv:
      endpoint: https://<direktiv-url>/api/namespaces/<namespace>/broadcast
      insecureSkipVerify: true
      token: <direktiv-namespace-token>
      event-on-error: true
```

Install the service on the Kubernetes platform using:

```sh
kubectl apply -f kubernetes/install.yaml
```