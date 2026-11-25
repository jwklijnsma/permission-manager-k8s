# Users and kubeconfigs

## What a user is

| Object | Name | Created when |
|---|---|---|
| `Permissionmanageruser` (cluster scoped, group `permissionmanager.user`) | `permissionmanagerusers.permissionmanager.user.<username>` | the user is created |
| ServiceAccount in the permission-manager namespace | `<username>` | the first kubeconfig is generated |
| Token Secret (`kubernetes.io/service-account-token`) | `<username>` | first kubeconfig, `secret` token mode only |
| RoleBindings | `<username>___<template>___<namespace>` | templates are assigned to namespaces |
| ClusterRoleBindings | `<username>___<template>all_namespaces`, `<username>___template-cluster-resources___<access>` | *all namespaces* templates, cluster access |

Bindings carry the label `generated_for_user=<username>`:

```shell
kubectl get rolebindings,clusterrolebindings -A -l generated_for_user=jane.doe
```

## Usernames

Usernames become ServiceAccount names, so they must be valid
[DNS-1123 subdomains](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-subdomain-names):

- lower case letters, digits, `-` and `.`
- start and end with a letter or digit
- at most 253 characters

`jane.doe` and `team-a.ci` are valid; `Jane`, `jane_doe` and `jane@example.com`
are not. E-mail addresses (upstream issue #42) can't be ServiceAccount names; to
grant permissions to an OIDC identity such as an e-mail address, use
[additional `User` subjects](../configuration/subjects.md) with a matching
username claim, or bind it directly with your own RoleBinding.

## Kubeconfigs

*Show kubeconfig* on the user page calls `POST /api/create-kubeconfig`. Permission
Manager:

1. creates the user's ServiceAccount if it does not exist;
2. gets a token for it, depending on the [token mode](../configuration/kubeconfig-tokens.md);
3. reads the cluster CA from the token Secret or from the `kube-root-ca.crt` ConfigMap;
4. returns a kubeconfig with one cluster (`CLUSTER_NAME`, `CONTROL_PLANE_ADDRESS`),
   one user (the token) and one context `<username>@<cluster>` with the chosen
   default namespace.

If no token can be obtained (for example the token controller never fills the
Secret), the request fails with HTTP 500 and an explanation instead of returning
a kubeconfig with an empty token.

## Deleting a user

Deleting a user from the UI deletes its bindings, its `PermissionManagerUser`,
its ServiceAccount and its token Secret. Every token issued for the
ServiceAccount (both modes) becomes invalid immediately, so kubeconfigs that were
handed out stop working.
