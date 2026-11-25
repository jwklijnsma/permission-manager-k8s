# Templates

A template is a ClusterRole whose name starts with
`template-namespaced-resources___`. The part after the prefix is the name shown
in the UI.

The chart creates one ClusterRole per entry in `config.templates`:

```yaml
config:
  templates:
    - name: developer
      rules:
        - apiGroups: ["*"]
          resources: ["pods", "pods/log", "deployments", "services"]
          verbs: ["*"]
    - name: viewer
      rules:
        - apiGroups: ["", "apps"]
          resources: ["pods", "deployments", "services"]
          verbs: ["get", "list", "watch"]
```

Templates can also be created with plain `kubectl apply`; the UI lists every
ClusterRole with the prefix:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: template-namespaced-resources___viewer
rules:
  - apiGroups: [""]
    resources: ["pods", "services", "configmaps"]
    verbs: ["get", "list", "watch"]
```

## Read and write in the UI

The template summary shows two columns per resource. They are derived from the
Kubernetes verbs of each rule:

| Column | Checked when the rule has |
|---|---|
| **read** | `*`, `get`, `list` or `watch` |
| **write** | `*`, `create`, `update`, `patch`, `delete` or `deletecollection` |

Kubernetes has no `read` verb (upstream issue #186): use `get`, `list` and
`watch`. Earlier versions only looked for `*`, `read` and `create`, so templates
with explicit verbs looked as if they granted nothing (upstream issue #84).
Rules without `resources` (for example `nonResourceURLs`) are skipped in the
summary instead of breaking it.

## Cluster access

The *cluster access* radio buttons bind one of two ClusterRoles that are seeded
by the chart:

| Choice | ClusterRole |
|---|---|
| none | - |
| read | `template-cluster-resources___read-only` |
| write | `template-cluster-resources___admin` |
