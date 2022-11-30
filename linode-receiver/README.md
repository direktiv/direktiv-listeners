# Linode Event Listener

The Linode Event Listener is a poller which creates CloudEvents internally to Direktiv using the [Linode APIs](https://www.linode.com/docs/api/). 

It uses the Golang Linode SDK [linodego](https://github.com/linode/linodego) and calls the https://www.linode.com/docs/api/account/#events-list endpoint. 

The endpoint returns a collection of Event objects representing actions taken on your Account from the last 90 days. The Events returned depend on your grants.

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
    linode:
      linodeAuthToken: <user-level API key is your authentication token for the Equinix Metal API across the projects and organizations that you have access to>

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