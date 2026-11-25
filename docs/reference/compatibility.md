# Kubernetes compatibility

| Permission Manager | Kubernetes |
|---|---|
| this fork, 1.10.x | **1.24 - 1.34** |
| upstream 1.9.x | 1.20 - 1.23 reliably; kubeconfig generation could fail on 1.24+ |

This answers upstream issues #68, #173 and #193.

Why 1.24 as minimum:

- the chart declares `kubeVersion: ">=1.24.0-0"` and renders only `networking.k8s.io/v1`
  Ingresses and `apiextensions.k8s.io/v1` CRDs;
- kubeconfig generation relies on the `kube-root-ca.crt` ConfigMap (1.20+) and,
  in `tokenrequest` mode, on the TokenRequest API.

Why it works up to 1.34:

- the server only uses stable APIs that are unchanged in 1.34: `v1` ServiceAccounts,
  Secrets, ConfigMaps, Namespaces, `rbac.authorization.k8s.io/v1` and the
  `permissionmanager.user/v1alpha1` CRD;
- the Go client (`client-go` v0.26) talks to newer API servers for these stable
  APIs; the [version skew policy](https://kubernetes.io/releases/version-skew-policy/)
  is not a hard limit for clients. Client library upgrades come through Renovate.

The e2e tests (`make test-e2e CLUSTER_VERSION=1.34.0 ...`) create a kind cluster
of the requested version; the kind configuration uses a version independent
kubeadm patch so it works with every node image.
