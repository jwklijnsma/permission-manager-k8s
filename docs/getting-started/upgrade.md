# Upgrade

## Helm

```shell
helm upgrade permission-manager oci://ghcr.io/jwklijnsma/charts/permission-manager \
  --namespace permission-manager --reuse-values
```

The chart upgrades the Deployment, the ClusterRole of Permission Manager and the
seeded templates. Users (`PermissionManagerUser` resources), their
ServiceAccounts and their bindings are plain cluster objects and are not touched.

## From upstream sighupio/permission-manager 1.x

The data model did not change, so existing users, templates and bindings keep
working. Things to check:

- **Image.** The chart now defaults to `ghcr.io/jwklijnsma/permission-manager-k8s`
  and `image.tag` defaults to `v` + `appVersion`. Remove an `image.tag: v1.9.0`
  override from your values, otherwise you keep running the old image.
- **ClusterRole.** Permission Manager now also needs `create` on
  `serviceaccounts/token` and `get` on the `kube-root-ca.crt` ConfigMap. The
  chart and `deploy.yml` include it; adjust it if you maintain your own RBAC.
- **PORT.** The chart sets `PORT` in the Deployment, so an
  [existing Secret](../configuration/secrets.md) does not need it.
- **Ingress.** `ingress.enabled` is `false` by default (the previous fork commit
  had turned it on with a placeholder host). Only `networking.k8s.io/v1` is
  rendered.
- **Deleting a user** now also deletes its ServiceAccount and token Secret.
  Kubeconfigs of deleted users stop working, which is the intended behaviour.
- **Legacy users** (created before upstream 1.6, with `User` subjects only) are
  still detected and offered for migration. Bindings that contain the
  ServiceAccount *and* a `User` subject (see [binding subjects](../configuration/subjects.md))
  are not considered legacy.

## Back up

Everything Permission Manager manages lives in the Kubernetes API:

```shell
kubectl get permissionmanagerusers -o yaml > users.yaml
kubectl get clusterroles -o yaml | yq '.items |= map(select(.metadata.name | test("^template-")))' > templates.yaml
kubectl get rolebindings -A -l generated_for_user -o yaml > rolebindings.yaml
kubectl get clusterrolebindings -o yaml | yq '.items |= map(select(.metadata.name | test("___")))' > clusterrolebindings.yaml
```
