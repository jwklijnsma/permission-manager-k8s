# Kubeconfig tokens

Every user authenticates with a token of its own ServiceAccount. Since
Kubernetes 1.24 the token controller no longer creates a token Secret for every
ServiceAccount automatically, which broke kubeconfig generation in some setups
(upstream issue #168). Permission Manager supports two ways to get a token.

| | `secret` (default) | `tokenrequest` |
|---|---|---|
| How | creates a `kubernetes.io/service-account-token` Secret named after the user and waits for the token controller to fill it | calls the [TokenRequest API](https://kubernetes.io/docs/reference/kubernetes-api/authentication-resources/token-request-v1/) |
| Expiry | never (until the Secret or the ServiceAccount is deleted) | `tokenExpiration`, default `8760h` (1 year), minimum `10m` |
| Stored in the cluster | yes, in the Secret | no |
| Same token on every download | yes | no, a new token each time |
| Revocation | delete the user (or its Secret) | delete the user; tokens expire on their own |

=== "Helm"

    ```yaml
    config:
      kubeconfig:
        tokenMode: tokenrequest
        tokenExpiration: 720h
    ```

=== "Environment"

    ```shell
    KUBECONFIG_TOKEN_MODE=tokenrequest
    KUBECONFIG_TOKEN_EXPIRATION=720h
    ```

!!! note "Maximum expiration"
    The API server can cap token lifetimes with
    `--service-account-max-token-expiration`; a longer expiration is silently
    shortened. Managed clusters often cap it.

## Secret mode details

- The Secret is created with the `kubernetes.io/service-account.name` and
  `kubernetes.io/service-account.uid` annotations. Permission Manager waits up
  to 30 seconds for the token to appear and returns an error if it doesn't.
- If a Secret with the user's name exists but belongs to an older
  ServiceAccount (different UID), it is deleted and recreated so a stale token
  is never handed out.
- A Secret with the user's name that is not a service account token is never
  overwritten: kubeconfig generation fails with an error instead.
- Kubernetes 1.29+ can clean up *auto-generated* legacy token Secrets that are
  unused for a long time (`LegacyServiceAccountTokenCleanUp`). Secrets created
  by Permission Manager are created explicitly and are not part of that cleanup.

## CA certificate

The kubeconfig embeds the cluster CA. It comes from the token Secret (`ca.crt`)
or, in `tokenrequest` mode or when the Secret has none, from the
`kube-root-ca.crt` ConfigMap that Kubernetes publishes in every namespace.

If your users reach the API server through a load balancer with a different
certificate, set `CONTROL_PLANE_ADDRESS` to an address served with the cluster
CA, or distribute the right CA separately.

## Required permissions

The ClusterRole of Permission Manager (chart and `deploy.yml`) contains:

```yaml
- apiGroups: [""]
  resources: ["serviceaccounts", "secrets"]
  verbs: ["get", "list", "create", "update", "delete", "watch"]
- apiGroups: [""]
  resources: ["serviceaccounts/token"]
  verbs: ["create"]
- apiGroups: [""]
  resources: ["configmaps"]
  resourceNames: ["kube-root-ca.crt"]
  verbs: ["get"]
```
