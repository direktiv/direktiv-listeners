# Equinix Event Listener

The Equinix Event Listener is a poller which creates CloudEvents internally to Direktiv using the Equinix APIs. 

It uses the Golang Equinix SDK [packngo](https://github.com/packethost/packngo) and calls the https://api.equinix.com/metal/v1/organizations/{id}/events endpoint. 

The endpoint retrieves , Project & Device level events (see https://deploy.equinix.com/developers/api/metal/#tag/Events/operation/findOrganizationEvents)

## Install

Modify kubernetes/install.yaml and add the following information:

```yaml
    equinix:
      organizationId: <packet organization id from Organizations page>
      packetAuthToken: <user-level API key is your authentication token for the Equinix Metal API across the projects and organizations that you have access to>

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