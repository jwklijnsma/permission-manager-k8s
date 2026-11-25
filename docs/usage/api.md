# HTTP API and automation

Everything the UI does goes through a small JSON API under `/api`, protected by
the same [authentication](../configuration/authentication.md) as the UI. It can
be used from scripts and pipelines to create users without the UI (upstream
issue #191).

All requests are `POST` with `Content-Type: application/json` unless noted.
Validation errors return HTTP 400 with `{"error": "..."}`.

| Endpoint | Body | Effect |
|---|---|---|
| `GET /api/list-users` | - | list users |
| `GET /api/list-namespace` | - | list namespaces |
| `GET /api/rbac` | - | ClusterRoles, ClusterRoleBindings, Roles, RoleBindings |
| `/api/create-user` | `{"name": "jane.doe"}` | create the `PermissionManagerUser` |
| `/api/delete-user` | `{"username": "jane.doe"}` | delete the user, its ServiceAccount and token Secret |
| `/api/create-rolebinding` | see below | create a RoleBinding |
| `/api/delete-rolebinding` | `{"rolebindingName": "...", "namespace": "..."}` | |
| `/api/create-cluster-rolebinding` | see below | create a ClusterRoleBinding |
| `/api/delete-cluster-rolebinding` | `{"rolebindingName": "..."}` | |
| `/api/create-cluster-role` | `{"roleName": "...", "rules": [...]}` | create a ClusterRole (templates) |
| `/api/delete-cluster-role` | `{"roleName": "..."}` | |
| `/api/create-kubeconfig` | `{"username": "jane.doe", "namespace": "dev"}` | returns `{"ok": true, "kubeconfig": "..."}` |
| `/api/check-legacy-user` | `{"username": "...", "namespaces": [...]}` | detects bindings from upstream < 1.6 |

## Subjects

`subjects` in the binding endpoints accept the kinds `ServiceAccount`, `User`
and `Group`; anything else is rejected. `ServiceAccount` subjects are always
pinned to the permission-manager namespace. See
[binding subjects](../configuration/subjects.md).

## Example: create a user with the developer template in `dev`

```shell
PM=https://permission-manager.example.com
AUTH="admin:changeMe"
USER=jane.doe

curl -fsS -u "$AUTH" -H 'Content-Type: application/json' \
  "$PM/api/create-user" -d "{\"name\":\"$USER\"}"

curl -fsS -u "$AUTH" -H 'Content-Type: application/json' "$PM/api/create-rolebinding" -d @- <<JSON
{
  "rolebindingName": "${USER}___template-namespaced-resources___developer___dev",
  "namespace": "dev",
  "generated_for_user": "${USER}",
  "roleKind": "ClusterRole",
  "roleName": "template-namespaced-resources___developer",
  "subjects": [{"kind": "ServiceAccount", "name": "${USER}"}]
}
JSON

curl -fsS -u "$AUTH" -H 'Content-Type: application/json' \
  "$PM/api/create-kubeconfig" -d "{\"username\":\"$USER\",\"namespace\":\"dev\"}" \
  | jq -r .kubeconfig > "$USER.kubeconfig"
```

Use the naming scheme `<user>___<template>___<namespace>` for RoleBindings and
`<user>___<template>` for ClusterRoleBindings so the UI shows the bindings on the
user's page.

## Users as code (GitOps)

Users, templates and bindings are plain Kubernetes objects, so they can also be
kept in Git and applied by Argo CD or Flux (upstream issue #116), with
Permission Manager used to look at them and to generate kubeconfigs:

```yaml
apiVersion: permissionmanager.user/v1alpha1
kind: Permissionmanageruser
metadata:
  name: permissionmanagerusers.permissionmanager.user.jane.doe
spec:
  name: jane.doe
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: jane.doe___template-namespaced-resources___developer___dev
  namespace: dev
  labels:
    generated_for_user: jane.doe
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: template-namespaced-resources___developer
subjects:
  - kind: ServiceAccount
    name: jane.doe
    namespace: permission-manager
```
