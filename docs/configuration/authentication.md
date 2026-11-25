# Authentication

Anyone who can use Permission Manager can grant themselves any permission in the
cluster, so access to it must be restricted.

## Basic auth (default)

The UI and the API are protected with HTTP basic auth. The username is `admin`
unless `BASIC_AUTH_USERNAME` / `config.basicAuthUsername` is set; the password
comes from `BASIC_AUTH_PASSWORD` / `config.basicAuthPassword`. Credentials are
compared in constant time.

The server refuses to start when basic auth is enabled and the password is empty.

## Disabling basic auth behind an authenticating proxy

When Permission Manager is only reachable through a proxy that already
authenticates users (oauth2-proxy, an ingress with external auth, a service
mesh...), basic auth can be turned off (upstream issue #190):

=== "Helm"

    ```yaml
    config:
      basicAuthEnabled: false
    ```

=== "Environment"

    ```shell
    BASIC_AUTH_DISABLED=true
    ```

The setting fails closed:

- only an explicit boolean (`true`, `1`, `TRUE`...) disables basic auth;
- an invalid value such as `yes please` stops the server at start-up;
- an empty or missing `BASIC_AUTH_DISABLED` keeps basic auth on.

The server logs a warning on start when basic auth is disabled.

!!! danger "Make sure nothing bypasses the proxy"
    With basic auth off, the Service itself is unauthenticated. Keep the Service
    `ClusterIP`, don't enable a second Ingress, and restrict in-cluster access
    with a NetworkPolicy that only allows the proxy.

### Example: ingress-nginx with oauth2-proxy

```yaml
config:
  basicAuthEnabled: false
ingress:
  enabled: true
  className: nginx
  annotations:
    nginx.ingress.kubernetes.io/auth-url: "https://oauth2.example.com/oauth2/auth"
    nginx.ingress.kubernetes.io/auth-signin: "https://oauth2.example.com/oauth2/start?rd=$escaped_request_uri"
  hosts:
    - host: permission-manager.example.com
      paths:
        - path: /
          pathType: Prefix
```

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: permission-manager-only-from-ingress
  namespace: permission-manager
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: permission-manager
  policyTypes: [Ingress]
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: ingress-nginx
```
