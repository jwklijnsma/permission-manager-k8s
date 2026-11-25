# Quick start

After [installing](install.md), open the UI and log in as `admin`.

1. **Create a user.** Click *Create new user* and choose a name. Names are used
   as ServiceAccount names, so they must be DNS-1123 subdomains: lower case
   letters, digits, `-` and `.`, starting and ending with a letter or digit,
   at most 253 characters (for example `jane.doe`).
2. **Pick templates.** For each template (for example `developer`) choose the
   namespaces it applies to, or *all namespaces*.
3. **Pick the cluster access.** *None*, *read* (`template-cluster-resources___read-only`)
   or *write* (`template-cluster-resources___admin`) on cluster-scoped resources.
4. **Save** and check the summary.
5. **Download the kubeconfig** from the user page, choosing the default namespace
   of the context.

Try it:

```shell
export KUBECONFIG=./jane.doe.yaml
kubectl auth whoami
kubectl get pods -n dev          # allowed with the developer template in dev
kubectl get pods -n kube-system  # forbidden
```

`kubectl auth whoami` shows
`system:serviceaccount:permission-manager:jane.doe`: every user authenticates as
its own ServiceAccount.

Next: [how tokens are created](../configuration/kubeconfig-tokens.md) and how to
[automate users](../usage/api.md).
