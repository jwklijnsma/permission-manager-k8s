# Binding subjects

By default every RoleBinding and ClusterRoleBinding generated for a user has one
subject: the user's ServiceAccount in the permission-manager namespace.

If your users also log in with OIDC or client certificates under the same name,
the bindings can include a `User` and/or `Group` subject with the username as
well (upstream issues #117, #41, #91 and PR #194):

=== "Helm"

    ```yaml
    config:
      additionalSubjectKinds:
        - User
    ```

=== "Environment"

    ```shell
    ADDITIONAL_SUBJECT_KINDS=User        # or User,Group
    ```

A binding generated for `jane.doe` then looks like:

```yaml
subjects:
  - kind: ServiceAccount
    name: jane.doe
    namespace: permission-manager
  - apiGroup: rbac.authorization.k8s.io
    kind: User
    name: jane.doe
```

The extra subjects are added by the server, for every ServiceAccount subject of a
binding, so the UI and API clients need no change. Only bindings created after the
setting is enabled get them; to update existing users, save them again in the UI.

!!! warning
    A `User` or `Group` subject grants the permissions to *any* identity the API
    server authenticates with that name. Only enable it when usernames in
    Permission Manager match the identities of your authentication provider.

## Validation of subjects

The API accepts only `ServiceAccount`, `User` and `Group` subjects:

- `ServiceAccount` subjects are pinned to the permission-manager namespace, so a
  request can't bind ServiceAccounts of other namespaces;
- `User` and `Group` subjects get `apiGroup: rbac.authorization.k8s.io` and no namespace;
- duplicate subjects are removed;
- a binding needs at least one subject and `roleKind` must be `ClusterRole` or `Role`.
